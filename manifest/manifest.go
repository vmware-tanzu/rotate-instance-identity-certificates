// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

// Package manifest performs the necessary operations on the Cloud Foundry
// manifest in order to rotate the Diego instance identity certificate.
package manifest

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/mitchellh/pointerstructure"
	"gopkg.in/yaml.v2"
)

const (
	IntermediateCertName                = "diego-instance-identity-intermediate-ca-2018"
	IntermediateCertRegenName           = IntermediateCertName + "-riic-regen"
	IntermediateCertRegenVariable       = "((" + IntermediateCertRegenName + ".certificate))"
	IntermediatePrivateKeyRegenVariable = "((" + IntermediateCertRegenName + ".private_key))"

	RootCertName          = "/cf/diego-instance-identity-root-ca"
	RootCertRegenName     = RootCertName + "-riic-regen"
	RootCertRegenVariable = "((" + RootCertRegenName + ".certificate))"
)

type Manifest struct {
	DirectorName   string
	DeploymentName string
	Path           string
	Content        map[string]interface{}
	updater        Updater
}

// NewManifest creates and deserializes the manifest from the specified file
func NewManifest(directorName, path string) (*Manifest, error) {
	// Read the source unmodified manifest into a struct
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not read bosh manifest from %s: %v", path, err)
	}
	defer f.Close()

	m := &Manifest{
		Path:         path,
		DirectorName: directorName,
	}
	if err := yaml.NewDecoder(f).Decode(&m.Content); err != nil {
		return nil, fmt.Errorf("could not deserialize bosh manifest %s: %v", path, err)
	}
	m.DeploymentName = m.Content["name"].(string)

	if strings.HasPrefix(m.DeploymentName, "cf-") {
		m.updater = NewCFUpdater(m)
	} else if strings.HasPrefix(m.DeploymentName, "p-isolation-segment") {
		m.updater = NewIsoUpdater(m)
	} else if strings.HasPrefix(m.DeploymentName, "pas-windows-") {
		m.updater = NewWinUpdater(m)
	} else {
		return nil, fmt.Errorf("unknown manifest deployment type %s", m.DeploymentName)
	}

	return m, nil
}

// Update performs modifications to the source manifest that can be used to
// rotate the instance identity certificates. The rotation procedure adds a new
// root CA, and anew intermedaite CA signed by the new root.
//
// It's important that the certs are added in order of precedence (i.e. CA
// before intermediate) to avoid the bosh deploy error: "Config Server failed
// to generate value".
func (m *Manifest) Update(withNewIntermediate io.Writer) error {
	if err := m.updater.useNewIntermediateCert(); err != nil {
		return err
	}
	if err := m.updater.useNewRootCert(); err != nil {
		return err
	}

	// Serialize the mutated manifest back out to the new file
	if err := yaml.NewEncoder(withNewIntermediate).Encode(m.Content); err != nil {
		return fmt.Errorf("cannot add new intermediate CA: %w", err)
	}

	return nil
}

// OpsManProductName returns the associated Opsman tile product name as is
// returned from 'om available-products'
func (m *Manifest) OpsManProductName() string {
	return m.updater.opsmanProductName()
}

// IntermediateCertPath returns the full credhub deployment specific path to
// the intermediate CA.
func (m *Manifest) IntermediateCertPath() string {
	return m.credhubDeploymentPath(IntermediateCertName)
}

// IntermediateCertRegenPath returns the full credhub deployment specific path
// to the regenerated intermediate CA.
func (m *Manifest) IntermediateCertRegenPath() string {
	return m.credhubDeploymentPath(IntermediateCertRegenName)
}

// credhubDeploymentPath returns the full deployment specific path to the cert.
func (m *Manifest) credhubDeploymentPath(certName string) string {
	return fmt.Sprintf("/%s/%s/%s", m.DirectorName, m.DeploymentName, certName)
}

func (m *Manifest) addIntermediateCertRegenVariable() error {
	// Add a new intermediate signed the new root and make sure the diego cells
	// are using this new intermediate instead of the one about to expire
	intermediate, err := m.cloneVariable(IntermediateCertName, IntermediateCertRegenName)
	if err != nil {
		return err
	}
	if _, err := pointerstructure.Set(intermediate, "/ca", RootCertRegenName); err != nil {
		return fmt.Errorf("could not set .ca on new intermediate cert: %v", err)
	}
	if _, err := pointerstructure.Set(intermediate, "/options/ca", RootCertRegenName); err != nil {
		return fmt.Errorf("could not set .options.ca on new intermediate cert: %v", err)
	}
	return nil
}

func (m *Manifest) addRootCertRegenVariable() error {
	_, err := m.cloneVariable(RootCertName, RootCertRegenName)
	return err
}

// cloneVariable makes a copy of the a variable in the manifest.
// It returns the copy so additional modifications can be made.
func (m *Manifest) cloneVariable(variableName string, newName string) (map[interface{}]interface{}, error) {
	vars, ok := m.Content["variables"]
	if !ok {
		return nil, errors.New("manifest is missing /variables section")
	}

	vl, ok := vars.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expected []interface{} at /variables, but got %T", vars)
	}

	for _, variable := range vl {
		vv, ok := variable.(map[interface{}]interface{})
		if !ok {
			return nil, fmt.Errorf("variable has unexpected type %T", variable)
		}
		n := vv["name"]
		name, ok := n.(string)
		if !ok {
			return nil, errors.New("expected variable to have a string name")
		}
		if name == variableName {
			// make a new variable that is a copy of the existing variable
			copy := make(map[interface{}]interface{})
			for k, v := range vv {
				copy[k] = v
			}
			copy["name"] = newName

			// add the copy to the manifest's list of variables
			m.Content["variables"] = prepend(m.Content["variables"].([]interface{}), copy)

			return copy, nil
		}
	}

	return nil, fmt.Errorf("could not find variable %s in manifest", variableName)
}

func addRootCertRegen(props map[interface{}]interface{}, path string) error {
	val, err := pointerstructure.Get(props, path)
	if err != nil {
		return fmt.Errorf("cannot add cert to %v: %w", path, err)
	}
	switch vs := val.(type) {
	case string:
		v := RootCertRegenVariable + "\n" + vs
		_, err = pointerstructure.Set(props, path, v)
	case []interface{}:
		vs = prepend(vs, RootCertRegenVariable)
		_, err = pointerstructure.Set(props, path, vs)
	default:
		return fmt.Errorf("cannot add cert to %v: unexpected type %T", path, val)
	}

	if err != nil {
		return fmt.Errorf("cannot add cert to %v: cannot update value: %v", path, err)
	}
	return nil
}

// properties gets the job properties for a particular job
// in the specified instance group
func (m *Manifest) properties(instanceGroup, jobName string) map[interface{}]interface{} {
	igs, ok := m.Content["instance_groups"]
	if !ok {
		return nil
	}

	igl, ok := igs.([]interface{})
	if !ok {
		return nil
	}

	for _, ig := range igl {
		i, ok := ig.(map[interface{}]interface{})
		if !ok {
			continue
		}
		if i["name"] == instanceGroup {
			jobs, ok := i["jobs"]
			if !ok {
				return nil
			}
			jl, ok := jobs.([]interface{})
			if !ok {
				return nil
			}
			for _, job := range jl {
				job, ok := job.(map[interface{}]interface{})
				if !ok {
					return nil
				}
				if job["name"] == jobName {
					result := job["properties"]
					if r, ok := result.(map[interface{}]interface{}); ok {
						return r
					}
				}
			}
		}
	}

	return nil
}

func prepend(x []interface{}, y interface{}) []interface{} {
	x = append(x, 0)
	copy(x[1:], x)
	x[0] = y
	return x
}
