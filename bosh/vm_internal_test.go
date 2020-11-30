// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package bosh

import (
	"io/ioutil"
	"testing"
)

var expectedVMNames = []string{
	"isolated_diego_cell/88dabc7b-be3e-440e-999c-1c1efe4b6b14",
	"isolated_diego_cell/c3c50ebc-d7a7-443e-9b29-6f7b46ffb6c0",
	"isolated_diego_cell/f767ccac-da5c-4f1c-870d-2e6c82a57f24",
	"isolated_router/47cbcaed-ad6c-4864-bd88-49514b64b3ee",
	"isolated_router/7739f5cf-74c5-43b1-8a96-ea1dc99adc00",
	"isolated_router/afdcc6ad-76c8-4e41-b692-8fbdb4bafc54",
}

func TestLoadVMs(t *testing.T) {
	f, err := ioutil.ReadFile("testdata/iso-vms.json")
	if err != nil {
		t.Fatalf("Failed to read test data iso-vms.json: %s", err)
	}

	vms, err := loadVMs("p-isolation-segment-guid", f)
	if err != nil {
		t.Fatalf("Failed to parse VMs from iso-vms.json: %s", err)
	}

	if len(vms) != 6 {
		t.Fatalf("Expected 6 VMs but got %d", len(vms))
	}
	for _, vm := range vms {
		if vm.DeploymentName != "p-isolation-segment-guid" {
			t.Fatalf("Expected VM to have deployment name p-isolation-segment-guid but got %s", vm.DeploymentName)
		}
		foundVMName := false
		for _, expectedVMName := range expectedVMNames {
			if expectedVMName == vm.Name {
				foundVMName = true
				break
			}
		}
		if !foundVMName {
			t.Fatalf("VM name %s is not in the expected list: %v", vm.Name, expectedVMNames)
		}
	}
}
