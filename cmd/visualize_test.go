package cmd

import (
	"errors"
	"github.com/DimitarPetrov/printracer/parser/parserfakes"
	"github.com/DimitarPetrov/printracer/vis/visfakes"
	"strings"
	"testing"
)

func TestVisualizeCmd(t *testing.T) {
	fakeVisualizer := &visfakes.FakeVisualizer{}
	fakeParser := &parserfakes.FakeParser{}
	cmd := NewVisualizeCmd(fakeParser, fakeVisualizer).Prepare()

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestVisualizeCmdReturnsErrorWhenParserReturnError(t *testing.T) {
	fakeVisualizer := &visfakes.FakeVisualizer{}
	fakeParser := &parserfakes.FakeParser{}
	cmd := NewVisualizeCmd(fakeParser, fakeVisualizer).Prepare()

	expectedErr := errors.New("error")
	fakeParser.ParseReturns(nil, expectedErr)

	if err := cmd.Execute(); err == nil {
		t.Error("Expected error to have occured!")
	} else if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Error("Assertion failed!")
	}
}

func TestVisualizeCmdReturnsErrorWhenVisualizerReturnError(t *testing.T) {
	fakeVisualizer := &visfakes.FakeVisualizer{}
	fakeParser := &parserfakes.FakeParser{}
	cmd := NewVisualizeCmd(fakeParser, fakeVisualizer).Prepare()

	expectedErr := errors.New("error")
	fakeVisualizer.VisualizeReturns(expectedErr)

	if err := cmd.Execute(); err == nil {
		t.Error("Expected error to have occured!")
	} else if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Error("Assertion failed!")
	}
}

func TestVisualizeCmdErrorWhileValidatingArgs(t *testing.T) {
	fakeVisualizer := &visfakes.FakeVisualizer{}
	fakeParser := &parserfakes.FakeParser{}
	cmd := NewVisualizeCmd(fakeParser, fakeVisualizer)

	if err := cmd.Validate([]string{"test"}); err == nil {
		t.Error("Expected error to have occured!")
	} else if !strings.Contains(err.Error(), "error opening input file") {
		t.Error("Assertion failed!")
	}
}
