package cmd

import (
	"fmt"
	"github.com/DimitarPetrov/printracer/tracing"
	"github.com/spf13/cobra"
	"os"
)

type RevertCmd struct {
	deinstrumenter tracing.CodeDeinstrumenter
	importsGroomer tracing.ImportsGroomer
}

func NewRevertCmd(deinstrumenter tracing.CodeDeinstrumenter, importsGroomer tracing.ImportsGroomer) *RevertCmd {
	return &RevertCmd{
		deinstrumenter: deinstrumenter,
		importsGroomer: importsGroomer,
	}
}

func (rc *RevertCmd) Prepare() *cobra.Command {
	return &cobra.Command{
		Use:          "revert",
		Aliases:      []string{"r"},
		Short:        "Reverts previously instrumented directory of go files",
		PreRunE:      commonPreRunE(rc),
		RunE:         commonRunE(rc),
		SilenceUsage: true,
	}
}

func (rc *RevertCmd) Run() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting current working directory: %v", err)
	}

	return mapDirectory(wd, func(path string) error {
		err := rc.deinstrumenter.DeinstrumentDirectory(path)
		if err != nil {
			return err
		}
		return rc.importsGroomer.RemoveUnusedImportFromDirectory(path, map[string]string{"fmt":"", "runtime":"rt", "crypto/rand":""}) // TODO: flag for import aliases
	})
}
