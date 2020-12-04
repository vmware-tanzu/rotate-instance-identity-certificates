// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/credhub"
)

// CredhubRunner interfaces with credhub
type CredhubRunner interface {
	GetCertificate(certPath string) (*credhub.Certificate, error)
}
