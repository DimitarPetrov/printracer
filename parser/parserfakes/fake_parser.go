// Code generated by counterfeiter. DO NOT EDIT.
package parserfakes

import (
	"io"
	"sync"

	"github.com/DimitarPetrov/printracer/parser"
)

type FakeParser struct {
	ParseStub        func(io.Reader) ([]parser.FuncEvent, error)
	parseMutex       sync.RWMutex
	parseArgsForCall []struct {
		arg1 io.Reader
	}
	parseReturns struct {
		result1 []parser.FuncEvent
		result2 error
	}
	parseReturnsOnCall map[int]struct {
		result1 []parser.FuncEvent
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeParser) Parse(arg1 io.Reader) ([]parser.FuncEvent, error) {
	fake.parseMutex.Lock()
	ret, specificReturn := fake.parseReturnsOnCall[len(fake.parseArgsForCall)]
	fake.parseArgsForCall = append(fake.parseArgsForCall, struct {
		arg1 io.Reader
	}{arg1})
	fake.recordInvocation("Parse", []interface{}{arg1})
	fake.parseMutex.Unlock()
	if fake.ParseStub != nil {
		return fake.ParseStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.parseReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeParser) ParseCallCount() int {
	fake.parseMutex.RLock()
	defer fake.parseMutex.RUnlock()
	return len(fake.parseArgsForCall)
}

func (fake *FakeParser) ParseCalls(stub func(io.Reader) ([]parser.FuncEvent, error)) {
	fake.parseMutex.Lock()
	defer fake.parseMutex.Unlock()
	fake.ParseStub = stub
}

func (fake *FakeParser) ParseArgsForCall(i int) io.Reader {
	fake.parseMutex.RLock()
	defer fake.parseMutex.RUnlock()
	argsForCall := fake.parseArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeParser) ParseReturns(result1 []parser.FuncEvent, result2 error) {
	fake.parseMutex.Lock()
	defer fake.parseMutex.Unlock()
	fake.ParseStub = nil
	fake.parseReturns = struct {
		result1 []parser.FuncEvent
		result2 error
	}{result1, result2}
}

func (fake *FakeParser) ParseReturnsOnCall(i int, result1 []parser.FuncEvent, result2 error) {
	fake.parseMutex.Lock()
	defer fake.parseMutex.Unlock()
	fake.ParseStub = nil
	if fake.parseReturnsOnCall == nil {
		fake.parseReturnsOnCall = make(map[int]struct {
			result1 []parser.FuncEvent
			result2 error
		})
	}
	fake.parseReturnsOnCall[i] = struct {
		result1 []parser.FuncEvent
		result2 error
	}{result1, result2}
}

func (fake *FakeParser) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.parseMutex.RLock()
	defer fake.parseMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeParser) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ parser.Parser = new(FakeParser)
