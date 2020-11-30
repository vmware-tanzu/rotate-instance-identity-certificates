// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package validate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/bosh"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/manifest"
)

var diegoCellPrefixes = []string{"diego_cell", "windows_diego_cell", "isolated_diego_cell"}

// Diego can be used to validate instance identity certs on diego cells
type Diego struct {
	bosh    BoshRunner
	credhub CredhubRunner
}

// NewDiego creates a diego cell cert validator
func NewDiego(bosh BoshRunner, credhub CredhubRunner) *Diego {
	return &Diego{
		credhub: credhub,
		bosh:    bosh,
	}
}

// ValidateCerts checks that each diego cell instance identity cert
// matches the current active intermediate issuing CA.
func (v *Diego) ValidateCerts(manifest *manifest.Manifest, diegoCellFilter func(bosh.VM) bool) error {
	intermediate, err := v.credhub.GetCertificate(manifest.IntermediateCertPath())
	if err != nil {
		return err
	}

	log.Printf("Validating %s diego cell certificates\n", manifest.DeploymentName)
	err = v.checkDiegoCerts(manifest.DeploymentName, intermediate.Value.Certificate, diegoCellFilter)
	if err != nil {
		return fmt.Errorf("validation certs on diego cells: %w", err)
	}
	return nil
}

func (v *Diego) checkDiegoCerts(deploymentName string, credhubIntermediateCert string, diegoCellFilter func(bosh.VM) bool) error {
	vms, err := v.bosh.GetDeploymentVMs(deploymentName)
	if err != nil {
		return err
	}

	for _, vm := range vms {
		if isDiegoCell(vm) && diegoCellFilter(vm) {
			err = v.checkDiegoCert(vm, credhubIntermediateCert)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Diego) checkDiegoCert(diegoCell bosh.VM, credhubIntermediateCert string) error {
	log.Println("Validating cert on", diegoCell.Name)

	instanceIdentityCert, err := ioutil.TempFile("", "instance-identity-*.crt")
	if err != nil {
		return fmt.Errorf("failed to create temp instance identity cert file: %w", err)
	}
	defer os.Remove(instanceIdentityCert.Name())

	source := fmt.Sprintf("%s:%s", diegoCell.Name, identityCertPath(diegoCell))
	err = v.bosh.ScpFile(diegoCell.DeploymentName, source, instanceIdentityCert.Name())
	if err != nil {
		return err
	}

	instanceIdentity, err := ioutil.ReadFile(instanceIdentityCert.Name())
	if err != nil {
		return fmt.Errorf("could not read %s instance identity certificate from %s: %w",
			diegoCell, instanceIdentityCert.Name(), err)
	}

	cic := strings.TrimSpace(credhubIntermediateCert)
	dii := strings.TrimSpace(string(instanceIdentity))

	if cic != dii {
		return fmt.Errorf("%w: for instance %s, expected:\n%s\nbut got:\n%s",
			CertMismatchError, diegoCell, cic, dii)
	}
	return nil
}

func identityCertPath(diegoCell bosh.VM) string {
	if strings.HasPrefix(diegoCell.Name, "windows") {
		return "/var/vcap/jobs/rep_windows/config/certs/rep/instance_identity.crt"
	}
	return "/var/vcap/jobs/rep/config/certs/rep/instance_identity.crt"
}

func isDiegoCell(vm bosh.VM) bool {
	for _, p := range diegoCellPrefixes {
		if strings.HasPrefix(vm.Name, p) {
			return true
		}
	}
	return false
}
