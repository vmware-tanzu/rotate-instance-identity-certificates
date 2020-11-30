// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package validate

import "github.com/vmware-tanzu/rotate-instance-identity-certificates/bosh"

var AllInstancesFilter = func(vm bosh.VM) bool { return true }

func FirstInstanceFilter() func(vm bosh.VM) bool {
	count := 0
	return func(vm bosh.VM) bool {
		count++
		return count == 1
	}
}
