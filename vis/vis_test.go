package vis

import (
	"bytes"
	"github.com/DimitarPetrov/printracer/parser"
	"io/ioutil"
	"math"
	"os"
	"reflect"
	"strings"
	"testing"
)

var inputEvents = []parser.FuncEvent{
	&parser.InvocationEvent{
		Caller: "runtime.main",
		Callee: "main.main",
		CallID: "1d8ca74e-c860-8a75-fc36-fe6d34350f0c",
	},
	&parser.InvocationEvent{
		Caller: "main.main",
		Callee: "main.foo",
		CallID: "973355a9-2ec6-095c-9137-7a1081ac0a5f",
		Args:   "with args (5) (false)",
	},
	&parser.InvocationEvent{
		Caller: "main.foo",
		Callee: "main.bar",
		CallID: "6c294dfd-4c6a-39b1-474e-314bee73f514",
		Args:   "with args (test string)",
	},
	&parser.InvocationEvent{
		Caller: "main.bar",
		Callee: "main.baz",
		CallID: "a019a297-0a6e-a792-0e3f-23c33a44622f",
	},
	&parser.ReturningEvent{
		Caller: "main.bar",
		Callee: "main.baz",
		CallID: "a019a297-0a6e-a792-0e3f-23c33a44622f",
	},
	&parser.ReturningEvent{
		Caller: "main.foo",
		Callee: "main.bar",
		CallID: "6c294dfd-4c6a-39b1-474e-314bee73f514",
	},
	&parser.ReturningEvent{
		Caller: "main.main",
		Callee: "main.foo",
		CallID: "973355a9-2ec6-095c-9137-7a1081ac0a5f",
	},
	&parser.ReturningEvent{
		Caller: "runtime.main",
		Callee: "main.main",
		CallID: "1d8ca74e-c860-8a75-fc36-fe6d34350f0c",
	},
}

var fullDiagram = `"runtime.main"->"main.main": (1)
"main.main"->"main.foo": (2)
"main.foo"->"main.bar": (3)
"main.bar"->"main.baz": (4)
"main.baz"-->"main.bar": (5)
"main.bar"-->"main.foo": (6)
"main.foo"-->"main.main": (7)
"main.main"-->"runtime.main": (8)
`
var fullTableRows = []TableRow{
	{Args: "calling ", CallID: "1d8ca74e-c860-8a75-fc36-fe6d34350f0c"},
	{Args: "calling with args (5) (false)", CallID: "973355a9-2ec6-095c-9137-7a1081ac0a5f"},
	{Args: "calling with args (test string)", CallID: "6c294dfd-4c6a-39b1-474e-314bee73f514"},
	{Args: "calling ", CallID: "a019a297-0a6e-a792-0e3f-23c33a44622f"},
	{Args: "returning", CallID: "a019a297-0a6e-a792-0e3f-23c33a44622f"},
	{Args: "returning", CallID: "6c294dfd-4c6a-39b1-474e-314bee73f514"},
	{Args: "returning", CallID: "973355a9-2ec6-095c-9137-7a1081ac0a5f"},
	{Args: "returning", CallID: "1d8ca74e-c860-8a75-fc36-fe6d34350f0c"},
}

var diagramWith2DepthLimit = `"runtime.main"->"main.main": (1)
"main.main"->"main.foo": (2)
"main.foo"-->"main.main": (3)
"main.main"-->"runtime.main": (4)
`
var tableRowsWith2DepthLimit = []TableRow{
	{Args: "calling ", CallID: "1d8ca74e-c860-8a75-fc36-fe6d34350f0c"},
	{Args: "calling with args (5) (false)", CallID: "973355a9-2ec6-095c-9137-7a1081ac0a5f"},
	{Args: "returning", CallID: "973355a9-2ec6-095c-9137-7a1081ac0a5f"},
	{Args: "returning", CallID: "1d8ca74e-c860-8a75-fc36-fe6d34350f0c"},
}

var diagramWithFooStartingFunc = `"main.foo"->"main.bar": (1)
"main.bar"->"main.baz": (2)
"main.baz"-->"main.bar": (3)
"main.bar"-->"main.foo": (4)
`
var tableRowsWithFooStartingFunc = []TableRow{
	{Args: "calling with args (test string)", CallID: "6c294dfd-4c6a-39b1-474e-314bee73f514"},
	{Args: "calling ", CallID: "a019a297-0a6e-a792-0e3f-23c33a44622f"},
	{Args: "returning", CallID: "a019a297-0a6e-a792-0e3f-23c33a44622f"},
	{Args: "returning", CallID: "6c294dfd-4c6a-39b1-474e-314bee73f514"},
}

var diagramWithFooStartingFuncAnd2DepthLimit = `"main.foo"->"main.bar": (1)
"main.bar"-->"main.foo": (2)
`
var tableRowsWithFooStartingFuncAnd2DepthLimit = []TableRow{
	{Args: "calling with args (test string)", CallID: "6c294dfd-4c6a-39b1-474e-314bee73f514"},
	{Args: "returning", CallID: "6c294dfd-4c6a-39b1-474e-314bee73f514"},
}

func TestVisualizerConstructTemplateData(t *testing.T) {
	tests := []struct {
		Name         string
		MaxDepth     int
		StartingFunc string
		Diagram      string
		TableRows    []TableRow
	}{
		{Name: "ConstructTemplateData", MaxDepth: math.MaxInt32, Diagram: fullDiagram, TableRows: fullTableRows},
		{Name: "ConstructTemplateDataWithDepthLimit", MaxDepth: 2, Diagram: diagramWith2DepthLimit, TableRows: tableRowsWith2DepthLimit},
		{Name: "ConstructTemplateDataWithFooStartingFunc", MaxDepth: math.MaxInt32, StartingFunc: "main.foo", Diagram: diagramWithFooStartingFunc, TableRows: tableRowsWithFooStartingFunc},
		{Name: "ConstructTemplateDataWithFooStartingFuncAndDepthLimit", MaxDepth: 1, StartingFunc: "main.foo", Diagram: diagramWithFooStartingFuncAnd2DepthLimit, TableRows: tableRowsWithFooStartingFuncAnd2DepthLimit},
	}

	visualizer := visualizer{}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			diagramData, err := visualizer.constructTemplateData(inputEvents, test.MaxDepth, test.StartingFunc)
			if err != nil {
				t.Fatal(err)
			}
			if diagramData.Diagram != test.Diagram {
				t.Errorf("Assertion failed! Expected diagram data: %s bug got: %s", test.Diagram, diagramData.Diagram)
			}
			if !reflect.DeepEqual(diagramData.TableRows, test.TableRows) {
				t.Errorf("Assertion failed! Expected args: %v bug got: %v", test.TableRows, diagramData.TableRows)
			}
		})
	}
}

func TestVisualize(t *testing.T) {
	tests := []struct {
		Name         string
		MaxDepth     int
		StartingFunc string
		Diagram      string
		TableRows    []TableRow
	}{
		{Name: "ConstructTemplateData", MaxDepth: math.MaxInt32, Diagram: fullDiagram, TableRows: fullTableRows},
		{Name: "ConstructTemplateDataWithDepthLimit", MaxDepth: 2, Diagram: diagramWith2DepthLimit, TableRows: tableRowsWith2DepthLimit},
		{Name: "ConstructTemplateDataWithFooStartingFunc", MaxDepth: math.MaxInt32, StartingFunc: "main.foo", Diagram: diagramWithFooStartingFunc, TableRows: tableRowsWithFooStartingFunc},
		{Name: "ConstructTemplateDataWithFooStartingFuncAndDepthLimit", MaxDepth: 1, StartingFunc: "main.foo", Diagram: diagramWithFooStartingFuncAnd2DepthLimit, TableRows: tableRowsWithFooStartingFuncAnd2DepthLimit},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := NewVisualizer().Visualize(inputEvents, test.MaxDepth, test.StartingFunc, "test")
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.Remove("test.html"); err != nil {
					t.Fatal(err)
				}
			}()

			html, err := ioutil.ReadFile("test.html")
			if err != nil {
				t.Fatal(err)
			}

			encodedDiagram := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(test.Diagram, "->", `-\x3e`), "\n", `\n`), `"`, `\x22`)
			if !bytes.Contains(html, []byte(encodedDiagram)) {
				t.Error("Assertion failed! Expected html file to contain diagram data")
			}
			for _, row := range test.TableRows {
				if !bytes.Contains(html, []byte(row.Args)) {
					t.Errorf("Assertion failed! Expected html file to contain arg %s", row.Args)
				}
				if !bytes.Contains(html, []byte(row.CallID)) {
					t.Errorf("Assertion failed! Expected html file to contain callID %s", row.CallID)
				}
			}
		})
	}
}
