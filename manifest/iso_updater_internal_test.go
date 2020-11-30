// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package manifest

import (
	"testing"
)

func TestIsoInstanceGroupNaming(t *testing.T) {
	type test struct {
		instanceGroupName         string
		deploymentName            string
		expectedInstanceGroupName string
	}

	tests := []test{
		{
			instanceGroupName:         "isolated_router",
			deploymentName:            "p-isolation-segment-iso1-pub-065aba009c17a59d5cc9",
			expectedInstanceGroupName: "isolated_router_iso1_pub",
		},
		{
			instanceGroupName:         "isolated_router",
			deploymentName:            "p-isolation-segment-iso1_pub-065aba009c17a59d5cc9",
			expectedInstanceGroupName: "isolated_router_iso1_pub",
		},
		{
			instanceGroupName:         "isolated_router",
			deploymentName:            "p-isolation-segment-iso1 pub-065aba009c17a59d5cc9",
			expectedInstanceGroupName: "isolated_router_iso1_pub",
		},
		{
			instanceGroupName:         "diego_cell",
			deploymentName:            "p-isolation-segment-iso1 pub-065aba009c17a59d5cc9",
			expectedInstanceGroupName: "diego_cell_iso1_pub",
		},
	}

	for _, tc := range tests {
		m, err := NewManifest("p-bosh", "testdata/p-isolation-segment-manifest.yml")
		if err != nil {
			t.Fatal(err)
		}

		u := &IsoUpdater{
			manifest: m,
		}
		u.manifest.DeploymentName = tc.deploymentName

		ig := u.instanceGroupName(tc.instanceGroupName)
		if ig != tc.expectedInstanceGroupName {
			t.Fatalf("Expected instance group name to match %s, but got %s", tc.expectedInstanceGroupName, ig)
		}
	}
}
