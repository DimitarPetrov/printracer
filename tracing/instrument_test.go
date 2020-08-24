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

import "fmt"

func test(i int, b bool) int {
	fmt.Printf("Entering function test with args (%v) (%v)\n", i, b)
	defer fmt.Printf("Exiting function test\n")
	if b {
		return i
	}
	return 0
}

func main() {
	fmt.Printf("Entering function main\n")
	defer fmt.Printf("Exiting function main\n")
	i := test(2, false)
}
`

const codeWithFmtImport = `package a

import "fmt"

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

import "fmt"

func test(i int, b bool) int {
	fmt.Printf("Entering function test with args (%v) (%v)\n", i, b)
	defer fmt.Printf("Exiting function test\n")
	if b {
		return i
	}
	return 0
}

func main() {
	fmt.Printf("Entering function main\n")
	defer fmt.Printf("Exiting function main\n")
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
	"fmt"
	"strconv"
)

func test(i int, b bool) int {
	fmt.Printf("Entering function test with args (%v) (%v)\n", i, b)
	defer fmt.Printf("Exiting function test\n")
	if b {
		return i
	}
	return 0
}

func main() {
	fmt.Printf("Entering function main\n")
	defer fmt.Printf("Exiting function main\n")
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
	"fmt"
	"strconv"
)

func test(i int, b bool) int {
	fmt.Printf("Entering function test with args (%v) (%v)\n", i, b)
	defer fmt.Printf("Exiting function test\n")
	if b {
		return i
	}
	return 0
}

func main() {
	fmt.Printf("Entering function main\n")
	defer fmt.Printf("Exiting function main\n")
	i := test(2, false)
	s := strconv.Itoa(i)
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
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", test.InputCode, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			var buff bytes.Buffer
			if err := InstrumentFile(fset, file, &buff); err != nil {
				t.Fatal(err)
			}

			if buff.String() != test.OutputCode {
				t.Error("Assertion failed!")
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
	}

	i := 0
	for _, filePair := range filePairs {
		if err := ioutil.WriteFile(fmt.Sprintf("test/test%d.go", i), []byte(filePair.InputCode), 0777); err != nil {
			t.Fatal(err)
		}
		i++
	}

	if err := InstrumentDirectory("test"); err != nil {
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
