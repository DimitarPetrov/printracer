package parser

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	input := `Entering function main
Entering function foo with args (1) (true)
Entering function bar with args (test string)
Entering function baz
Exiting function baz
Exiting function bar
Exiting function foo
Exiting function main`

	expected := []FuncEvent{
		&InvocationEvent{
			Name: "main",
		},
		&InvocationEvent{
			Name: "foo",
			Args: "with args (1) (true)",
		},
		&InvocationEvent{
			Name: "bar",
			Args: "with args (test string)",
		},
		&InvocationEvent{
			Name: "baz",
		},
		&ReturningEvent{Name: "baz"},
		&ReturningEvent{Name: "bar"},
		&ReturningEvent{Name: "foo"},
		&ReturningEvent{Name: "main"},
	}

	actual, err := NewParser().Parse(bytes.NewBufferString(input))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Assertion Failed!")
	}
}
