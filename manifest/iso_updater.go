// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"strings"

	"github.com/mitchellh/pointerstructure"
)

type IsoUpdater struct {
	manifest *Manifest
}

func NewIsoUpdater(manifest *Manifest) Updater {
	return &IsoUpdater{
		manifest: manifest,
	}
}

func (u *IsoUpdater) opsmanProductName() string {
	return "p-isolation-segment"
}

func (u *IsoUpdater) useNewIntermediateCert() error {
	err := u.manifest.addIntermediateCertRegenVariable()
	if err != nil {
		return err
	}
	repProps := u.manifest.properties(u.instanceGroupName("isolated_diego_cell"), "rep")
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
func (u *IsoUpdater) useNewRootCert() error {
	gorouterProps := u.manifest.properties(u.instanceGroupName("isolated_router"), "gorouter")
	if err := addRootCertRegen(gorouterProps, "/router/ca_certs"); err != nil {
		return err
	}

	repProperties := u.manifest.properties(u.instanceGroupName("isolated_diego_cell"), "rep")
	if err := addRootCertRegen(repProperties, "/containers/trusted_ca_certificates"); err != nil {
		return err
	}

	cflinuxfs2Properties := u.manifest.properties(u.instanceGroupName("isolated_diego_cell"), "cflinuxfs2-rootfs-setup")
	if cflinuxfs2Properties != nil {
		if err := addRootCertRegen(cflinuxfs2Properties, "/cflinuxfs2-rootfs/trusted_certs"); err != nil {
			return err
		}
	}

	cflinuxfs3Properties := u.manifest.properties(u.instanceGroupName("isolated_diego_cell"), "cflinuxfs3-rootfs-setup")
	if cflinuxfs3Properties != nil {
		if err := addRootCertRegen(cflinuxfs3Properties, "/cflinuxfs3-rootfs/trusted_certs"); err != nil {
			return err
		}
	}

	return nil
}

func (u *IsoUpdater) instanceGroupName(prefix string) string {
	return prefix + u.optionalIsoSuffix()
}

func (u *IsoUpdater) optionalIsoSuffix() string {
	isoName := strings.TrimPrefix(u.manifest.DeploymentName, u.opsmanProductName()+"-")
	si := strings.LastIndex(isoName, "-")
	if si > -1 {
		return "_" + replaceInvalidNameChars(isoName[:si])
	}
	return ""
}
