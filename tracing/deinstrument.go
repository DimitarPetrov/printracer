package tracing

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"strings"
)

type codeDeinstrumenter struct {
}

func NewCodeDeinstrumenter() CodeDeinstrumenter {
	return &codeDeinstrumenter{}
}

func (cd *codeDeinstrumenter) DeinstrumentDirectory(path string) error {
	fset := token.NewFileSet()
	filter := func(info os.FileInfo) bool {
		return testsFilter(info) && generatedFilter(path, info)
	}
	pkgs, err := parser.ParseDir(fset, path, filter, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed parsing go files in directory %s: %v", path, err)
	}

	for _, pkg := range pkgs {
		if err := cd.DeinstrumentPackage(fset, pkg); err != nil {
			return err
		}
	}
	return nil
}

func (cd *codeDeinstrumenter) DeinstrumentPackage(fset *token.FileSet, pkg *ast.Package) error {
	for fileName, file := range pkg.Files {
		sourceFile, err := os.OpenFile(fileName, os.O_TRUNC|os.O_WRONLY, 0664)
		if err != nil {
			return fmt.Errorf("failed opening file %s: %v", fileName, err)
		}
		if err := cd.DeinstrumentFile(fset, file, sourceFile); err != nil {
			return fmt.Errorf("failed deinstrumenting file %s: %v", fileName, err)
		}
	}
	return nil
}

func (cd *codeDeinstrumenter) DeinstrumentFile(fset *token.FileSet, file *ast.File, out io.Writer) error {
	// Needed because ast does not support floating comments and deletes them.
	// In order to preserve all comments we just pre-parse it to dst which treats them as first class citizens.
	f, err := decorator.DecorateFile(fset, file)
	if err != nil {
		return fmt.Errorf("failed converting file from ast to dst: %v", err)
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

	return decorator.Fprint(out, f)
}
