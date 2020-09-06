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
			if len(t.Body.List) >= instrumentationStmtsCount {
				firstStmntDecorations := t.Body.List[0].Decorations().Start.All()
				secondStmntDecorations := t.Body.List[instrumentationStmtsCount-1].Decorations().End.All()
				if len(firstStmntDecorations) > 0 && firstStmntDecorations[0] == "/* prinTracer */" &&
					len(secondStmntDecorations) > 0 && secondStmntDecorations[0] == "/* prinTracer */" {
					t.Body.List = t.Body.List[instrumentationStmtsCount:]
					t.Body.List[0].Decorations().Before = dst.None
				}
			}
		}
		return true
	})

	return decorator.Fprint(out, f)
}
