// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package validate_test

import (
	"errors"
	"io/ioutil"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/bosh"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/credhub"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/manifest"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/validate"

	"testing"
)

type routerBoshRunner struct {
	routerConfig string
}

type routerCredhubRunner struct {
	cert *credhub.Certificate
}

func (b routerBoshRunner) GetDeploymentVMs(deploymentName string) (vms []bosh.VM, err error) {
	return []bosh.VM{
		{
			Name:           "diego_cell/guid1",
			DeploymentName: "cf-guid",
		},
		{
			Name:           "router/guid1",
			DeploymentName: "cf-guid",
		},
	}, nil
}

func (b routerBoshRunner) ScpFile(deploymentName, source, target string) error {
	// pretend to grab the router config
	return ioutil.WriteFile(target, []byte(b.routerConfig), 0644)
}

func (c routerCredhubRunner) GetCertificate(certPath string) (*credhub.Certificate, error) {
	return c.cert, nil
}

func TestValidateRouterCACerts(t *testing.T) {
	newCert := func() *credhub.Certificate {
		c := &credhub.Certificate{Name: "/cf/ca"}
		c.Value.Certificate = "cert"
		c.Value.CA = "cacert"
		c.Value.PrivateKey = "key"
		return c
	}
	credhubRunner := routerCredhubRunner{
		cert: newCert(),
	}
	boshRunner := routerBoshRunner{
		routerConfig: "ca_certs: cacert",
	}

	manifest := &manifest.Manifest{
		DeploymentName: "cf-guid",
		Path:           "/tmp/cf.yml",
	}

	t.Run("matching certs", func(t *testing.T) {
		v := validate.NewRouter(boshRunner, credhubRunner)
		err := v.ValidateCerts(manifest, validate.AllInstancesFilter)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("mismatched certs", func(t *testing.T) {
		boshRunner.routerConfig = "ca_certs: some_other_cert"
		v := validate.NewRouter(boshRunner, credhubRunner)
		err := v.ValidateCerts(manifest, validate.AllInstancesFilter)
		if !errors.Is(err, validate.CertMismatchError) {
			t.Fatal("Expected an error to be returned about certs not matching, error was", err)
		}
	})
}
