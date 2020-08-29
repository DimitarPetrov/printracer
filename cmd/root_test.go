package cmd

import (
	"testing"
)

func TestRootCmd(t *testing.T) {
	cmd := NewRootCmd().Prepare()
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}
