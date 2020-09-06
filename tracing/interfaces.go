package tracing

import (
	"go/ast"
	"go/token"
	"io"
)

//go:generate counterfeiter . CodeInstrumenter
type CodeInstrumenter interface {
	InstrumentFile(fset *token.FileSet, file *ast.File, out io.Writer) error
	InstrumentPackage(fset *token.FileSet, pkg *ast.Package) error
	InstrumentDirectory(path string) error
}

//go:generate counterfeiter . CodeDeinstrumenter
type CodeDeinstrumenter interface {
	DeinstrumentFile(fset *token.FileSet, file *ast.File, out io.Writer) error
	DeinstrumentPackage(fset *token.FileSet, pkg *ast.Package) error
	DeinstrumentDirectory(path string) error
}

//go:generate counterfeiter . ImportsGroomer
type ImportsGroomer interface {
	RemoveUnusedImportFromFile(fset *token.FileSet, file *ast.File, out io.Writer, importsToRemove []string) error
	RemoveUnusedImportFromPackage(fset *token.FileSet, pkg *ast.Package, importsToRemove []string) error
	RemoveUnusedImportFromDirectory(path string, importsToRemove []string) error
}
