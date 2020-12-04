// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"fmt"

	"github.com/mitchellh/pointerstructure"
)

type CFUpdater struct {
	manifest *Manifest
}

func NewCFUpdater(manifest *Manifest) Updater {
	return &CFUpdater{
		manifest: manifest,
	}
}

func (u *CFUpdater) opsmanProductName() string {
	return "cf"
}

func (u *CFUpdater) useNewIntermediateCert() error {
	err := u.manifest.addIntermediateCertRegenVariable()
	if err != nil {
		return err
	}
	repProps := u.manifest.properties("diego_cell", "rep")
	if _, err := pointerstructure.Set(repProps, "/diego/executor/instance_identity_ca_cert", IntermediateCertRegenVariable); err != nil {
		return err
	}
	if _, err := pointerstructure.Set(repProps, "/diego/executor/instance_identity_key", IntermediatePrivateKeyRegenVariable); err != nil {
		return err
	}
	return nil
}

// trustNewRoot ensures that the new root certificate variable is a trusted
// CA for each of the required jobs in the manifest
func (u *CFUpdater) useNewRootCert() error {
	err := u.manifest.addRootCertRegenVariable()
	if err != nil {
		return err
	}

	gorouterProps := u.manifest.properties("router", "gorouter")
	if err := addRootCertRegen(gorouterProps, "/router/ca_certs"); err != nil {
		return err
	}

	credhubProps := u.manifest.properties("credhub", "credhub")
	if err := addRootCertRegen(credhubProps, "/credhub/authentication/mutual_tls/trusted_cas"); err != nil {
		return err
	}

	repProperties := u.manifest.properties("diego_cell", "rep")
	if err := addRootCertRegen(repProperties, "/containers/trusted_ca_certificates"); err != nil {
		return err
	}

	cflinuxfs2Properties := u.manifest.properties("diego_cell", "cflinuxfs2-rootfs-setup")
	if cflinuxfs2Properties != nil {
		if err := addRootCertRegen(cflinuxfs2Properties, "/cflinuxfs2-rootfs/trusted_certs"); err != nil {
			return err
		}
	}

	cflinuxfs3Properties := u.manifest.properties("diego_cell", "cflinuxfs3-rootfs-setup")
	if cflinuxfs3Properties != nil {
		if err := addRootCertRegen(cflinuxfs3Properties, "/cflinuxfs3-rootfs/trusted_certs"); err != nil {
			return err
		}
	}

	sshProxyProperties := u.manifest.properties("diego_brain", "ssh_proxy")
	if err := addRootCertRegen(sshProxyProperties, "/diego/ssh_proxy/bbs/ca_cert"); err != nil {
		return err
	}

	backendTLS, err := pointerstructure.Get(sshProxyProperties, "/backends/tls/enabled")
	if err != nil {
		return fmt.Errorf("couldn't check whether SSH proxy backend TLS is enabled: %w", err)
	}

	if enabled, ok := backendTLS.(bool); ok && enabled {
		if err := addRootCertRegen(sshProxyProperties, "/backends/tls/ca_certificates"); err != nil {
			return err
		}
	}

	return nil
}
