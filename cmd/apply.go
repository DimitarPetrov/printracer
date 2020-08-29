package cmd

import (
	"fmt"
	"github.com/DimitarPetrov/printracer/tracing"
	"github.com/spf13/cobra"
	"os"
)

type ApplyCmd struct {
}

func NewApplyCmd() *ApplyCmd {
	return &ApplyCmd{}
}

func (ac *ApplyCmd) Prepare() *cobra.Command {
	return &cobra.Command{
		Use:     "apply",
		Aliases: []string{"a"},
		Short:   "Instruments a directory of go files",
		PreRunE: commonPreRunE(ac),
		RunE:    commonRunE(ac),
	}
}

func (ac *ApplyCmd) Run() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %v", err)
	}

	return mapDirectory(wd, func(path string) error {
		err := tracing.InstrumentDirectory(path)
		if err != nil {
			return err
		}
		return tracing.RemoveUnusedImportFromDirectory(path, "fmt")
	})
}
