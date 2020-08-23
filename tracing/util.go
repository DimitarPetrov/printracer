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
