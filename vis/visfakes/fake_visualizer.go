// Code generated by counterfeiter. DO NOT EDIT.
package visfakes

import (
	"sync"

	"github.com/DimitarPetrov/printracer/parser"
	"github.com/DimitarPetrov/printracer/vis"
)

type FakeVisualizer struct {
	VisualizeStub        func([]parser.FuncEvent, int, string, string) error
	visualizeMutex       sync.RWMutex
	visualizeArgsForCall []struct {
		arg1 []parser.FuncEvent
		arg2 int
		arg3 string
		arg4 string
	}
	visualizeReturns struct {
		result1 error
	}
	visualizeReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeVisualizer) Visualize(arg1 []parser.FuncEvent, arg2 int, arg3 string, arg4 string) error {
	var arg1Copy []parser.FuncEvent
	if arg1 != nil {
		arg1Copy = make([]parser.FuncEvent, len(arg1))
		copy(arg1Copy, arg1)
	}
	fake.visualizeMutex.Lock()
	ret, specificReturn := fake.visualizeReturnsOnCall[len(fake.visualizeArgsForCall)]
	fake.visualizeArgsForCall = append(fake.visualizeArgsForCall, struct {
		arg1 []parser.FuncEvent
		arg2 int
		arg3 string
		arg4 string
	}{arg1Copy, arg2, arg3, arg4})
	fake.recordInvocation("Visualize", []interface{}{arg1Copy, arg2, arg3, arg4})
	fake.visualizeMutex.Unlock()
	if fake.VisualizeStub != nil {
		return fake.VisualizeStub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.visualizeReturns
	return fakeReturns.result1
}

func (fake *FakeVisualizer) VisualizeCallCount() int {
	fake.visualizeMutex.RLock()
	defer fake.visualizeMutex.RUnlock()
	return len(fake.visualizeArgsForCall)
}

func (fake *FakeVisualizer) VisualizeCalls(stub func([]parser.FuncEvent, int, string, string) error) {
	fake.visualizeMutex.Lock()
	defer fake.visualizeMutex.Unlock()
	fake.VisualizeStub = stub
}

func (fake *FakeVisualizer) VisualizeArgsForCall(i int) ([]parser.FuncEvent, int, string, string) {
	fake.visualizeMutex.RLock()
	defer fake.visualizeMutex.RUnlock()
	argsForCall := fake.visualizeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *FakeVisualizer) VisualizeReturns(result1 error) {
	fake.visualizeMutex.Lock()
	defer fake.visualizeMutex.Unlock()
	fake.VisualizeStub = nil
	fake.visualizeReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeVisualizer) VisualizeReturnsOnCall(i int, result1 error) {
	fake.visualizeMutex.Lock()
	defer fake.visualizeMutex.Unlock()
	fake.VisualizeStub = nil
	if fake.visualizeReturnsOnCall == nil {
		fake.visualizeReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.visualizeReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeVisualizer) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.visualizeMutex.RLock()
	defer fake.visualizeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeVisualizer) recordInvocation(key string, args []interface{}) {
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

var _ vis.Visualizer = new(FakeVisualizer)
