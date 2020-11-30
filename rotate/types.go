// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package rotate

import (
	"io"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/bosh"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/credhub"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/manifest"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . BoshRunner
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . DiegoValidator
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . RouterValidator
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CredhubRunner
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . OpsManager
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ManifestLoader

// BoshRunner interfaces with bosh
type BoshRunner interface {
	GetDeploymentVMs(deploymentName string) (vms []bosh.VM, err error)
	Deploy(deploymentName string, manifestFilename string) error
	DeployWithFlags(deploymentName string, manifestFilename string, flags ...string) error
}

// DiegoValidator validates the identity certs on diego cells
type DiegoValidator interface {
	ValidateCerts(manifest *manifest.Manifest, diegoCellFilter func(bosh.VM) bool) error
}

// RouterValidator validates the CA certs on the routers
type RouterValidator interface {
	ValidateCerts(manifest *manifest.Manifest, routerFilter func(bosh.VM) bool) error
}

// CredhubRunner interfaces with credhub
type CredhubRunner interface {
	GetCertificate(certPath string) (*credhub.Certificate, error)
	Delete(path string) error
	Import(credentialJsonPath string) error
	ImportCertificates(certs []credhub.Certificate) error
}

// OpsManager interfaces with opsman
type OpsManager interface {
	CheckPendingChanges() (bool, error)
	ApplyChanges(stdout io.Writer, ignoreWarnings bool, product ...string) error
}

// ManifestLoader loads bosh manifests from existing deployments
type ManifestLoader interface {
	GetAllManifestsWithDiegoCells() (manifests []manifest.Manifest, err error)
}
