package parser

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	input := `Entering function main.main called by runtime.main; callID=1d8ca74e-c860-8a75-fc36-fe6d34350f0c
Entering function main.foo called by main.main with args (5) (false); callID=973355a9-2ec6-095c-9137-7a1081ac0a5f
Entering function main.bar called by main.foo with args (test string); callID=6c294dfd-4c6a-39b1-474e-314bee73f514
Entering function main.baz called by main.bar; callID=a019a297-0a6e-a792-0e3f-23c33a44622f
Exiting function main.baz called by main.bar; callID=a019a297-0a6e-a792-0e3f-23c33a44622f
Exiting function main.bar called by main.foo; callID=6c294dfd-4c6a-39b1-474e-314bee73f514
Exiting function main.foo called by main.main; callID=973355a9-2ec6-095c-9137-7a1081ac0a5f
Exiting function main.main called by runtime.main; callID=1d8ca74e-c860-8a75-fc36-fe6d34350f0c`

	expected := []FuncEvent{
		&InvocationEvent{
			Caller: "runtime.main",
			Callee: "main.main",
			CallID: "1d8ca74e-c860-8a75-fc36-fe6d34350f0c",
		},
		&InvocationEvent{
			Caller: "main.main",
			Callee: "main.foo",
			CallID: "973355a9-2ec6-095c-9137-7a1081ac0a5f",
			Args:   "with args (5) (false)",
		},
		&InvocationEvent{
			Caller: "main.foo",
			Callee: "main.bar",
			CallID: "6c294dfd-4c6a-39b1-474e-314bee73f514",
			Args:   "with args (test string)",
		},
		&InvocationEvent{
			Caller: "main.bar",
			Callee: "main.baz",
			CallID: "a019a297-0a6e-a792-0e3f-23c33a44622f",
		},
		&ReturningEvent{
			Caller: "main.bar",
			Callee: "main.baz",
			CallID: "a019a297-0a6e-a792-0e3f-23c33a44622f",
		},
		&ReturningEvent{
			Caller: "main.foo",
			Callee: "main.bar",
			CallID: "6c294dfd-4c6a-39b1-474e-314bee73f514",
		},
		&ReturningEvent{
			Caller: "main.main",
			Callee: "main.foo",
			CallID: "973355a9-2ec6-095c-9137-7a1081ac0a5f",
		},
		&ReturningEvent{
			Caller: "runtime.main",
			Callee: "main.main",
			CallID: "1d8ca74e-c860-8a75-fc36-fe6d34350f0c",
		},
	}

	actual, err := NewParser().Parse(bytes.NewBufferString(input))
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Error("Assertion Failed!")
	}
}
