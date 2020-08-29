package cmd

import (
	"fmt"
	"github.com/DimitarPetrov/printracer/parser"
	"github.com/DimitarPetrov/printracer/vis"
	"github.com/spf13/cobra"
	"io"
	"math"
	"os"
)

type VisualizeCmd struct {
	input io.Reader

	outputFile   string
	maxDepth     int
	startingFunc string
}

func NewVisualizeCmd() *VisualizeCmd {
	return &VisualizeCmd{}
}

func (vc *VisualizeCmd) Prepare() *cobra.Command {
	result := &cobra.Command{
		Use:     "visualize",
		Aliases: []string{"v"},
		Short:   "Generates html sequence diagram of a given trace (file with output of already instrumented code).",
		PreRunE: commonPreRunE(vc),
		RunE:    commonRunE(vc),
	}

	result.Flags().StringVarP(&vc.outputFile, "output", "o", "calls", "name of the resulting html file when visualizing")
	result.Flags().IntVarP(&vc.maxDepth, "depth", "d", math.MaxInt32, "maximum depth in call graph")
	result.Flags().StringVarP(&vc.startingFunc, "func", "f", "", "name of the starting function in the visualization (the root of the diagram)")
	return result
}

func (vc *VisualizeCmd) Parse(args []string) error {
	vc.input = os.Stdin
	if len(args) > 0 {
		f, err := os.Open(args[0])
		if err != nil {
			return fmt.Errorf("error opening input file %s: %v", args[0], err)
		}
		vc.input = f
	}
	return nil
}

func (vc *VisualizeCmd) Run() error {
	traceParser := parser.NewParser(vc.input)
	events, err := traceParser.Parse()
	if err != nil {
		return fmt.Errorf("error while parsing input: %v", err)
	}
	if err := vis.Visualize(events, vc.maxDepth, vc.startingFunc, vc.outputFile); err != nil {
		return fmt.Errorf("error visualizing sequence diagram: %v", err)
	}
	return nil
}
