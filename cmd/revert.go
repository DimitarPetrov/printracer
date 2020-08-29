package cmd

import (
	"fmt"
	"github.com/DimitarPetrov/printracer/tracing"
	"github.com/spf13/cobra"
	"os"
)

type RevertCmd struct {
}

func NewRevertCmd() *RevertCmd {
	return &RevertCmd{}
}

func (rc *RevertCmd) Prepare() *cobra.Command {
	return &cobra.Command{
		Use:     "revert",
		Aliases: []string{"r"},
		Short:   "Reverts previously instrumented directory of go files",
		PreRunE: commonPreRunE(rc),
		RunE:    commonRunE(rc),
	}
}

func (rc *RevertCmd) Run() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %v", err)
	}

	return mapDirectory(wd, func(path string) error {
		err := tracing.DeinstrumentDirectory(path)
		if err != nil {
			return err
		}
		return tracing.RemoveUnusedImportFromDirectory(path, "fmt")
	})
}
