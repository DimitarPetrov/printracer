package vis

import (
	"bytes"
	"fmt"
	"github.com/DimitarPetrov/printracer/parser"
	"html/template"
	"os"
)

const reportTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="utf-8">
    <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.2/css/bootstrap.min.css">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/9.12.0/styles/github.min.css"/>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/underscore.js/1.8.3/underscore-min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/raphael/2.2.7/raphael.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/clipboard.js/2.0.4/clipboard.min.js"></script>
    <script src="https://bramp.github.io/js-sequence-diagrams/js/sequence-diagram-min.js"></script>
    <script src="https://code.jquery.com/jquery-3.3.1.slim.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/popper.js/1.14.3/umd/popper.min.js"></script>
    <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.1.2/js/bootstrap.min.js"></script>
    <style>
        body {
            padding-top: 2rem;
            padding-bottom: 2rem;
        }
    </style>
</head>
<body>
<div class="container-fluid">
    <div class="card text-center">
        <div class="card-body">
            <div id="diagram" class="justify-content-center"></div>
        </div>
    </div>
    <br><br>
    <p class="lead">Calls</p>
    <table class="table">
        <thead>
        <tr>
            <th scope="col">#</th>
            <th scope="col">Arguments</th>
        </tr>
        </thead>
        <tbody>
        {{ range $i, $e := .Args }}
        <tr id="arg-{{$i}}">
            <th scope="row">{{ inc $i }}</th>
            <td>
                <pre style="max-height: 1000px; margin-bottom: 0; border: 1px solid #eee;"><code id="event-message-{{$i}}">{{ $e }}</code></pre>
            </td>
        </tr>
        {{ end }}
        </tbody>
    </table>
</div>
<script>
    Diagram.parse("{{ .Diagram }}").drawSVG("diagram", {theme: 'simple', 'font-size': 14});
</script>
</body>
</html>
`

var templateFuncs = &template.FuncMap{
	"inc": func(i int) int {
		return i + 1
	},
}

//go:generate counterfeiter . Visualizer
type Visualizer interface {
	Visualize(events []parser.FuncEvent, maxDepth int, startingFunc string, outputFile string) error
}

type visualizer struct {
}

func NewVisualizer() Visualizer {
	return &visualizer{}
}

func (v *visualizer) Visualize(events []parser.FuncEvent, maxDepth int, startingFunc string, outputFile string) error {
	tmpl, err := template.New("sequenceDiagram").
		Funcs(*templateFuncs).
		Parse(reportTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	templateData, err := v.constructTemplateData(events, maxDepth, startingFunc)
	if err != nil {
		return err
	}

	var out bytes.Buffer
	err = tmpl.Execute(&out, templateData)
	if err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	fileName := fmt.Sprintf("%s.html", outputFile)
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("error creating file %s: %v", fileName, err)
	}

	_, err = f.WriteString(out.String())
	if err != nil {
		return fmt.Errorf("error writing diagram data to file %s: %v", fileName, err)
	}
	return nil
}

type stack []parser.FuncEvent

func (s *stack) Push(event parser.FuncEvent) {
	*s = append(*s, event)
}

func (s *stack) Pop() parser.FuncEvent {
	event := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return event
}

func (s *stack) Peek() parser.FuncEvent {
	return (*s)[len(*s)-1]
}

func (s *stack) Length() int {
	return len(*s)
}

func (s *stack) Empty() bool {
	return s.Length() == 0
}

type templateData struct {
	Args     []string
	Diagram  string
	MetaJSON template.JS
}

type sequenceDiagramData struct {
	data  bytes.Buffer
	count int
}

func (r *sequenceDiagramData) addFunctionInvocation(source string, target string) {
	r.addRecord(source, "->", target)
}

func (r *sequenceDiagramData) addFunctionReturn(source string, target string) {
	r.addRecord(source, "-->", target)
}

func (r *sequenceDiagramData) addRecord(source string, operation string, target string) {
	r.count++
	r.data.WriteString(fmt.Sprintf("%s%s%s: (%d)\n", source, operation, target, r.count))
}

func (r *sequenceDiagramData) String() string {
	return r.data.String()
}

func (v *visualizer) constructTemplateData(events []parser.FuncEvent, maxDepth int, startingFunc string) (templateData, error) {
	diagramData := &sequenceDiagramData{}

	if len(startingFunc) > 0 {
		for i := 0; i < len(events); i++ {
			if events[i].FuncName() == startingFunc {
				events = events[i:]
				break
			}
		}

		for i := 1; i < len(events); i++ {
			if events[i].FuncName() == startingFunc {
				events = events[:i+1]
				break
			}
		}
	}

	stack := stack(make([]parser.FuncEvent, 0, len(events)))

	var args []string

	for i := 0; i < len(events); i++ {
		event := events[i]
		switch event := event.(type) {
		case *parser.InvocationEvent:
			if stack.Length() < maxDepth {
				prev := event
				if !stack.Empty() {
					prev = stack.Peek().(*parser.InvocationEvent)
				}
				diagramData.addFunctionInvocation(prev.Name, event.FuncName())
				args = append(args, fmt.Sprintf("calling %s", event.Args))
				stack.Push(event)
			}
		case *parser.ReturningEvent:
			if stack.Peek().FuncName() == event.FuncName() {
				prev := stack.Pop()
				if !stack.Empty() {
					prev = stack.Peek()
				}
				diagramData.addFunctionReturn(event.FuncName(), prev.FuncName())
				args = append(args, "returning")
			}
		}
	}

	return templateData{
		Diagram: diagramData.String(),
		Args:    args,
	}, nil
}
