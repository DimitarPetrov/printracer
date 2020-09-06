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
const callerFuncNameVarName = "caller"
const callIDVarName = "callID"
const instrumentationStmtsCount = 9

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
	astutil.AddImport(fset, file, "runtime")
	astutil.AddImport(fset, file, "rand")

	// Needed because ast does not support floating comments and deletes them.
	// In order to preserve all comments we just pre-parse it to dst which treats them as first class citizens.
	f, err := decorator.DecorateFile(fset, file)
	if err != nil {
		return fmt.Errorf("failed converting file from ast to dst: %v", err)
	}

	dst.Inspect(f, func(n dst.Node) bool {
		switch t := n.(type) {
		case *dst.FuncDecl:
			var enteringStringFormat = "Function %s called by %s"
			var exitingStringFormat = "Exiting function %s called by %s; callID=%s"

			args := []dst.Expr{
				&dst.BasicLit{
					Kind:  token.STRING,
					Value: funcNameVarName,
				},
				&dst.BasicLit{
					Kind:  token.STRING,
					Value: callerFuncNameVarName,
				},
			}

			if len(t.Type.Params.List) > 0 {
				enteringStringFormat += " with args"

				for _, param := range t.Type.Params.List {
					enteringStringFormat += " (%v)"
					args = append(args, &dst.BasicLit{
						Kind:  token.STRING,
						Value: param.Names[0].Name,
					})
				}
			}
			args = append(args, &dst.BasicLit{
				Kind:  token.STRING,
				Value: callIDVarName,
			})
			args = append([]dst.Expr{
				&dst.BasicLit{
					Kind:  token.STRING,
					Value: `"` + enteringStringFormat + `; callID=%s\n"`,
				},
			}, args...)

			instrumentationStmts := [instrumentationStmtsCount]dst.Stmt{
				newAssignStmt(funcNameVarName, t.Name.Name),
				newAssignStmt(callerFuncNameVarName, "unknown"),
				newGetFuncNameIfStatement("0", "funcPC", funcNameVarName),
				newGetFuncNameIfStatement("1", "callerPC", callerFuncNameVarName),
				newMakeByteSliceStmt(),
				newRandReadStmt(),
				newParseUUIDFromByteSliceStmt(callIDVarName),
				&dst.ExprStmt{
					X: newPrintExprWithArgs(args),
				},
				&dst.DeferStmt{
					Call: newPrintExprWithArgs([]dst.Expr{
						&dst.BasicLit{
							Kind:  token.STRING,
							Value: `"` + exitingStringFormat + `\n"`,
						},
						&dst.BasicLit{
							Kind:  token.STRING,
							Value: funcNameVarName,
						},
						&dst.BasicLit{
							Kind:  token.STRING,
							Value: callerFuncNameVarName,
						},
						&dst.BasicLit{
							Kind:  token.STRING,
							Value: callIDVarName,
						},
					}),
				},
			}

			t.Body.List = append(instrumentationStmts[:], t.Body.List...)

			t.Body.List[0].Decorations().Before = dst.EmptyLine
			t.Body.List[0].Decorations().Start.Append("/* prinTracer */")
			t.Body.List[instrumentationStmtsCount-1].Decorations().After = dst.EmptyLine
			t.Body.List[instrumentationStmtsCount-1].Decorations().End.Append("/* prinTracer */")
		}
		return true
	})
	return decorator.Fprint(out, f)
}
