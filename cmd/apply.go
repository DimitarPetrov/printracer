package cmd

import (
	"fmt"
	"github.com/DimitarPetrov/printracer/tracing"
	"github.com/spf13/cobra"
	"os"
)

type ApplyCmd struct {
	instrumenter   tracing.CodeInstrumenter
	importsGroomer tracing.ImportsGroomer
}

func NewApplyCmd(instrumenter tracing.CodeInstrumenter, importsGroomer tracing.ImportsGroomer) *ApplyCmd {
	return &ApplyCmd{
		instrumenter:   instrumenter,
		importsGroomer: importsGroomer,
	}
}

func (ac *ApplyCmd) Prepare() *cobra.Command {
	return &cobra.Command{
		Use:          "apply",
		Aliases:      []string{"a"},
		Short:        "Instruments a directory of go files",
		PreRunE:      commonPreRunE(ac),
		RunE:         commonRunE(ac),
		SilenceUsage: true,
	}
}

func (ac *ApplyCmd) Run() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %v", err)
	}

	return mapDirectory(wd, func(path string) error {
		err := ac.instrumenter.InstrumentDirectory(path)
		if err != nil {
			return err
		}
		return ac.importsGroomer.RemoveUnusedImportFromDirectory(path, "fmt")
	})
}
