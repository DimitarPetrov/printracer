package cmd

import (
	"fmt"
	"github.com/DimitarPetrov/printracer/parser"
	"github.com/DimitarPetrov/printracer/tracing"
	"github.com/DimitarPetrov/printracer/vis"
	"github.com/spf13/cobra"
	"math"
	"os"
	"path/filepath"
)

func BuildRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "printracer",
		Short: "Printracer CLI",
		Long:  `printracer instruments every go file in the current working directory to print every function execution along with its arguments.`,
	}

	rootCmd.AddCommand(buildApplyCommand())
	rootCmd.AddCommand(buildRevertCommand())
	rootCmd.AddCommand(buildVisualizeCommand())

	return rootCmd
}

func buildApplyCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "apply",
		Aliases: []string{"a"},
		Short:   "Instruments a directory of go files",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
}

func buildRevertCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "revert",
		Aliases: []string{"r"},
		Short:   "Reverts previously instrumented directory of go files",
		RunE: func(cmd *cobra.Command, args []string) error {
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
		},
	}
}

func buildVisualizeCommand() *cobra.Command {
	var outputFile string
	var maxDepth int
	var startingFunc string
	result := &cobra.Command{
		Use:     "visualize",
		Aliases: []string{"v"},
		Short:   "Generates html sequence diagram of a given trace (file with output of already instrumented code).",
		RunE: func(cmd *cobra.Command, args []string) error {
			in := os.Stdin
			var err error
			if len(args) > 0 {
				in, err = os.Open(args[0])
				if err != nil {
					return fmt.Errorf("error opening input file %s: %v", args[0], err)
				}
			}
			traceParser := parser.NewParser(in)
			events, err := traceParser.Parse()
			if err != nil {
				return fmt.Errorf("error while parsing input: %v", err)
			}
			if err := vis.Visualize(events, maxDepth, startingFunc, outputFile); err != nil {
				return fmt.Errorf("error visualizing sequence diagram: %v", err)
			}
			return nil
		},
	}

	result.Flags().StringVarP(&outputFile, "output", "o", "calls", "name of the resulting html file when visualizing")
	result.Flags().IntVarP(&maxDepth, "depth", "d", math.MaxInt32, "maximum depth in call graph")
	result.Flags().StringVarP(&startingFunc, "func", "f", "", "name of the starting function in the visualization (the root of the diagram)")
	return result
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
