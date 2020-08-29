package tracing

import (
	"fmt"
	"github.com/dave/dst/decorator"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"io"
	"os"
)

type importsGroomer struct {
}

func NewImportsGroomer() ImportsGroomer {
	return &importsGroomer{}
}

func (ig *importsGroomer) RemoveUnusedImportFromDirectory(path string, importToRemove string) error {
	fset := token.NewFileSet()
	filter := func(info os.FileInfo) bool {
		return testsFilter(info) && generatedFilter(path, info)
	}
	pkgs, err := parser.ParseDir(fset, path, filter, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed parsing go files in directory %s: %v", path, err)
	}

	for _, pkg := range pkgs {
		if err := ig.RemoveUnusedImportFromPackage(fset, pkg, importToRemove); err != nil {
			return err
		}
	}
	return nil
}

func (ig *importsGroomer) RemoveUnusedImportFromPackage(fset *token.FileSet, pkg *ast.Package, importToRemove string) error {
	for fileName, file := range pkg.Files {
		sourceFile, err := os.OpenFile(fileName, os.O_TRUNC|os.O_WRONLY, 0664)
		if err != nil {
			return fmt.Errorf("failed opening file %s: %v", fileName, err)
		}
		if err := ig.RemoveUnusedImportFromFile(fset, file, sourceFile, importToRemove); err != nil {
			return fmt.Errorf("failed removing import %s from file %s: %v", importToRemove, fileName, err)
		}
	}
	return nil
}

func (ig *importsGroomer) RemoveUnusedImportFromFile(fset *token.FileSet, file *ast.File, out io.Writer, importToRemove string) error {
	if !astutil.UsesImport(file, importToRemove) {
		astutil.DeleteImport(fset, file, importToRemove)
	}
	// Needed because ast does not support floating comments and deletes them.
	// In order to preserve all comments we just pre-parse it to dst which treats them as first class citizens.
	f, err := decorator.DecorateFile(fset, file)
	if err != nil {
		return fmt.Errorf("failed converting file from ast to dst: %v", err)
	}

	return decorator.Fprint(out, f)
}
