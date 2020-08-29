package main

import (
	"github.com/DimitarPetrov/printracer/cmd"
	"log"
)

func main() {
	if err := cmd.NewRootCmd().Prepare().Execute(); err != nil {
		log.Fatal(err)
	}
}
