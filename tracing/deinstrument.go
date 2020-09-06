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
	"reflect"
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
				if len(firstStmntDecorations) > 0 && firstStmntDecorations[0] == printracerCommentWatermark &&
					len(secondStmntDecorations) > 0 && secondStmntDecorations[0] == printracerCommentWatermark {

					if checkInstrumentationStatementsIntegrity(t) {
						t.Body.List = t.Body.List[instrumentationStmtsCount:]
						t.Body.List[0].Decorations().Before = dst.None
					}
				}
			}
		}
		return true
	})

	return decorator.Fprint(out, f)
}

func checkInstrumentationStatementsIntegrity(f *dst.FuncDecl) bool {
	stmts := f.Body.List
	instrumentationStmts := buildInstrumentationStmts(f)

	for i := 0; i < instrumentationStmtsCount; i++ {
		if !equalStmt(stmts[i], instrumentationStmts[i]) {
			return false
		}
	}
	return true
}

func equalStmt(stmt1, stmt2 dst.Stmt) bool {
	switch t := stmt1.(type) {
	case *dst.AssignStmt:
		instStmt, ok := stmt2.(*dst.AssignStmt)
		if !ok {
			return false
		}
		if !(equalExprSlice(t.Lhs, instStmt.Lhs) && equalExprSlice(t.Rhs, instStmt.Rhs) && reflect.DeepEqual(t.Tok, instStmt.Tok)) {
			return false
		}
		return true
	case *dst.IfStmt:
		instStmt, ok := stmt2.(*dst.IfStmt)
		if !ok {
			return false
		}
		if !(equalStmt(t.Init, instStmt.Init) && equalExpr(t.Cond, instStmt.Cond) && equalStmt(t.Body, instStmt.Body) && equalStmt(t.Else, instStmt.Else)) {
			return false
		}
		return true
	case *dst.ExprStmt:
		instStmt, ok := stmt2.(*dst.ExprStmt)
		if !ok {
			return false
		}
		if !(equalExpr(t.X, instStmt.X)) {
			return false
		}
		return true
	case *dst.DeferStmt:
		instStmt, ok := stmt2.(*dst.DeferStmt)
		if !ok {
			return false
		}
		if !(equalExpr(t.Call, instStmt.Call)) {
			return false
		}
		return true
	case *dst.BlockStmt:
		instStmt, ok := stmt2.(*dst.BlockStmt)
		if !ok {
			return false
		}
		if len(t.List) != len(instStmt.List) || t.RbraceHasNoPos != instStmt.RbraceHasNoPos {
			return false
		}
		for i, stmt1 := range t.List {
			if !equalStmt(stmt1, instStmt.List[i]) {
				return false
			}
		}
		return true
	}
	return reflect.DeepEqual(stmt1, stmt2)
}

func equalExprSlice(exprSlice1, exprSlice2 []dst.Expr) bool {
	if len(exprSlice1) != len(exprSlice2) {
		return false
	}
	for i, expr1 := range exprSlice1 {
		if !equalExpr(expr1, exprSlice2[i]) {
			return false
		}
	}
	return true
}

func equalExpr(expr1, expr2 dst.Expr) bool {
	switch t := expr1.(type) {
	case *dst.Ident:
		instExpr, ok := expr2.(*dst.Ident)
		if !ok {
			instExpr, ok := expr2.(*dst.BasicLit)
			if !ok {
				return false
			}
			return t.Name == instExpr.Value
		}
		return t.Name == instExpr.Name && t.Path == instExpr.Path
	case *dst.CallExpr:
		instExpr, ok := expr2.(*dst.CallExpr)
		if !ok {
			return false
		}
		if !(equalExprSlice(t.Args, instExpr.Args) && equalExpr(t.Fun, instExpr.Fun)) {
			return false
		}
		return true
	case *dst.SelectorExpr:
		instExpr, ok := expr2.(*dst.SelectorExpr)
		if !ok {
			return false
		}
		if !(equalExpr(t.X, instExpr.X) && equalExpr(t.Sel, instExpr.Sel)) {
			return false
		}
		return true
	case *dst.SliceExpr:
		instExpr, ok := expr2.(*dst.SliceExpr)
		if !ok {
			return false
		}
		if !(t.Slice3 == instExpr.Slice3 && equalExpr(t.X, instExpr.X) && equalExpr(t.High, instExpr.High) && equalExpr(t.Low, instExpr.Low) && equalExpr(t.Max, instExpr.Max)) {
			return false
		}
		return true
	case *dst.BasicLit:
		instExpr, ok := expr2.(*dst.BasicLit)
		if !ok {
			return false
		}
		return t.Value == instExpr.Value && t.Kind == instExpr.Kind
	}
	return reflect.DeepEqual(expr1, expr2)
}
