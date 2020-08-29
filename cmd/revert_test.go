package cmd

import (
	"errors"
	"github.com/DimitarPetrov/printracer/tracing/tracingfakes"
	"testing"
)

func TestRevertCmd(t *testing.T) {
	fakeDeinstrumenter := &tracingfakes.FakeCodeDeinstrumenter{}
	fakeImportsGroomer := &tracingfakes.FakeImportsGroomer{}
	cmd := NewRevertCmd(fakeDeinstrumenter, fakeImportsGroomer)

	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestRevertCmdReturnsErrorWhenDeinstrumenterReturnError(t *testing.T) {
	fakeDeinstrumenter := &tracingfakes.FakeCodeDeinstrumenter{}
	fakeImportsGroomer := &tracingfakes.FakeImportsGroomer{}
	cmd := NewRevertCmd(fakeDeinstrumenter, fakeImportsGroomer)

	expectedErr := errors.New("error")
	fakeDeinstrumenter.DeinstrumentDirectoryReturns(expectedErr)

	if err := cmd.Run(); err != expectedErr {
		t.Error("Assertion failed!")
	}
}

func TestRevertCmdReturnsErrorWhenImportsGroomerReturnError(t *testing.T) {
	fakeDeinstrumenter := &tracingfakes.FakeCodeDeinstrumenter{}
	fakeImportsGroomer := &tracingfakes.FakeImportsGroomer{}
	cmd := NewRevertCmd(fakeDeinstrumenter, fakeImportsGroomer)

	expectedErr := errors.New("error")
	fakeImportsGroomer.RemoveUnusedImportFromDirectoryReturns(expectedErr)

	if err := cmd.Run(); err != expectedErr {
		t.Error("Assertion failed!")
	}
}
