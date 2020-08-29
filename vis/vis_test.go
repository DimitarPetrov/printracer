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
		Name: "main",
	},
	&parser.InvocationEvent{
		Name: "foo",
		Args: "with args (1) (true)",
	},
	&parser.InvocationEvent{
		Name: "bar",
		Args: "with args (test string)",
	},
	&parser.InvocationEvent{
		Name: "baz",
	},
	&parser.ReturningEvent{Name: "baz"},
	&parser.ReturningEvent{Name: "bar"},
	&parser.ReturningEvent{Name: "foo"},
	&parser.ReturningEvent{Name: "main"},
}

var fullDiagram = `main->main: (1)
main->foo: (2)
foo->bar: (3)
bar->baz: (4)
baz-->bar: (5)
bar-->foo: (6)
foo-->main: (7)
main-->main: (8)
`
var fullArgs = []string{
	"calling ",
	"calling with args (1) (true)",
	"calling with args (test string)",
	"calling ",
	"returning",
	"returning",
	"returning",
	"returning",
}

var diagramWith2DepthLimit = `main->main: (1)
main->foo: (2)
foo-->main: (3)
main-->main: (4)
`
var argsWith2DepthLimit = []string{
	"calling ",
	"calling with args (1) (true)",
	"returning",
	"returning",
}

var diagramWithFooStartingFunc = `foo->foo: (1)
foo->bar: (2)
bar->baz: (3)
baz-->bar: (4)
bar-->foo: (5)
foo-->foo: (6)
`
var argsWithFooStartingFunc = []string{
	"calling with args (1) (true)",
	"calling with args (test string)",
	"calling ",
	"returning",
	"returning",
	"returning",
}

var diagramWithFooStartingFuncAnd2DepthLimit = `foo->foo: (1)
foo->bar: (2)
bar-->foo: (3)
foo-->foo: (4)
`
var argsWithFooStartingFuncAnd2DepthLimit = []string{
	"calling with args (1) (true)",
	"calling with args (test string)",
	"returning",
	"returning",
}

func TestVisualizerConstructTemplateData(t *testing.T) {
	tests := []struct {
		Name         string
		MaxDepth     int
		StartingFunc string
		Diagram      string
		Args         []string
	}{
		{Name: "ConstructTemplateData", MaxDepth: math.MaxInt32, Diagram: fullDiagram, Args: fullArgs},
		{Name: "ConstructTemplateDataWithDepthLimit", MaxDepth: 2, Diagram: diagramWith2DepthLimit, Args: argsWith2DepthLimit},
		{Name: "ConstructTemplateDataWithFooStartingFunc", MaxDepth: math.MaxInt32, StartingFunc: "foo", Diagram: diagramWithFooStartingFunc, Args: argsWithFooStartingFunc},
		{Name: "ConstructTemplateDataWithFooStartingFuncAndDepthLimit", MaxDepth: 2, StartingFunc: "foo", Diagram: diagramWithFooStartingFuncAnd2DepthLimit, Args: argsWithFooStartingFuncAnd2DepthLimit},
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
			if !reflect.DeepEqual(diagramData.Args, test.Args) {
				t.Errorf("Assertion failed! Expected args: %v bug got: %v", test.Args, diagramData.Args)
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
		Args         []string
	}{
		{Name: "ConstructTemplateData", MaxDepth: math.MaxInt32, Diagram: fullDiagram, Args: fullArgs},
		{Name: "ConstructTemplateDataWithDepthLimit", MaxDepth: 2, Diagram: diagramWith2DepthLimit, Args: argsWith2DepthLimit},
		{Name: "ConstructTemplateDataWithFooStartingFunc", MaxDepth: math.MaxInt32, StartingFunc: "foo", Diagram: diagramWithFooStartingFunc, Args: argsWithFooStartingFunc},
		{Name: "ConstructTemplateDataWithFooStartingFuncAndDepthLimit", MaxDepth: 2, StartingFunc: "foo", Diagram: diagramWithFooStartingFuncAnd2DepthLimit, Args: argsWithFooStartingFuncAnd2DepthLimit},
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

			encodedDiagram := strings.ReplaceAll(strings.ReplaceAll(test.Diagram, "->", `-\x3e`), "\n", `\n`)
			if !bytes.Contains(html, []byte(encodedDiagram)) {
				t.Error("Assertion failed! Expected html file to contain diagram data")
			}
			for _, arg := range test.Args {
				if !bytes.Contains(html, []byte(arg)) {
					t.Errorf("Assertion failed! Expected html file to contain arg %s", arg)
				}
			}
		})
	}
}
