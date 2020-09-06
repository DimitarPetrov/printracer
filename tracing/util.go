package tracing

import (
	"bufio"
	"fmt"
	"github.com/dave/dst"
	"go/token"
	"os"
	"strings"
)

// Filter excluding go test files from directory
func testsFilter(info os.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}

// Filter excluding generated go files from directory.
// Generated file is considered a file which matches one of the following:
// 1. The name of the file contains "generated"
// 2. First line of the file contains "generated" or "GENERATED"
func generatedFilter(path string, info os.FileInfo) bool {
	if strings.Contains(info.Name(), "generated") {
		return false
	}

	f, err := os.Open(path + "/" + info.Name())
	if err != nil {
		panic(fmt.Sprintf("Failed opening file %s: %v", path+"/"+info.Name(), err))
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

// Returns dst expresion like: fmt.Printf("msg\n")
func newPrintExprWithMessage(msg string) *dst.CallExpr {
	return newPrintExprWithArgs([]dst.Expr{
		&dst.BasicLit{
			Kind:  token.STRING,
			Value: `"` + msg + `\n"`,
		},
	})
}

// Return dst expresion like: fmt.Printf(args...)
func newPrintExprWithArgs(args []dst.Expr) *dst.CallExpr {
	return &dst.CallExpr{
		Fun: &dst.SelectorExpr{
			X:   &dst.Ident{Name: "fmt"},
			Sel: &dst.Ident{Name: "Printf"},
		},
		Args: args,
	}
}

// Return dst statement like: varName := "value"
func newAssignStmt(varName, value string) *dst.AssignStmt {
	return &dst.AssignStmt{
		Lhs: []dst.Expr{
			&dst.Ident{
				Name: varName,
			},
		},
		Tok: token.DEFINE,
		Rhs: []dst.Expr{
			&dst.BasicLit{
				Kind:  token.STRING,
				Value: `"` + value + `"`,
			},
		},
	}
}

/* Return dst statement like:
if funcPcVarName, _, _, ok := runtime.Caller(funcIndex); ok {
	funcNameVarName = runtime.FuncForPC(funcPcVarName).Name()
}
*/
func newGetFuncNameIfStatement(funcIndex, funcPcVarName, funcNameVarName string) *dst.IfStmt {
	return &dst.IfStmt{
		Init: &dst.AssignStmt{
			Lhs: []dst.Expr{
				&dst.Ident{
					Name: funcPcVarName,
				},
				&dst.Ident{
					Name: "_",
				},
				&dst.Ident{
					Name: "_",
				},
				&dst.Ident{
					Name: "ok",
				},
			},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.CallExpr{
					Fun: &dst.SelectorExpr{
						X: &dst.Ident{
							Name: "runtime",
						},
						Sel: &dst.Ident{
							Name: "Caller",
						},
					},
					Args: []dst.Expr{
						&dst.BasicLit{
							Kind:  token.INT,
							Value: funcIndex,
						},
					},
				},
			},
		},
		Cond: &dst.Ident{
			Name: "ok",
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.AssignStmt{
					Lhs: []dst.Expr{
						&dst.Ident{
							Name: funcNameVarName,
						},
					},
					Tok: token.ASSIGN,
					Rhs: []dst.Expr{
						&dst.CallExpr{
							Fun: &dst.SelectorExpr{
								X: &dst.CallExpr{
									Fun: &dst.SelectorExpr{
										X: &dst.Ident{
											Name: "runtime",
										},
										Sel: &dst.Ident{
											Name: "FuncForPC",
										},
									},
									Args: []dst.Expr{
										&dst.Ident{
											Name: funcPcVarName,
										},
									},
								},
								Sel: &dst.Ident{
									Name: "Name",
								},
							},
						},
					},
				},
			},
		},
	}
}

// Returns dst statement like:
// idBytes := make([]byte, 16)
func newMakeByteSliceStmt() *dst.AssignStmt {
	return &dst.AssignStmt{
		Lhs: []dst.Expr{
			&dst.Ident{
				Name: "idBytes",
			},
		},
		Tok: token.DEFINE,
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun: &dst.Ident{
					Name: "make",
				},
				Args: []dst.Expr{
					&dst.ArrayType{
						Elt: &dst.Ident{
							Name: "byte",
						},
					},
					&dst.BasicLit{
						Kind:  token.INT,
						Value: "16",
					},
				},
			},
		},
	}
}

func newRandReadStmt() *dst.AssignStmt {
	return &dst.AssignStmt{
		Lhs: []dst.Expr{
			&dst.Ident{
				Name: "_",
			},
			&dst.Ident{
				Name: "_",
			},
		},
		Tok: token.ASSIGN,
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun: &dst.SelectorExpr{
					X: &dst.Ident{
						Name: "rand",
					},
					Sel: &dst.Ident{
						Name: "Read",
					},
				},
				Args: []dst.Expr{
					&dst.Ident{
						Name: "idBytes",
					},
				},
			},
		},
	}
}

func newParseUUIDFromByteSliceStmt(callIDVarName string) *dst.AssignStmt {
	return &dst.AssignStmt{
		Lhs: []dst.Expr{
			&dst.Ident{
				Name: callIDVarName,
			},
		},
		Tok: token.DEFINE,
		Rhs: []dst.Expr{
			&dst.CallExpr{
				Fun: &dst.SelectorExpr{
					X: &dst.Ident{
						Name: "fmt",
					},
					Sel: &dst.Ident{
						Name: "Sprintf",
					},
				},
				Args: []dst.Expr{
					&dst.BasicLit{
						Kind:  token.STRING,
						Value: "\"%x-%x-%x-%x-%x\"",
					},
					&dst.SliceExpr{
						X: &dst.Ident{
							Name: "idBytes",
						},
						Low: &dst.BasicLit{
							Kind:  token.INT,
							Value: "0",
						},
						High: &dst.BasicLit{
							Kind:  token.INT,
							Value: "4",
						},
						Slice3: false,
					},
					&dst.SliceExpr{
						X: &dst.Ident{
							Name: "idBytes",
						},
						Low: &dst.BasicLit{
							Kind:  token.INT,
							Value: "4",
						},
						High: &dst.BasicLit{
							Kind:  token.INT,
							Value: "6",
						},
						Slice3: false,
					},
					&dst.SliceExpr{
						X: &dst.Ident{
							Name: "idBytes",
						},
						Low: &dst.BasicLit{
							Kind:  token.INT,
							Value: "6",
						},
						High: &dst.BasicLit{
							Kind:  token.INT,
							Value: "8",
						},
						Slice3: false,
					},
					&dst.SliceExpr{
						X: &dst.Ident{
							Name: "idBytes",
						},
						Low: &dst.BasicLit{
							Kind:  token.INT,
							Value: "8",
						},
						High: &dst.BasicLit{
							Kind:  token.INT,
							Value: "10",
						},
						Slice3: false,
					},
					&dst.SliceExpr{
						X: &dst.Ident{
							Name: "idBytes",
						},
						Low: &dst.BasicLit{
							Kind:  token.INT,
							Value: "10",
						},
						Slice3: false,
					},
				},
			},
		},
	}
}
