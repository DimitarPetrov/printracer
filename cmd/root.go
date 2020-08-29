package cmd

import "github.com/spf13/cobra"

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

	rootCmd.AddCommand(NewApplyCmd().Prepare())
	rootCmd.AddCommand(NewRevertCmd().Prepare())
	rootCmd.AddCommand(NewVisualizeCmd().Prepare())

	return rootCmd
}

func (rc *RootCmd) Run() error {
	return nil
}
