// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package manifest_test

import (
	"testing"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/manifest"
)

func TestOpsManProductName(t *testing.T) {
	tests := []struct {
		manifest    string
		productName string
	}{
		{"testdata/cf-manifest.yml", "cf"},
		{"testdata/p-isolation-segment-manifest.yml", "p-isolation-segment"},
		{"testdata/pas-windows-manifest.yml", "pas-windows"},
	}

	for _, tc := range tests {
		m, err := manifest.NewManifest("p-bosh", tc.manifest)
		if err != nil {
			t.Fatal(err)
		}

		n := m.OpsManProductName()
		if n != tc.productName {
			t.Errorf("Expected opsman product name %s, but got %s", tc.productName, n)
		}
	}
}
