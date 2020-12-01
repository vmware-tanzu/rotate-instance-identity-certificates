// Code generated by counterfeiter. DO NOT EDIT.
package rotatefakes

import (
	"sync"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/bosh"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/manifest"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/rotate"
)

type FakeDiegoValidator struct {
	ValidateCertsStub        func(*manifest.Manifest, func(bosh.VM) bool) error
	validateCertsMutex       sync.RWMutex
	validateCertsArgsForCall []struct {
		arg1 *manifest.Manifest
		arg2 func(bosh.VM) bool
	}
	validateCertsReturns struct {
		result1 error
	}
	validateCertsReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDiegoValidator) ValidateCerts(arg1 *manifest.Manifest, arg2 func(bosh.VM) bool) error {
	fake.validateCertsMutex.Lock()
	ret, specificReturn := fake.validateCertsReturnsOnCall[len(fake.validateCertsArgsForCall)]
	fake.validateCertsArgsForCall = append(fake.validateCertsArgsForCall, struct {
		arg1 *manifest.Manifest
		arg2 func(bosh.VM) bool
	}{arg1, arg2})
	stub := fake.ValidateCertsStub
	fakeReturns := fake.validateCertsReturns
	fake.recordInvocation("ValidateCerts", []interface{}{arg1, arg2})
	fake.validateCertsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeDiegoValidator) ValidateCertsCallCount() int {
	fake.validateCertsMutex.RLock()
	defer fake.validateCertsMutex.RUnlock()
	return len(fake.validateCertsArgsForCall)
}

func (fake *FakeDiegoValidator) ValidateCertsCalls(stub func(*manifest.Manifest, func(bosh.VM) bool) error) {
	fake.validateCertsMutex.Lock()
	defer fake.validateCertsMutex.Unlock()
	fake.ValidateCertsStub = stub
}

func (fake *FakeDiegoValidator) ValidateCertsArgsForCall(i int) (*manifest.Manifest, func(bosh.VM) bool) {
	fake.validateCertsMutex.RLock()
	defer fake.validateCertsMutex.RUnlock()
	argsForCall := fake.validateCertsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeDiegoValidator) ValidateCertsReturns(result1 error) {
	fake.validateCertsMutex.Lock()
	defer fake.validateCertsMutex.Unlock()
	fake.ValidateCertsStub = nil
	fake.validateCertsReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeDiegoValidator) ValidateCertsReturnsOnCall(i int, result1 error) {
	fake.validateCertsMutex.Lock()
	defer fake.validateCertsMutex.Unlock()
	fake.ValidateCertsStub = nil
	if fake.validateCertsReturnsOnCall == nil {
		fake.validateCertsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.validateCertsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeDiegoValidator) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.validateCertsMutex.RLock()
	defer fake.validateCertsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeDiegoValidator) recordInvocation(key string, args []interface{}) {
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

var _ rotate.DiegoValidator = new(FakeDiegoValidator)