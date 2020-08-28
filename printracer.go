package main

import (
	"github.com/DimitarPetrov/printracer/cmd"
	"log"
)

func main() {
	if err := cmd.BuildRootCommand().Execute(); err != nil {
		log.Fatal(err)
	}
}
