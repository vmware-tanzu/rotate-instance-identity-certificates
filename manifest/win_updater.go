// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"strings"

	"github.com/mitchellh/pointerstructure"
)

type WinUpdater struct {
	manifest *Manifest
}

func NewWinUpdater(manifest *Manifest) Updater {
	return &WinUpdater{
		manifest: manifest,
	}
}

func (u *WinUpdater) opsmanProductName() string {
	return "pas-windows"
}

func (u *WinUpdater) useNewIntermediateCert() error {
	err := u.manifest.addIntermediateCertRegenVariable()
	if err != nil {
		return err
	}
	repProps := u.manifest.properties(u.instanceGroupName(), "rep_windows")
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
func (u *WinUpdater) useNewRootCert() error {
	repProperties := u.manifest.properties(u.instanceGroupName(), "rep_windows")
	if err := addRootCertRegen(repProperties, "/containers/trusted_ca_certificates"); err != nil {
		return err
	}

	cflinuxfs2Properties := u.manifest.properties(u.instanceGroupName(), "windows1803fs")
	if cflinuxfs2Properties != nil {
		if err := addRootCertRegen(cflinuxfs2Properties, "/windows-rootfs/trusted_certs"); err != nil {
			return err
		}
	}

	return nil
}

func (u *WinUpdater) instanceGroupName() string {
	return "windows_diego_cell" + u.optionalIsoSuffix()
}

func (u *WinUpdater) optionalIsoSuffix() string {
	isoName := strings.TrimPrefix(u.manifest.DeploymentName, u.opsmanProductName()+"-")
	si := strings.LastIndex(isoName, "-")
	if si > -1 {
		return "_" + replaceInvalidNameChars(isoName[:si])
	}
	return ""
}
