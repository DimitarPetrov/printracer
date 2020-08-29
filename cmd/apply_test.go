package cmd

import (
	"errors"
	"github.com/DimitarPetrov/printracer/tracing/tracingfakes"
	"testing"
)

func TestApplyCmd(t *testing.T) {
	fakeInstrumenter := &tracingfakes.FakeCodeInstrumenter{}
	fakeImportsGroomer := &tracingfakes.FakeImportsGroomer{}
	cmd := NewApplyCmd(fakeInstrumenter, fakeImportsGroomer).Prepare()

	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestApplyCmdReturnsErrorWhenInstrumenterReturnError(t *testing.T) {
	fakeInstrumenter := &tracingfakes.FakeCodeInstrumenter{}
	fakeImportsGroomer := &tracingfakes.FakeImportsGroomer{}
	cmd := NewApplyCmd(fakeInstrumenter, fakeImportsGroomer).Prepare()

	expectedErr := errors.New("error")
	fakeInstrumenter.InstrumentDirectoryReturns(expectedErr)

	if err := cmd.Execute(); err != expectedErr {
		t.Error("Assertion failed!")
	}
}

func TestApplyCmdReturnsErrorWhenImportsGroomerReturnError(t *testing.T) {
	fakeInstrumenter := &tracingfakes.FakeCodeInstrumenter{}
	fakeImportsGroomer := &tracingfakes.FakeImportsGroomer{}
	cmd := NewApplyCmd(fakeInstrumenter, fakeImportsGroomer).Prepare()

	expectedErr := errors.New("error")
	fakeImportsGroomer.RemoveUnusedImportFromDirectoryReturns(expectedErr)

	if err := cmd.Execute(); err != expectedErr {
		t.Error("Assertion failed!")
	}
}