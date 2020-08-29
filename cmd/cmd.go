package cmd

import (
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

type Command interface {
	Prepare() *cobra.Command
	Run() error
}

type CommandWithArgs interface {
	Validate([]string) error
}

func commonRunE(cmd Command) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		return cmd.Run()
	}
}

func commonPreRunE(cmd Command) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		if cmdWithArgs, ok := cmd.(CommandWithArgs); ok {
			if err := cmdWithArgs.Validate(args); err != nil {
				return err
			}
		}
		return nil
	}
}

func mapDirectory(dir string, operation func(string) error) error {
	return filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == "vendor" {
				return filepath.SkipDir
			}

			if info.IsDir() {
				return operation(path)
			}
			return nil
		})
}
