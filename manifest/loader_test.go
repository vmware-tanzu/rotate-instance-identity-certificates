// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package manifest_test

import (
	"testing"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/manifest"
)

type boshExecutor struct{}
type omExecutor struct{}

func (b boshExecutor) GetDiegoDeployments() (manifests []string, err error) {
	return []string{"cf-guid"}, nil
}

func (o omExecutor) GetBoshManifest(deploymentName string) ([]byte, error) {
	return []byte("name: cf-guid\nvariables:\n  - name: diego-instance-identity-intermediate-ca-2018"), nil
}

func (o omExecutor) GetBoshDirectorName() (string, error) {
	return "p-bosh", nil
}

func TestNewLoader(t *testing.T) {
	e := boshExecutor{}
	o := omExecutor{}
	l := manifest.NewLoader(o, e)
	if l == nil {
		t.Errorf("couldn't create loader")
	}

	manifests, err := l.GetAllManifestsWithDiegoCells()
	if err != nil {
		t.Fatal(err)
	}
	if len(manifests) != 1 {
		t.Errorf("Expected 1 manifest, but got %d", len(manifests))
	}

	m := manifests[0]
	if m.DeploymentName != "cf-guid" {
		t.Errorf("Expected deployment name cf-guid, but got %s", m.DeploymentName)
	}
}
