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

func TestDeinstrumentFile(t *testing.T) {
	tests := []struct {
		Name       string
		InputCode  string
		OutputCode string
	}{
		{Name: "DeinstrumentFileWithoutImports", InputCode: resultCodeWithoutImports, OutputCode: codeWithoutImports},
		{Name: "DeinstrumentFileWithFmtImportOnly", InputCode: resultCodeWithFmtImport, OutputCode: codeWithFmtImport},
		{Name: "DeinstrumentFileWithMultipleImports", InputCode: resultCodeWithMultipleImports, OutputCode: codeWithMultipleImports},
		{Name: "DeinstrumentFileWithoutFmtImport", InputCode: resultCodeWithImportsWithoutFmt, OutputCode: codeWithImportsWithoutFmt},
		{Name: "DeinstrumentFileWithoutFunctions", InputCode: resultCodeWithoutFunction, OutputCode: codeWithoutFunction},
		{Name: "DeinstrumentFileWithoutPreviousInstrumentation", InputCode: codeWithMultipleImports, OutputCode: codeWithMultipleImports},
		{Name: "DeinstrumentFileDoesNotChangeManuallyEditedFunctions", InputCode: editedResultCodeWithoutImports, OutputCode: editedResultCodeWithoutImports},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", test.InputCode, parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}
			var buff bytes.Buffer
			if err := NewCodeDeinstrumenter().DeinstrumentFile(fset, file, &buff); err != nil {
				t.Fatal(err)
			}

			file, err = parser.ParseFile(fset, "", buff.String(), parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}

			var buff2 bytes.Buffer
			if err := NewImportsGroomer().RemoveUnusedImportFromFile(fset, file, &buff2, map[string]string{"fmt":"", "runtime":"rt", "rand":""}); err != nil {
				t.Fatal(err)
			}

			if buff2.String() != test.OutputCode {
				t.Errorf("Assertion failed! Expected %s god %s", test.OutputCode, buff2.String())
			}
		})
	}
}

func TestDeinstrumentDirectory(t *testing.T) {
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
		{InputCode: resultCodeWithoutImports, OutputCode: codeWithoutImports},
		{InputCode: resultCodeWithFmtImport, OutputCode: codeWithFmtImport},
		{InputCode: resultCodeWithMultipleImports, OutputCode: codeWithMultipleImports},
		{InputCode: resultCodeWithImportsWithoutFmt, OutputCode: codeWithImportsWithoutFmt},
		{InputCode: resultCodeWithoutFunction, OutputCode: codeWithoutFunction},
	}

	i := 0
	for _, filePair := range filePairs {
		if err := ioutil.WriteFile(fmt.Sprintf("test/test%d.go", i), []byte(filePair.InputCode), 0777); err != nil {
			t.Fatal(err)
		}
		i++
	}

	if err := NewCodeDeinstrumenter().DeinstrumentDirectory("test"); err != nil {
		t.Fatal(err)
	}

	if err := NewImportsGroomer().RemoveUnusedImportFromDirectory("test", map[string]string{"fmt":"", "runtime":"rt", "rand":""}); err != nil {
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
