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

const funcNameVarName = "funcName"
const funcPCVarName = "funcPC"

const callerFuncNameVarName = "caller"
const defaultCallerName = "unknown"
const callerFuncPCVarName = "callerPC"

const callIDVarName = "callID"

const printracerCommentWatermark = "/* prinTracer */"

const instrumentationStmtsCount = 9 // Acts like a contract of how many statements instrumentation adds and deinstrumentation removes.

func buildInstrumentationStmts(f *dst.FuncDecl) [instrumentationStmtsCount]dst.Stmt {
	return [instrumentationStmtsCount]dst.Stmt{
		newAssignStmt(funcNameVarName, f.Name.Name),
		newAssignStmt(callerFuncNameVarName, defaultCallerName),
		newGetFuncNameIfStatement("0", funcPCVarName, funcNameVarName),
		newGetFuncNameIfStatement("1", callerFuncPCVarName, callerFuncNameVarName),
		newMakeByteSliceStmt(),
		newRandReadStmt(),
		newParseUUIDFromByteSliceStmt(callIDVarName),
		&dst.ExprStmt{
			X: newPrintExprWithArgs(buildEnteringFunctionArgs(f)),
		},
		&dst.DeferStmt{
			Call: newPrintExprWithArgs(buildExitFunctionArgs()),
		},
	}
}

type codeInstrumenter struct {
}

func NewCodeInstrumenter() CodeInstrumenter {
	return &codeInstrumenter{}
}

func (ci *codeInstrumenter) InstrumentDirectory(path string) error {
	fset := token.NewFileSet()
	filter := func(info os.FileInfo) bool {
		return testsFilter(info) && generatedFilter(path, info)
	}
	pkgs, err := parser.ParseDir(fset, path, filter, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed parsing go files in directory %s: %v", path, err)
	}

	for _, pkg := range pkgs {
		if err := ci.InstrumentPackage(fset, pkg); err != nil {
			return err
		}
	}
	return nil
}

func (ci *codeInstrumenter) InstrumentPackage(fset *token.FileSet, pkg *ast.Package) error {
	for fileName, file := range pkg.Files {
		sourceFile, err := os.OpenFile(fileName, os.O_TRUNC|os.O_WRONLY, 0664)
		if err != nil {
			return fmt.Errorf("failed opening file %s: %v", fileName, err)
		}
		if err := ci.InstrumentFile(fset, file, sourceFile); err != nil {
			return fmt.Errorf("failed instrumenting file %s: %v", fileName, err)
		}
	}
	return nil
}

func (ci *codeInstrumenter) InstrumentFile(fset *token.FileSet, file *ast.File, out io.Writer) error {
	astutil.AddImport(fset, file, "fmt")
	astutil.AddNamedImport(fset, file, "rt", "runtime")
	astutil.AddImport(fset, file, "crypto/rand")

	// Needed because ast does not support floating comments and deletes them.
	// In order to preserve all comments we just pre-parse it to dst which treats them as first class citizens.
	f, err := decorator.DecorateFile(fset, file)
	if err != nil {
		return fmt.Errorf("failed converting file from ast to dst: %v", err)
	}

	dst.Inspect(f, func(n dst.Node) bool {
		switch t := n.(type) {
		case *dst.FuncDecl:
			instrumentationStmts := buildInstrumentationStmts(t)
			t.Body.List = append(instrumentationStmts[:], t.Body.List...)

			t.Body.List[0].Decorations().Before = dst.EmptyLine
			t.Body.List[0].Decorations().Start.Append(printracerCommentWatermark)
			t.Body.List[instrumentationStmtsCount-1].Decorations().After = dst.EmptyLine
			t.Body.List[instrumentationStmtsCount-1].Decorations().End.Append(printracerCommentWatermark)
		}
		return true
	})
	return decorator.Fprint(out, f)
}
