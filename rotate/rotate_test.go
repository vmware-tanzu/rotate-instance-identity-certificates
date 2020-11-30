// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package rotate_test

import (
	"errors"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/credhub"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/manifest"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/rotate"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/rotate/rotatefakes"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/validate"
)

func TestRotate(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	var (
		om   *rotatefakes.FakeOpsManager
		bosh *rotatefakes.FakeBoshRunner
		ch   *rotatefakes.FakeCredhubRunner
		ml   *rotatefakes.FakeManifestLoader
		dv   *rotatefakes.FakeDiegoValidator
		rv   *rotatefakes.FakeRouterValidator

		r *rotate.CertRotator
	)

	setup := func() {
		t.Helper()

		om = &rotatefakes.FakeOpsManager{}
		bosh = &rotatefakes.FakeBoshRunner{}
		ch = &rotatefakes.FakeCredhubRunner{}
		ml = &rotatefakes.FakeManifestLoader{}
		dv = &rotatefakes.FakeDiegoValidator{}
		rv = &rotatefakes.FakeRouterValidator{}

		om.CheckPendingChangesReturns(false, nil)

		cf, err := manifest.NewManifest("p-bosh-12345", "testdata/cf-manifest.yml")
		if err != nil {
			t.Fatal(err)
		}
		ml.GetAllManifestsWithDiegoCellsReturns([]manifest.Manifest{*cf}, nil)

		cert := &credhub.Certificate{
			Name: "/p-bosh-12345/some-cert",
			Type: "certificate",
		}
		cert.Value.Certificate = "---SOME CERTIFICATE---"
		cert.Value.CA = "---SOME CA---"
		cert.Value.PrivateKey = "--- SOME PRIVATE KEY ---"

		ch.GetCertificateReturns(cert, nil)
		ch.ImportCertificatesReturns(nil)
		ch.DeleteReturns(nil)

		dv.ValidateCertsReturns(nil)
		rv.ValidateCertsReturns(nil)

		om.ApplyChangesReturns(nil)

		r = rotate.NewCertRotator(om, bosh, ch, ml, dv, rv)
	}

	t.Run("refuses to rotate with pending changes", func(t *testing.T) {
		setup()
		om.CheckPendingChangesReturns(true, nil)
		if err := r.RotateCerts("bosh"); err == nil {
			t.Fatal("expected error due to pending changes, but rotation succeeded")
		}
		if count := ml.GetAllManifestsWithDiegoCellsCallCount(); count > 0 {
			t.Errorf("GetAllManifests should not have been called, but received %d calls", count)
		}
	})

	t.Run("full rotate", func(t *testing.T) {
		setup()
		if err := r.RotateCerts("bosh"); err != nil {
			t.Fatal(err)
		}

		if count := bosh.DeployWithFlagsCallCount(); count != 1 {
			t.Errorf("expected 1 bosh deployment, but got %d", count)
		}
		if count := om.ApplyChangesCallCount(); count != 1 {
			t.Errorf("expected 1 apply changes, but got %d", count)
		}

		_, _, products := om.ApplyChangesArgsForCall(0)
		if p := products[0]; p != "cf" {
			t.Errorf("expected selective apply for cf, but apply was for product %v", p)
		}

		if count := ch.DeleteCallCount(); count != 2 {
			t.Errorf("expected 2 credhub delete calls, but got %d", count)
		}
		if count := rv.ValidateCertsCallCount(); count <= 0 {
			t.Errorf("expected router certs to be validated, but got %d calls", count)
		}
		if count := dv.ValidateCertsCallCount(); count <= 0 {
			t.Errorf("expected diego cell certs to be validated, but got %d calls", count)
		}
	})

	t.Run("invalid start phase starts at beginning", func(t *testing.T) {
		setup()
		if err := r.RotateCerts("NOT-A-VALID-START-PHASE"); err != nil {
			t.Fatal(err)
		}
		if count := bosh.DeployWithFlagsCallCount(); count != 1 {
			t.Errorf("expected 1 bosh deployment, but got %d", count)
		}
	})

	t.Run("start phase cleanup", func(t *testing.T) {
		setup()
		if err := r.RotateCerts("cleanup"); err != nil {
			t.Fatal(err)
		}
		// expect we delete the duplicate copies of the root and the intermediate
		if count := ch.DeleteCallCount(); count != 2 {
			t.Errorf("expected 2 credhub delete calls, but got %d", count)
		}
		// ensure we didn't run the apply changes
		if count := om.ApplyChangesCallCount(); count > 0 {
			t.Errorf("apply changes should not have been called, but received %d calls", count)
		}

		// ensure we delete the correct certs
		intermediate := ch.DeleteArgsForCall(0)
		root := ch.DeleteArgsForCall(1)

		if intermediate != "/p-bosh-12345/cf-a7e7cd52009e7c121d7e/diego-instance-identity-intermediate-ca-2018-riic-regen" {
			t.Errorf("deleted intermediate at %v from credhub, expected /p-bosh-12345/cf-a7e7cd52009e7c121d7e/diego-instance-identity-intermediate-ca-2018-riic-regen", intermediate)
		}
		if root != "/cf/diego-instance-identity-root-ca-riic-regen" {
			t.Errorf("deleted root at %v from credhub, expected /cf/diego-instance-identity-root-ca-riic-regen", root)
		}
	})

	t.Run("validate fails with cert mismatch", func(t *testing.T) {
		setup()
		dv.ValidateCertsReturns(validate.CertMismatchError)
		if err := r.RotateCerts("bosh"); err == nil {
			t.Fatal("expected error due to cert mismatch, but operation succeeded")
		}
	})

	t.Run("validate fails with unknown error, rotation continues", func(t *testing.T) {
		setup()
		dv.ValidateCertsReturns(errors.New("unexpected error"))
		if err := r.RotateCerts("bosh"); err != nil {
			t.Fatal("expected rotation to complete despite validation failure, got error", err)
		}
		if count := om.ApplyChangesCallCount(); count != 1 {
			t.Errorf("expected 1 apply changes, but got %d", count)
		}
		if count := ch.DeleteCallCount(); count != 2 {
			t.Errorf("expected 2 credhub delete calls, but got %d", count)
		}
	})

	t.Run("windows uses --recreate", func(t *testing.T) {
		setup()

		win, err := manifest.NewManifest("p-bosh-12345", "testdata/pas-windows-manifest.yml")
		if err != nil {
			t.Fatal(err)
		}
		ml.GetAllManifestsWithDiegoCellsReturns([]manifest.Manifest{*win}, nil)

		if err := r.RotateCerts("bosh"); err != nil {
			t.Fatal("expected rotation to complete despite validation failure, got error", err)
		}

		_, _, args := bosh.DeployWithFlagsArgsForCall(0)
		foundRecreate := false
		for _, arg := range args {
			if arg == "--recreate" {
				foundRecreate = true
				break
			}
		}
		if !foundRecreate {
			t.Errorf("expected the Windows deployment to include --recreate, args: %v",
				strings.Join(args, " "))
		}

	})
}
