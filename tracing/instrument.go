package tracing

import (
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"io"
	"os"
)

func InstrumentDirectory(path string) error {
	fset := token.NewFileSet()
	filter := func(info os.FileInfo) bool {
		return testsFilter(info) && generatedFilter(path, info)
	}
	pkgs, err := parser.ParseDir(fset, path, filter, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed parsing go files in directory %s: %v", path, err)
	}

	for _, pkg := range pkgs {
		if err := InstrumentPackage(fset, pkg); err != nil {
			return err
		}
	}
	return nil
}

func InstrumentPackage(fset *token.FileSet, pkg *ast.Package) error {
	for fileName, file := range pkg.Files {
		sourceFile, err := os.OpenFile(fileName, os.O_TRUNC|os.O_WRONLY, 0664)
		if err != nil {
			return fmt.Errorf("failed opening file %s: %v", fileName, err)
		}
		if err := InstrumentFile(fset, file, sourceFile); err != nil {
			return fmt.Errorf("failed instrumenting file %s: %v", fileName, err)
		}
	}
	return nil
}

func InstrumentFile(fset *token.FileSet, file *ast.File, out io.Writer) error {
	astutil.AddImport(fset, file, "fmt")

	// Needed because ast does not support floating comments and deletes them.
	// In order to preserve all comments we just pre-parse it to dst which treats them as first class citizens.
	f, err := decorator.DecorateFile(fset, file)
	if err != nil {
		return fmt.Errorf("failed converting file from ast to dst: %v", err)
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
	return decorator.Fprint(out, f)
}
