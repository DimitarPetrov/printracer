package cmd

import (
	"github.com/DimitarPetrov/printracer/parser"
	"github.com/DimitarPetrov/printracer/tracing"
	"github.com/DimitarPetrov/printracer/vis"
	"github.com/spf13/cobra"
)

type RootCmd struct {
	instrumenter   tracing.CodeInstrumenter
	deinstrumenter tracing.CodeDeinstrumenter
	importsGroomer tracing.ImportsGroomer
	parser         parser.Parser
	visualizer     vis.Visualizer
}

func NewRootCmd() *RootCmd {
	return &RootCmd{
		instrumenter:   tracing.NewCodeInstrumenter(),
		deinstrumenter: tracing.NewCodeDeinstrumenter(),
		importsGroomer: tracing.NewImportsGroomer(),
		parser:         parser.NewParser(),
		visualizer:     vis.NewVisualizer(),
	}
}

func (rc *RootCmd) Prepare() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "printracer",
		Short: "Printracer CLI",
		Long:  `printracer instruments every go file in the current working directory to print every function execution along with its arguments.`,
	}

	rootCmd.AddCommand(NewApplyCmd(rc.instrumenter, rc.importsGroomer).Prepare())
	rootCmd.AddCommand(NewRevertCmd(rc.deinstrumenter, rc.importsGroomer).Prepare())
	rootCmd.AddCommand(NewVisualizeCmd(rc.parser, rc.visualizer).Prepare())

	return rootCmd
}

func (rc *RootCmd) Run() error {
	return nil
}
