// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package manifest

import (
	"fmt"
	"io/ioutil"
)

type BoshExecutor interface {
	GetDiegoDeployments() (deployments []string, err error)
}

type OpsManExecutor interface {
	GetBoshDirectorName() (string, error)
	GetBoshManifest(deploymentName string) ([]byte, error)
}

type Loader struct {
	bosh BoshExecutor
	om   OpsManExecutor
}

func NewLoader(om OpsManExecutor, b BoshExecutor) *Loader {
	return &Loader{
		bosh: b,
		om:   om,
	}
}

func (l *Loader) GetAllManifestsWithDiegoCells() (manifests []Manifest, err error) {
	deployments, err := l.bosh.GetDiegoDeployments()
	if err != nil {
		return nil, fmt.Errorf("could not get diego deployments from bosh: %w", err)
	}

	for _, d := range deployments {
		m, err := l.newManifestFromDeployment(d)
		if err != nil {
			return nil, err
		}
		manifests = append(manifests, *m)
	}
	return manifests, nil
}

func (l *Loader) newManifestFromDeployment(deploymentName string) (*Manifest, error) {
	mbytes, err := l.om.GetBoshManifest(deploymentName)
	if err != nil {
		return nil, fmt.Errorf("could not get bosh manifest for deployment %s: %w",
			deploymentName, err)
	}

	directorName, err := l.om.GetBoshDirectorName()
	if err != nil {
		return nil, fmt.Errorf("could not get bosh director name: %w", err)
	}

	manifestFile, err := l.writeManifestToTempFile(deploymentName, mbytes)
	if err != nil {
		return nil, err
	}

	m, err := NewManifest(directorName, manifestFile)
	if err != nil {
		return nil, fmt.Errorf("could not create manifest instance from manifest file %s: %w",
			manifestFile, err)
	}

	if m.DeploymentName != deploymentName {
		return nil, fmt.Errorf("the manifest's deployment name didn't match, expected %s but got %s",
			deploymentName, m.DeploymentName)
	}

	return m, nil
}

func (l *Loader) writeManifestToTempFile(deploymentName string, manifest []byte) (string, error) {
	mfile, err := ioutil.TempFile("", fmt.Sprintf("%s-*.yml", deploymentName))
	if err != nil {
		return "", fmt.Errorf("could not create bosh manifest temp file for deployment %s: %w",
			deploymentName, err)
	}

	_, err = mfile.Write(manifest)
	if err != nil {
		return "", fmt.Errorf("could not write bosh manifest to %s: %w",
			mfile.Name(), err)
	}

	err = mfile.Close()
	if err != nil {
		return "", fmt.Errorf("could not close manifest file %s: %w",
			mfile.Name(), err)
	}

	return mfile.Name(), nil
}
