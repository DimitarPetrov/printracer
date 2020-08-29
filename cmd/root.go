package cmd

import (
	"github.com/DimitarPetrov/printracer/tracing"
	"github.com/spf13/cobra"
)

type RootCmd struct {
}

func NewRootCmd() *RootCmd {
	return &RootCmd{}
}

func (rc *RootCmd) Prepare() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "printracer",
		Short: "Printracer CLI",
		Long:  `printracer instruments every go file in the current working directory to print every function execution along with its arguments.`,
	}

	rootCmd.AddCommand(NewApplyCmd(tracing.NewCodeInstrumenter(), tracing.NewImportsGroomer()).Prepare())
	rootCmd.AddCommand(NewRevertCmd(tracing.NewCodeDeinstrumenter(), tracing.NewImportsGroomer()).Prepare())
	rootCmd.AddCommand(NewVisualizeCmd().Prepare())

	return rootCmd
}

func (rc *RootCmd) Run() error {
	return nil
}
