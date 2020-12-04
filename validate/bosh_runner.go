// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/bosh"
)

// BoshRunner interfaces with bosh
type BoshRunner interface {
	GetDeploymentVMs(deploymentName string) (vms []bosh.VM, err error)
	ScpFile(deploymentName, source, target string) error
}
