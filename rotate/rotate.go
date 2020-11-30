// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package rotate

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/credhub"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/manifest"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/validate"
)

// CertRotator rotates diego instance identity and associated root CA certs
type CertRotator struct {
	om              OpsManager
	credhub         CredhubRunner
	bosh            BoshRunner
	manifestLoader  ManifestLoader
	diegoValidator  DiegoValidator
	routerValidator RouterValidator
}

// NewCertRotator creates a new CertRotator instance
func NewCertRotator(
	om OpsManager,
	bosh BoshRunner,
	credhub CredhubRunner,
	manifestLoader ManifestLoader,
	diegoValidator DiegoValidator,
	routerValidator RouterValidator) *CertRotator {
	return &CertRotator{
		om:              om,
		credhub:         credhub,
		bosh:            bosh,
		manifestLoader:  manifestLoader,
		diegoValidator:  diegoValidator,
		routerValidator: routerValidator,
	}
}

// RotateCerts rotates all the instance identity certs for all deployements
// with diego cells (TAS, TASW, ISO).
//
// This is a 2 phase deployment process. The first bosh deploys are done
// directly via BOSH to add a new root and intermediate CAs. The final deploy
// is via Operations Manager which removes the temporary regen certs.
func (r *CertRotator) RotateCerts(startStage string) error {
	if err := r.checkPendingChanges(); err != nil {
		return err
	}

	manifests, err := r.getDiegoCellManifestsSorted()
	if err != nil {
		return err
	}

	switch startStage {
	default:
		log.Printf("[WARNING]: unknown start phase %s, starting at beginning", startStage)
		fallthrough
	case "bosh": // start by generating new manfiests and bosh deploying
		if err = r.addRegenCertsToBoshDeployments(manifests); err != nil {
			return err
		}
		fallthrough
	case "credhub": // start with the credhub overwrite and apply changes
		if err = r.rotateCertsInCredhub(manifests); err != nil {
			return err
		}
		err = r.validateCertsWereRotated(manifests)
		if errors.Is(err, validate.CertMismatchError) {
			return err
		}
		if err != nil {
			log.Println("[WARNING]: could not validate certs:", err)
		}
		fallthrough
	case "apply": // start with the apply changes
		if err = r.applyChanges(manifests); err != nil {
			return err
		}

		err = r.validateCertsWereRotated(manifests)
		if errors.Is(err, validate.CertMismatchError) {
			return err
		}
		if err != nil {
			log.Println("[WARNING]: could not validate certs", err)
		}
		fallthrough
	case "cleanup":
		return r.cleanupRegenCerts(manifests)
	}
}

func (r *CertRotator) checkPendingChanges() error {
	log.Println("Checking for pending changes")
	hasChanges, err := r.om.CheckPendingChanges()
	if err != nil {
		return fmt.Errorf("cannot check for pending changes: %w", err)
	}

	if hasChanges {
		return errors.New("cannot continue while there are pending changes")
	}

	return nil
}

// getDiegoCellManifestsSorted returns all bosh manifests that have diego cells
// sorted with CF first, then alphabetical. It's important to modify the CF
// deployment before any optional isolation segments or windows segments.
func (r *CertRotator) getDiegoCellManifestsSorted() ([]manifest.Manifest, error) {
	log.Println("Retrieving BOSH manifests for all Diego deployments")

	manifests, err := r.manifestLoader.GetAllManifestsWithDiegoCells()
	if err != nil {
		return nil, err
	}

	sort.Slice(manifests, func(i, j int) bool {
		if manifests[i].OpsManProductName() == "cf" {
			return true
		}
		return manifests[i].DeploymentName < manifests[j].DeploymentName
	})

	return manifests, nil
}

func (r *CertRotator) addRegenCertsToBoshDeployments(manifests []manifest.Manifest) error {
	for _, m := range manifests {
		err := r.rotateManifestCerts(&m)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CertRotator) rotateCertsInCredhub(manifests []manifest.Manifest) error {
	log.Println("Rotating identity certs in Credhub")

	rootRegenCert, err := r.credhub.GetCertificate(manifest.RootCertRegenName)
	if err != nil {
		return err
	}

	// update the name to point at the original cert location so it's overwritten
	rootRegenCert.Name = manifest.RootCertName

	certsToImport := []credhub.Certificate{*rootRegenCert}

	// add all the intermediate identity certs to credhub
	for _, m := range manifests {
		log.Printf("Updating %s Credhub references to overwrite old certificates", m.DeploymentName)

		intermediateRegenCert, err := r.credhub.GetCertificate(m.IntermediateCertRegenPath())
		if err != nil {
			return err
		}

		intermediateRegenCert.Name = m.IntermediateCertPath()
		certsToImport = append(certsToImport, *intermediateRegenCert)
	}

	err = r.credhub.ImportCertificates(certsToImport)
	if err != nil {
		return fmt.Errorf("could not overwrite values in credhub: %w", err)
	}

	return nil
}

const ignoreWarnings = true

func (r *CertRotator) applyChanges(manifests []manifest.Manifest) (err error) {
	log.Println("Removing temporary regen certificate enties from BOSH deployments")

	for _, m := range manifests {
		log.Printf("Applying changes to %s", m.OpsManProductName())
		if err = r.om.ApplyChanges(os.Stdout, ignoreWarnings, m.OpsManProductName()); err != nil {
			return err
		}
	}

	return nil
}

func (r *CertRotator) validateCertsWereRotated(manifests []manifest.Manifest) error {
	for _, m := range manifests {
		// check the cert only on the first diego cell as a sanity check
		err := r.diegoValidator.ValidateCerts(&m, validate.FirstInstanceFilter())
		if err != nil {
			return err
		}

		// check the first gorouter
		err = r.routerValidator.ValidateCerts(&m, validate.FirstInstanceFilter())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CertRotator) cleanupRegenCerts(manifests []manifest.Manifest) error {
	log.Println("Removing duplicate regen certificates from credhub")
	for _, m := range manifests {
		err := r.credhub.Delete(m.IntermediateCertRegenPath())
		if err != nil {
			return err
		}
	}
	return r.credhub.Delete(manifest.RootCertRegenName)
}

// rotateManifestCerts performs the instance identity certificate rotation on
// the specified deployment's manifest
func (r *CertRotator) rotateManifestCerts(cfManifest *manifest.Manifest) error {
	log.Printf("Creating BOSH manifest with regen certs for %s", cfManifest.DeploymentName)
	withIntermediate, err := ioutil.TempFile("", cfManifest.DeploymentName+"-intermediate-regen-*.yml")
	if err != nil {
		return err
	}

	if err := cfManifest.Update(withIntermediate); err != nil {
		withIntermediate.Close()
		os.RemoveAll(withIntermediate.Name())
		return err
	}
	withIntermediate.Close()

	// add --recreate for TASW, as the cert injector gets stuck otherwise
	var flags []string
	if cfManifest.OpsManProductName() == "pas-windows" {
		flags = append(flags, "--recreate")
	}
	log.Printf("BOSH deploying %s with new identity certs", cfManifest.DeploymentName)
	err = r.bosh.DeployWithFlags(cfManifest.DeploymentName, withIntermediate.Name(), flags...)
	if err != nil {
		return fmt.Errorf("bosh deploy with new identity certs failed: %w", err)
	}

	return nil
}
