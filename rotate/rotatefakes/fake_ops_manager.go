// Code generated by counterfeiter. DO NOT EDIT.
package rotatefakes

import (
	"io"
	"sync"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/rotate"
)

type FakeOpsManager struct {
	ApplyChangesStub        func(io.Writer, bool, ...string) error
	applyChangesMutex       sync.RWMutex
	applyChangesArgsForCall []struct {
		arg1 io.Writer
		arg2 bool
		arg3 []string
	}
	applyChangesReturns struct {
		result1 error
	}
	applyChangesReturnsOnCall map[int]struct {
		result1 error
	}
	CheckPendingChangesStub        func() (bool, error)
	checkPendingChangesMutex       sync.RWMutex
	checkPendingChangesArgsForCall []struct {
	}
	checkPendingChangesReturns struct {
		result1 bool
		result2 error
	}
	checkPendingChangesReturnsOnCall map[int]struct {
		result1 bool
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeOpsManager) ApplyChanges(arg1 io.Writer, arg2 bool, arg3 ...string) error {
	fake.applyChangesMutex.Lock()
	ret, specificReturn := fake.applyChangesReturnsOnCall[len(fake.applyChangesArgsForCall)]
	fake.applyChangesArgsForCall = append(fake.applyChangesArgsForCall, struct {
		arg1 io.Writer
		arg2 bool
		arg3 []string
	}{arg1, arg2, arg3})
	stub := fake.ApplyChangesStub
	fakeReturns := fake.applyChangesReturns
	fake.recordInvocation("ApplyChanges", []interface{}{arg1, arg2, arg3})
	fake.applyChangesMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3...)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeOpsManager) ApplyChangesCallCount() int {
	fake.applyChangesMutex.RLock()
	defer fake.applyChangesMutex.RUnlock()
	return len(fake.applyChangesArgsForCall)
}

func (fake *FakeOpsManager) ApplyChangesCalls(stub func(io.Writer, bool, ...string) error) {
	fake.applyChangesMutex.Lock()
	defer fake.applyChangesMutex.Unlock()
	fake.ApplyChangesStub = stub
}

func (fake *FakeOpsManager) ApplyChangesArgsForCall(i int) (io.Writer, bool, []string) {
	fake.applyChangesMutex.RLock()
	defer fake.applyChangesMutex.RUnlock()
	argsForCall := fake.applyChangesArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeOpsManager) ApplyChangesReturns(result1 error) {
	fake.applyChangesMutex.Lock()
	defer fake.applyChangesMutex.Unlock()
	fake.ApplyChangesStub = nil
	fake.applyChangesReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeOpsManager) ApplyChangesReturnsOnCall(i int, result1 error) {
	fake.applyChangesMutex.Lock()
	defer fake.applyChangesMutex.Unlock()
	fake.ApplyChangesStub = nil
	if fake.applyChangesReturnsOnCall == nil {
		fake.applyChangesReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.applyChangesReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeOpsManager) CheckPendingChanges() (bool, error) {
	fake.checkPendingChangesMutex.Lock()
	ret, specificReturn := fake.checkPendingChangesReturnsOnCall[len(fake.checkPendingChangesArgsForCall)]
	fake.checkPendingChangesArgsForCall = append(fake.checkPendingChangesArgsForCall, struct {
	}{})
	stub := fake.CheckPendingChangesStub
	fakeReturns := fake.checkPendingChangesReturns
	fake.recordInvocation("CheckPendingChanges", []interface{}{})
	fake.checkPendingChangesMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeOpsManager) CheckPendingChangesCallCount() int {
	fake.checkPendingChangesMutex.RLock()
	defer fake.checkPendingChangesMutex.RUnlock()
	return len(fake.checkPendingChangesArgsForCall)
}

func (fake *FakeOpsManager) CheckPendingChangesCalls(stub func() (bool, error)) {
	fake.checkPendingChangesMutex.Lock()
	defer fake.checkPendingChangesMutex.Unlock()
	fake.CheckPendingChangesStub = stub
}

func (fake *FakeOpsManager) CheckPendingChangesReturns(result1 bool, result2 error) {
	fake.checkPendingChangesMutex.Lock()
	defer fake.checkPendingChangesMutex.Unlock()
	fake.CheckPendingChangesStub = nil
	fake.checkPendingChangesReturns = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManager) CheckPendingChangesReturnsOnCall(i int, result1 bool, result2 error) {
	fake.checkPendingChangesMutex.Lock()
	defer fake.checkPendingChangesMutex.Unlock()
	fake.CheckPendingChangesStub = nil
	if fake.checkPendingChangesReturnsOnCall == nil {
		fake.checkPendingChangesReturnsOnCall = make(map[int]struct {
			result1 bool
			result2 error
		})
	}
	fake.checkPendingChangesReturnsOnCall[i] = struct {
		result1 bool
		result2 error
	}{result1, result2}
}

func (fake *FakeOpsManager) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.applyChangesMutex.RLock()
	defer fake.applyChangesMutex.RUnlock()
	fake.checkPendingChangesMutex.RLock()
	defer fake.checkPendingChangesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeOpsManager) recordInvocation(key string, args []interface{}) {
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

var _ rotate.OpsManager = new(FakeOpsManager)