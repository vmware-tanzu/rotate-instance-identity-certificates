// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package manifest

import (
	"testing"
)

func TestWinInstanceGroupNaming(t *testing.T) {
	type test struct {
		deploymentName            string
		expectedInstanceGroupName string
	}

	tests := []test{
		{
			deploymentName:            "pas-windows-065aba009c17a59d5cc9",
			expectedInstanceGroupName: "windows_diego_cell",
		},
		{
			deploymentName:            "pas-windows-paswin-pub-065aba009c17a59d5cc9",
			expectedInstanceGroupName: "windows_diego_cell_paswin_pub",
		},
		{
			deploymentName:            "pas-windows-paswin_pub-065aba009c17a59d5cc9",
			expectedInstanceGroupName: "windows_diego_cell_paswin_pub",
		},
		{
			deploymentName:            "pas-windows-paswin pub-065aba009c17a59d5cc9",
			expectedInstanceGroupName: "windows_diego_cell_paswin_pub",
		},
		{
			deploymentName:            "pas-windows-_paswin_pub-065aba009c17a59d5cc9",
			expectedInstanceGroupName: "windows_diego_cell__paswin_pub",
		},
	}

	for _, tc := range tests {
		m, err := NewManifest("p-bosh", "testdata/pas-windows-replicated-manifest.yml")
		if err != nil {
			t.Fatal(err)
		}

		u := &WinUpdater{
			manifest: m,
		}
		u.manifest.DeploymentName = tc.deploymentName

		ig := u.instanceGroupName()
		if ig != tc.expectedInstanceGroupName {
			t.Fatalf("Expected Windows diego cell instance group name to match %s, but got %s", tc.expectedInstanceGroupName, ig)
		}
	}
}
