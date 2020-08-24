package main

import (
	"flag"
	"fmt"
	"github.com/DimitarPetrov/printracer/tracing"
	"log"
	"os"
	"path/filepath"
)

const apply = "apply"
const revert = "revert"

func init() {
	flag.Usage = func() {
		fmt.Fprintln(os.Stdout, `printracer instruments all go code in the current working directory to print every function execution along with its arguments.`)
		fmt.Fprintf(os.Stdout, "Usage: printracer [%s/%s]\n", apply, revert)
		flag.PrintDefaults()
	}
}

func main() {
	operation := parseOperation()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed getting current working directory: %v", err))
	}
	err = filepath.Walk(wd,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == "vendor" {
				return filepath.SkipDir
			}

			if info.IsDir() {
				switch operation {
				case apply:
					err := tracing.InstrumentDirectory(path)
					if err != nil {
						return err
					}
					return tracing.RemoveUnusedImportFromDirectory(path, "fmt")
				case revert:
					err := tracing.DeinstrumentDirectory(path)
					if err != nil {
						return err
					}
					return tracing.RemoveUnusedImportFromDirectory(path, "fmt")
				}
			}
			return nil
		})

	if err != nil {
		log.Fatal(fmt.Sprintf("Failed traversing directories: %v", err))
	}
}

func parseOperation() string {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Operation must be specified [%s/%s]. Use printracer --help for more information.\n", apply, revert)
		os.Exit(1)
	}
	operation := os.Args[1]
	if operation != apply && operation != revert {
		helpFlags := map[string]bool{
			"--help": true,
			"-help":  true,
			"--h":    true,
			"-h":     true,
		}
		if helpFlags[operation] {
			flag.Parse()
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Unsupported operation: %s. Only [%s/%s] operations are supported.\n Use stegify --help for more information.", operation, apply, revert)
		os.Exit(1)
	}

	os.Args = append(os.Args[:1], os.Args[2:]...) // needed because go flags implementation stop parsing after first non-flag argument
	return operation
}
