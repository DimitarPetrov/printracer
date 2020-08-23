package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const apply = "apply"
const revert = "revert"

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stdout, "Usage: printracer [%s/%s]\n", apply, revert)
		flag.PrintDefaults()
	}
}

func main() {
	operation := parseOperation()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed getting current working directory: %v", err))
	}
	err = filepath.Walk(wd,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == "vendor" {
				return filepath.SkipDir
			}

			if info.IsDir() {
				switch operation {
				case apply:
					return processDirectory(path)
				case revert:
					err := revertDirectory(path)
					if err != nil {
						return err
					}
					return removeUnusedImports(path)
				}
			}
			return nil
		})

	if err != nil {
		log.Fatal(fmt.Sprintf("Failed traversing directories: %v", err))
	}
}

func removeUnusedImports(path string) error {
	fset := token.NewFileSet()
	filter := func(info os.FileInfo) bool {
		return testsFilter(info) && generatedFilter(path, info)
	}
	pkgs, err := parser.ParseDir(fset, path, filter, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed parsing go files in directory %s: %v", path, err)
	}

	for _, pkg := range pkgs {
		for fileName, file := range pkg.Files {
			if !astutil.UsesImport(file, "fmt") {
				astutil.DeleteImport(fset, file, "fmt")
			}

			// Needed because ast does not support floating comments and deletes them.
			// In order to preserve all comments we just pre-parse it to dst which treats them as first class citizens.
			f, err := decorator.DecorateFile(fset, file)
			if err != nil {
				return fmt.Errorf("failed converting file %s from ast to dst: %v", fileName, err)
			}

			sourceFile, err := os.OpenFile(fileName, os.O_TRUNC|os.O_WRONLY, 0664)
			if err != nil {
				return fmt.Errorf("failed opening file %s: %v", fileName, err)
			}
			err = decorator.Fprint(sourceFile, f)
			if err != nil {
				return fmt.Errorf("failed writing file %s: %v", fileName, err)
			}
		}
	}
	return nil
}

func revertDirectory(path string) error {
	fset := token.NewFileSet()
	filter := func(info os.FileInfo) bool {
		return testsFilter(info) && generatedFilter(path, info)
	}
	pkgs, err := parser.ParseDir(fset, path, filter, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed parsing go files in directory %s: %v", path, err)
	}

	for _, pkg := range pkgs {
		for fileName, file := range pkg.Files {
			// Needed because ast does not support floating comments and deletes them.
			// In order to preserve all comments we just pre-parse it to dst which treats them as first class citizens.
			f, err := decorator.DecorateFile(fset, file)
			if err != nil {
				return fmt.Errorf("failed converting file %s from ast to dst: %v", fileName, err)
			}
			dst.Inspect(f, func(n dst.Node) bool {
				switch t := n.(type) {
				case *dst.FuncDecl:
					if len(t.Body.List) > 1 {
						stmt1, ok1 := t.Body.List[0].(*dst.ExprStmt)
						stmt2, ok2 := t.Body.List[1].(*dst.DeferStmt)
						if ok1 && ok2 {
							expr1, ok := stmt1.X.(*dst.CallExpr)
							if ok {
								selExpr1, ok1 := expr1.Fun.(*dst.SelectorExpr)
								selExpr2, ok2 := stmt2.Call.Fun.(*dst.SelectorExpr)
								if ok1 && ok2 {
									package1, ok1 := selExpr1.X.(*dst.Ident)
									package2, ok2 := selExpr2.X.(*dst.Ident)
									if ok1 && ok2 && package1.Name == "fmt" && package2.Name == "fmt" &&
										selExpr1.Sel.Name == "Printf" && selExpr2.Sel.Name == "Printf" {

										expr1Arg, ok1 := expr1.Args[0].(*dst.BasicLit)
										expr2Arg, ok2 := stmt2.Call.Args[0].(*dst.BasicLit)
										if ok1 && ok2 && expr1Arg.Kind == token.STRING && expr2Arg.Kind == token.STRING &&
											strings.Contains(expr1Arg.Value, fmt.Sprintf("Entering function %s", t.Name)) &&
											strings.Contains(expr2Arg.Value, fmt.Sprintf("Exiting function %s", t.Name)) {
											t.Body.List = t.Body.List[2:]
										}
									}
								}
							}
						}
					}
				}
				return true
			})

			sourceFile, err := os.OpenFile(fileName, os.O_TRUNC|os.O_WRONLY, 0664)
			if err != nil {
				return fmt.Errorf("failed opening file %s: %v", fileName, err)
			}
			err = decorator.Fprint(sourceFile, f)
			if err != nil {
				return fmt.Errorf("failed writing file %s: %v", fileName, err)
			}
		}
	}
	return nil
}

func processDirectory(path string) error {
	fset := token.NewFileSet()
	filter := func(info os.FileInfo) bool {
		return testsFilter(info) && generatedFilter(path, info)
	}
	pkgs, err := parser.ParseDir(fset, path, filter, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed parsing go files in directory %s: %v", path, err)
	}

	for _, pkg := range pkgs {
		for fileName, file := range pkg.Files {
			astutil.AddImport(fset, file, "fmt")

			// Needed because ast does not support floating comments and deletes them.
			// In order to preserve all comments we just pre-parse it to dst which treats them as first class citizens.
			f, err := decorator.DecorateFile(fset, file)
			if err != nil {
				return fmt.Errorf("failed converting file %s from ast to dst: %v", fileName, err)
			}

			dst.Inspect(f, func(n dst.Node) bool {
				switch t := n.(type) {
				case *dst.FuncDecl:
					var enteringStringFormat = fmt.Sprintf("Entering function %s", t.Name)
					var exitingStringFormat = fmt.Sprintf("Exiting function %s", t.Name)

					var args []dst.Expr

					if len(t.Type.Params.List) > 0 {
						enteringStringFormat += " with args"

						for _, param := range t.Type.Params.List {
							enteringStringFormat += " %v"
							args = append(args, &dst.BasicLit{
								Kind:  token.STRING,
								Value: param.Names[0].Name,
							})
						}
					}

					args = append([]dst.Expr{
						&dst.BasicLit{
							Kind:  token.STRING,
							Value: `"` + enteringStringFormat + `\n"`,
						},
					}, args...)

					t.Body.List = append([]dst.Stmt{
						&dst.ExprStmt{
							X: newPrintExprWithArgs(args),
						},
						&dst.DeferStmt{
							Call: newPrintExprWithMessage(exitingStringFormat),
						},
					}, t.Body.List...)
				}
				return true
			})

			sourceFile, err := os.OpenFile(fileName, os.O_TRUNC|os.O_WRONLY, 0664)
			if err != nil {
				return fmt.Errorf("failed opening file %s: %v", fileName, err)
			}
			err = decorator.Fprint(sourceFile, f)
			if err != nil {
				return fmt.Errorf("failed writing file %s: %v", fileName, err)
			}
		}
	}
	return nil
}

func generatedFilter(path string, info os.FileInfo) bool {
	if strings.Contains(info.Name(), "generated") {
		return false
	}

	f, err := os.Open(path + "/" + info.Name())
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed opening file %s: %v", path+"/"+info.Name(), err))
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	line := scanner.Text()

	if strings.Contains(line, "generated") || strings.Contains(line, "GENERATED") {
		return false
	}
	return true
}

func testsFilter(info os.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}

func newPrintExprWithMessage(msg string) *dst.CallExpr {
	return newPrintExprWithArgs([]dst.Expr{
		&dst.BasicLit{
			Kind:  token.STRING,
			Value: `"` + msg + `\n"`,
		},
	})
}

func newPrintExprWithArgs(args []dst.Expr) *dst.CallExpr {
	return &dst.CallExpr{
		Fun: &dst.SelectorExpr{
			X:   &dst.Ident{Name: "fmt"},
			Sel: &dst.Ident{Name: "Printf"},
		},
		Args: args,
	}
}

func parseOperation() string {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Operation must be specified [%s/%s]. Use printracer --help for more information.\n", apply, revert)
		os.Exit(1)
	}
	operation := os.Args[1]
	if operation != apply && operation != revert {
		helpFlags := map[string]bool{
			"--help": true,
			"-help":  true,
			"--h":    true,
			"-h":     true,
		}
		if helpFlags[operation] {
			flag.Parse()
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Unsupported operation: %s. Only [%s/%s] operations are supported.\n Use stegify --help for more information.", operation, apply, revert)
		os.Exit(1)
	}

	os.Args = append(os.Args[:1], os.Args[2:]...) // needed because go flags implementation stop parsing after first non-flag argument
	return operation
}
