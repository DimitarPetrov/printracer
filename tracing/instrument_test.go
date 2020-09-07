package tracing

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"testing"
)

const codeWithoutImports = `package a

func test(i int, b bool) int {
	if b {
		return i
	}
	return 0
}

func main() {
	i := test(2, false)
}
`

const resultCodeWithoutImports = `package a

import (
	"crypto/rand"
	"fmt"
	rt "runtime"
)

func test(i int, b bool) int {

	/* prinTracer */
	funcName := "test"
	caller := "unknown"
	if funcPC, _, _, ok := rt.Caller(0); ok {
		funcName = rt.FuncForPC(funcPC).Name()
	}
	if callerPC, _, _, ok := rt.Caller(1); ok {
		caller = rt.FuncForPC(callerPC).Name()
	}
	idBytes := make([]byte, 16)
	_, _ = rand.Read(idBytes)
	callID := fmt.Sprintf("%x-%x-%x-%x-%x", idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:])
	fmt.Printf("Function %s called by %s with args (%v) (%v); callID=%s\n", funcName, caller, i, b, callID)
	defer fmt.Printf("Exiting function %s called by %s; callID=%s\n", funcName, caller, callID) /* prinTracer */

	if b {
		return i
	}
	return 0
}

func main() {

	/* prinTracer */
	funcName := "main"
	caller := "unknown"
	if funcPC, _, _, ok := rt.Caller(0); ok {
		funcName = rt.FuncForPC(funcPC).Name()
	}
	if callerPC, _, _, ok := rt.Caller(1); ok {
		caller = rt.FuncForPC(callerPC).Name()
	}
	idBytes := make([]byte, 16)
	_, _ = rand.Read(idBytes)
	callID := fmt.Sprintf("%x-%x-%x-%x-%x", idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:])
	fmt.Printf("Function %s called by %s; callID=%s\n", funcName, caller, callID)
	defer fmt.Printf("Exiting function %s called by %s; callID=%s\n", funcName, caller, callID) /* prinTracer */

	i := test(2, false)
}
`

const editedResultCodeWithoutImports = `package a

import (
	"crypto/rand"
	"fmt"
	rt "runtime"
)

func test(i int, b bool) int {

	/* prinTracer */
	funcName := "test2"
	caller := "unknown2"
	if funcPC, _, _, ok := rt.Caller(0); ok {
		funcName = rt.FuncForPC(funcPC).Name()
	}
	if callerPC, _, _, ok := rt.Caller(1); ok {
		caller = rt.FuncForPC(callerPC).Name()
	}
	fmt.Println("test")
	idBytes := make([]byte, 16)
	_, _ = rand.Read(idBytes)
	callID := fmt.Sprintf("%x-%x-%x-%x-%x", idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:])
	fmt.Printf("Function %s called by %s with args (%v) (%v); callID=%s\n", funcName, caller, i, b, callID)
	defer fmt.Printf("Exiting function %s called by %s; callID=%s\n", funcName, caller, callID) /* prinTracer */

	if b {
		return i
	}
	return 0
}

func main() {

	funcName := "main"
	caller := "unknown"
	if funcPC, _, _, ok := rt.Caller(0); ok {
		funcName = rt.FuncForPC(funcPC).Name()
	}
	if callerPC, _, _, ok := rt.Caller(1); ok {
		caller = rt.FuncForPC(callerPC).Name()
	}
	idBytes := make([]byte, 16)
	_, _ = rand.Read(idBytes)
	callID := fmt.Sprintf("%x-%x-%x-%x-%x", idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:])
	fmt.Printf("Function %s called by %s; callID=%s\n", funcName, caller, callID)
	defer fmt.Printf("Exiting function %s called by %s; callID=%s\n", funcName, caller, callID)

	i := test(2, false)
}
`

const codeWithFmtImport = `package a

import (
	"fmt"
)

func test(i int, b bool) int {
	if b {
		return i
	}
	return 0
}

func main() {
	i := test(2, false)
	fmt.Println(i)
}
`
const resultCodeWithFmtImport = `package a

import (
	"crypto/rand"
	"fmt"
	rt "runtime"
)

func test(i int, b bool) int {

	/* prinTracer */
	funcName := "test"
	caller := "unknown"
	if funcPC, _, _, ok := rt.Caller(0); ok {
		funcName = rt.FuncForPC(funcPC).Name()
	}
	if callerPC, _, _, ok := rt.Caller(1); ok {
		caller = rt.FuncForPC(callerPC).Name()
	}
	idBytes := make([]byte, 16)
	_, _ = rand.Read(idBytes)
	callID := fmt.Sprintf("%x-%x-%x-%x-%x", idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:])
	fmt.Printf("Function %s called by %s with args (%v) (%v); callID=%s\n", funcName, caller, i, b, callID)
	defer fmt.Printf("Exiting function %s called by %s; callID=%s\n", funcName, caller, callID) /* prinTracer */

	if b {
		return i
	}
	return 0
}

func main() {

	/* prinTracer */
	funcName := "main"
	caller := "unknown"
	if funcPC, _, _, ok := rt.Caller(0); ok {
		funcName = rt.FuncForPC(funcPC).Name()
	}
	if callerPC, _, _, ok := rt.Caller(1); ok {
		caller = rt.FuncForPC(callerPC).Name()
	}
	idBytes := make([]byte, 16)
	_, _ = rand.Read(idBytes)
	callID := fmt.Sprintf("%x-%x-%x-%x-%x", idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:])
	fmt.Printf("Function %s called by %s; callID=%s\n", funcName, caller, callID)
	defer fmt.Printf("Exiting function %s called by %s; callID=%s\n", funcName, caller, callID) /* prinTracer */

	i := test(2, false)
	fmt.Println(i)
}
`

const codeWithMultipleImports = `package a

import (
	"fmt"
	"strconv"
)

func test(i int, b bool) int {
	if b {
		return i
	}
	return 0
}

func main() {
	i := test(2, false)
	fmt.Println(strconv.Itoa(i))
}
`

const resultCodeWithMultipleImports = `package a

import (
	"crypto/rand"
	"fmt"
	rt "runtime"
	"strconv"
)

func test(i int, b bool) int {

	/* prinTracer */
	funcName := "test"
	caller := "unknown"
	if funcPC, _, _, ok := rt.Caller(0); ok {
		funcName = rt.FuncForPC(funcPC).Name()
	}
	if callerPC, _, _, ok := rt.Caller(1); ok {
		caller = rt.FuncForPC(callerPC).Name()
	}
	idBytes := make([]byte, 16)
	_, _ = rand.Read(idBytes)
	callID := fmt.Sprintf("%x-%x-%x-%x-%x", idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:])
	fmt.Printf("Function %s called by %s with args (%v) (%v); callID=%s\n", funcName, caller, i, b, callID)
	defer fmt.Printf("Exiting function %s called by %s; callID=%s\n", funcName, caller, callID) /* prinTracer */

	if b {
		return i
	}
	return 0
}

func main() {

	/* prinTracer */
	funcName := "main"
	caller := "unknown"
	if funcPC, _, _, ok := rt.Caller(0); ok {
		funcName = rt.FuncForPC(funcPC).Name()
	}
	if callerPC, _, _, ok := rt.Caller(1); ok {
		caller = rt.FuncForPC(callerPC).Name()
	}
	idBytes := make([]byte, 16)
	_, _ = rand.Read(idBytes)
	callID := fmt.Sprintf("%x-%x-%x-%x-%x", idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:])
	fmt.Printf("Function %s called by %s; callID=%s\n", funcName, caller, callID)
	defer fmt.Printf("Exiting function %s called by %s; callID=%s\n", funcName, caller, callID) /* prinTracer */

	i := test(2, false)
	fmt.Println(strconv.Itoa(i))
}
`

const codeWithImportsWithoutFmt = `package a

import (
	"strconv"
)

func test(i int, b bool) int {
	if b {
		return i
	}
	return 0
}

func main() {
	i := test(2, false)
	s := strconv.Itoa(i)
}
`

const resultCodeWithImportsWithoutFmt = `package a

import (
	"crypto/rand"
	"fmt"
	rt "runtime"
	"strconv"
)

func test(i int, b bool) int {

	/* prinTracer */
	funcName := "test"
	caller := "unknown"
	if funcPC, _, _, ok := rt.Caller(0); ok {
		funcName = rt.FuncForPC(funcPC).Name()
	}
	if callerPC, _, _, ok := rt.Caller(1); ok {
		caller = rt.FuncForPC(callerPC).Name()
	}
	idBytes := make([]byte, 16)
	_, _ = rand.Read(idBytes)
	callID := fmt.Sprintf("%x-%x-%x-%x-%x", idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:])
	fmt.Printf("Function %s called by %s with args (%v) (%v); callID=%s\n", funcName, caller, i, b, callID)
	defer fmt.Printf("Exiting function %s called by %s; callID=%s\n", funcName, caller, callID) /* prinTracer */

	if b {
		return i
	}
	return 0
}

func main() {

	/* prinTracer */
	funcName := "main"
	caller := "unknown"
	if funcPC, _, _, ok := rt.Caller(0); ok {
		funcName = rt.FuncForPC(funcPC).Name()
	}
	if callerPC, _, _, ok := rt.Caller(1); ok {
		caller = rt.FuncForPC(callerPC).Name()
	}
	idBytes := make([]byte, 16)
	_, _ = rand.Read(idBytes)
	callID := fmt.Sprintf("%x-%x-%x-%x-%x", idBytes[0:4], idBytes[4:6], idBytes[6:8], idBytes[8:10], idBytes[10:])
	fmt.Printf("Function %s called by %s; callID=%s\n", funcName, caller, callID)
	defer fmt.Printf("Exiting function %s called by %s; callID=%s\n", funcName, caller, callID) /* prinTracer */

	i := test(2, false)
	s := strconv.Itoa(i)
}
`

const codeWithWatermarks = `package a

import (
	"crypto/rand"
	"fmt"
	rt "runtime"
)

func test(i int, b bool) int {
	/* prinTracer */
	if b {
		return i
	}
	return 0
}

func main() {
	/* prinTracer */
	i := test(2, false)
}
`

const codeWithoutFunction = `package a

type test struct {
	a int
}
`

const resultCodeWithoutFunction = `package a

import (
	"crypto/rand"
	"fmt"
	rt "runtime"
)

type test struct {
	a int
}
`

func TestInstrumentFile(t *testing.T) {
	tests := []struct {
		Name       string
		InputCode  string
		OutputCode string
	}{
		{Name: "InstrumentFileWithoutImports", InputCode: codeWithoutImports, OutputCode: resultCodeWithoutImports},
		{Name: "InstrumentFileWithFmtImportOnly", InputCode: codeWithFmtImport, OutputCode: resultCodeWithFmtImport},
		{Name: "InstrumentFileWithMultipleImports", InputCode: codeWithMultipleImports, OutputCode: resultCodeWithMultipleImports},
		{Name: "InstrumentFileWithoutFmtImport", InputCode: codeWithImportsWithoutFmt, OutputCode: resultCodeWithImportsWithoutFmt},
		{Name: "InstrumentFileWithoutFunctions", InputCode: codeWithoutFunction, OutputCode: resultCodeWithoutFunction},
		{Name: "InstrumentFileDoesNotAffectAlreadyInstrumentedFiles", InputCode: resultCodeWithFmtImport, OutputCode: resultCodeWithFmtImport},
		{Name: "FunctionsWithWatermarksShouldNotBeInstrumented", InputCode: codeWithWatermarks, OutputCode: codeWithWatermarks},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", test.InputCode, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			var buff bytes.Buffer
			if err := NewCodeInstrumenter().InstrumentFile(fset, file, &buff); err != nil {
				t.Fatal(err)
			}

			if buff.String() != test.OutputCode {
				t.Errorf("Assertion failed! Expected %s got %s", test.OutputCode, buff.String())
			}
		})
	}
}

func TestInstrumentDirectory(t *testing.T) {
	if err := os.Mkdir("test", 0777); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll("test"); err != nil {
			t.Fatal(err)
		}
	}()

	filePairs := []struct {
		InputCode  string
		OutputCode string
	}{
		{InputCode: codeWithoutImports, OutputCode: resultCodeWithoutImports},
		{InputCode: codeWithFmtImport, OutputCode: resultCodeWithFmtImport},
		{InputCode: codeWithMultipleImports, OutputCode: resultCodeWithMultipleImports},
		{InputCode: codeWithImportsWithoutFmt, OutputCode: resultCodeWithImportsWithoutFmt},
		{InputCode: codeWithoutFunction, OutputCode: resultCodeWithoutFunction},
		{InputCode: resultCodeWithFmtImport, OutputCode: resultCodeWithFmtImport},
	}

	i := 0
	for _, filePair := range filePairs {
		if err := ioutil.WriteFile(fmt.Sprintf("test/test%d.go", i), []byte(filePair.InputCode), 0777); err != nil {
			t.Fatal(err)
		}
		i++
	}

	if err := NewCodeInstrumenter().InstrumentDirectory("test"); err != nil {
		t.Fatal(err)
	}

	i = 0
	for _, filePair := range filePairs {
		data, err := ioutil.ReadFile(fmt.Sprintf("test/test%d.go", i))
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != filePair.OutputCode {
			t.Error("Assertion failed!")
		}
		i++
	}
}
