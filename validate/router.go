// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/bosh"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/manifest"
	"gopkg.in/yaml.v2"
)

const routerVMConfigPath = "/var/vcap/jobs/gorouter/config/gorouter.yml"

// DiegoValidator can be used to validate instance identity certs on diego cells
type Router struct {
	bosh    BoshRunner
	credhub CredhubRunner
}

// NewDiegoValidator creates a diego cell cert validator
func NewRouter(bosh BoshRunner, credhub CredhubRunner) *Router {
	return &Router{
		credhub: credhub,
		bosh:    bosh,
	}
}

// ValidateCerts checks that each diego cell instance identity cert
// matches the current active intermediate issuing CA.
func (v *Router) ValidateCerts(manifest *manifest.Manifest, routerVMFilter func(bosh.VM) bool) error {
	intermediate, err := v.credhub.GetCertificate(manifest.IntermediateCertPath())
	if err != nil {
		return err
	}

	log.Printf("Validating %s router certificates\n", manifest.DeploymentName)
	err = v.checkRouterCerts(manifest.DeploymentName, strings.TrimSpace(intermediate.Value.CA), routerVMFilter)
	if err != nil {
		return fmt.Errorf("validating certs on routers: %w", err)
	}
	return nil
}

func (v *Router) checkRouterCerts(deploymentName string, caCert string, routerVMFilter func(bosh.VM) bool) error {
	vms, err := v.bosh.GetDeploymentVMs(deploymentName)
	if err != nil {
		return err
	}

	for _, vm := range vms {
		if isRouter(vm) && routerVMFilter(vm) {
			err = v.checkRouterCert(vm, caCert)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Router) checkRouterCert(routerVM bosh.VM, caCertFromCredhub string) error {
	log.Println("Validating cert on", routerVM.Name)

	routerConfig, err := ioutil.TempFile("", "router-config-*.yml")
	if err != nil {
		return fmt.Errorf("failed to create temp instance router config file: %w", err)
	}
	defer os.Remove(routerConfig.Name())

	source := fmt.Sprintf("%s:%s", routerVM.Name, routerVMConfigPath)
	err = v.bosh.ScpFile(routerVM.DeploymentName, source, routerConfig.Name())
	if err != nil {
		return fmt.Errorf("failed to scp router config from %s to %s: %w",
			routerVM, routerConfig.Name(), err)
	}

	routerConfigContent, err := ioutil.ReadAll(routerConfig)
	if err != nil {
		return fmt.Errorf("failed to read router config from %s: %w",
			routerConfig.Name(), err)
	}

	type cacerts struct {
		CACerts string `yaml:"ca_certs"`
	}

	var c cacerts
	err = yaml.Unmarshal(routerConfigContent, &c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal router yaml from %s: %w",
			routerConfig.Name(), err)
	}

	if !strings.Contains(c.CACerts, caCertFromCredhub) {
		return fmt.Errorf("%w: for instance %s, expected it to contain CA:\n%s\nbut got:\n%s",
			CertMismatchError, routerVM, caCertFromCredhub, c.CACerts)
	}
	return nil
}

func isRouter(vm bosh.VM) bool {
	return strings.HasPrefix(vm.Name, "router")
}
