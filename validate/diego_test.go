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

type boshRunner struct {
	identityCert string
}

type credhubRunner struct {
	cert *credhub.Certificate
}

func (b boshRunner) GetDeploymentVMs(deploymentName string) (vms []bosh.VM, err error) {
	return []bosh.VM{
		{
			Name:           "diego_cell/guid1",
			DeploymentName: "cf-guid",
		},
		{
			Name:           "diego_cell/guid2",
			DeploymentName: "cf-guid",
		},
		{
			Name:           "router/guid1",
			DeploymentName: "cf-guid",
		},
	}, nil
}

func (b boshRunner) ScpFile(deploymentName, source, target string) error {
	// pretend to grab the cert from the diego cell
	return ioutil.WriteFile(target, []byte(b.identityCert), 0644)
}

func (c credhubRunner) GetCertificate(certPath string) (*credhub.Certificate, error) {
	return c.cert, nil
}

func TestValidateDiegoIdentityCerts(t *testing.T) {
	newCert := func() *credhub.Certificate {
		c := &credhub.Certificate{Name: "/cf/ca"}
		c.Value.Certificate = "cert"
		c.Value.CA = "cert"
		c.Value.PrivateKey = "cert"
		return c
	}
	credhubRunner := credhubRunner{
		cert: newCert(),
	}
	boshRunner := boshRunner{
		identityCert: credhubRunner.cert.Value.Certificate,
	}

	manifest := &manifest.Manifest{
		DeploymentName: "cf-guid",
		Path:           "/tmp/cf.yml",
	}

	t.Run("matching certs", func(t *testing.T) {
		v := validate.NewDiego(boshRunner, credhubRunner)
		err := v.ValidateCerts(manifest, validate.AllInstancesFilter)
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("mismatched certs", func(t *testing.T) {
		boshRunner.identityCert = "some_other_cert"
		v := validate.NewDiego(boshRunner, credhubRunner)
		err := v.ValidateCerts(manifest, validate.AllInstancesFilter)
		if !errors.Is(err, validate.CertMismatchError) {
			t.Fatal("Expected an error to be returned about certs not matching, error was", err)
		}
	})
}
