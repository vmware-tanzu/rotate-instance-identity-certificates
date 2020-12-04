// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/credhub"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/manifest"
)

// CertMismatchError is the error returned by validators when the certificate
// on the BOSH instance does not match the expected certificate in Credhub.
var CertMismatchError = errors.New("certificate on instance doesn't match credhub value")

// CertExpiration validates credhub cert expiration
type CertExpiration struct {
	credhub *credhub.Runner
}

// NewCertExpiration creates a cert expiration validator
func NewCertExpiration(credhub *credhub.Runner) *CertExpiration {
	return &CertExpiration{
		credhub: credhub,
	}
}

// CheckCertExpiration gets the root CA and intermediate expiration time
func (v *CertExpiration) CheckRootCertExpiration() (expiration time.Time, err error) {
	rootCred, err := v.credhub.GetCertificate(manifest.RootCertName)
	if err != nil {
		return expiration, fmt.Errorf("failed to retreive Root CA certificate from credhub: %w", err)
	}

	expiration, err = getExpiration(rootCred.Value.Certificate)
	if err != nil {
		return expiration, fmt.Errorf("failed to compute root expiration: %w", err)
	}

	return expiration, nil
}

// CheckCertExpiration gets the root CA and intermediate expiration time
func (v *CertExpiration) CheckIntermediateCertExpiration(cfManifest *manifest.Manifest) (expiration time.Time, err error) {
	intermediateCertPath := cfManifest.IntermediateCertPath()
	intermediateCred, err := v.credhub.GetCertificate(intermediateCertPath)
	if err != nil {
		return expiration, fmt.Errorf("failed to retreive intermediate certificate from credhub: %w", err)
	}

	expiration, err = getExpiration(intermediateCred.Value.Certificate)
	if err != nil {
		return expiration, fmt.Errorf("failed to compute intermediate expiration: %w", err)
	}

	return expiration, nil
}

func getExpiration(cred string) (time.Time, error) {
	block, _ := pem.Decode([]byte(cred))
	if block == nil {
		return time.Time{}, errors.New("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert.NotAfter, nil
}
