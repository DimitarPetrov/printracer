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

func TestRemoveUnusedImportsFromFile(t *testing.T) {
	tests := []struct {
		Name       string
		InputCode  string
		OutputCode string
	}{
		{Name: "RemoveUnusedImportsFromFileWithoutFunctions", InputCode: resultCodeWithoutFunction, OutputCode: codeWithoutFunction},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", test.InputCode, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			var buff bytes.Buffer
			if err := NewImportsGroomer().RemoveUnusedImportFromFile(fset, file, &buff, "fmt"); err != nil {
				t.Fatal(err)
			}

			if buff.String() != test.OutputCode {
				t.Error("Assertion failed!")
			}
		})
	}
}

func TestRemoveUnusedImportsFromDirectory(t *testing.T) {
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
		{InputCode: resultCodeWithoutFunction, OutputCode: codeWithoutFunction},
	}

	i := 0
	for _, filePair := range filePairs {
		if err := ioutil.WriteFile(fmt.Sprintf("test/test%d.go", i), []byte(filePair.InputCode), 0777); err != nil {
			t.Fatal(err)
		}
		i++
	}

	if err := NewImportsGroomer().RemoveUnusedImportFromDirectory("test", "fmt"); err != nil {
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
