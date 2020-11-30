// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package credhub

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

type Runner struct {
	env         []string
	credhubExec func(args ...string) ([]byte, error)
}

// NewRunner creates a new credhub runner instance, env should contain the
// environment variables necessary to connect to the credhub instance.
func NewRunner(env []string) *Runner {
	r := &Runner{
		env: env,
	}
	r.credhubExec = r.credhubCLI
	return r
}

func (r *Runner) GetCertificate(certPath string) (*Certificate, error) {
	output, err := r.credhubExec("get", "-n", certPath)
	if err != nil {
		return nil, fmt.Errorf("credhub get failed for %q: %w", certPath, err)
	}

	var cred Certificate
	err = yaml.Unmarshal(output, &cred)
	if err != nil {
		return nil, fmt.Errorf("could not parse certificate from %q: %w", certPath, err)
	}

	return &cred, nil
}

func (r *Runner) Delete(path string) error {
	_, err := r.credhubExec("delete", "-n", path)
	if err != nil {
		return fmt.Errorf("credhub delete failed %q: %w", path, err)
	}
	return nil
}

// ImportCertificates will bulk import multiple certs using a credentials file
func (r *Runner) ImportCertificates(certs []Certificate) error {
	type certificateImport struct {
		Credentials []Certificate `yaml:"credentials"`
	}
	certImport := certificateImport{
		Credentials: certs,
	}

	f, err := ioutil.TempFile("", "credhub-import-*.yml")
	if err != nil {
		return err
	}

	if err := yaml.NewEncoder(f).Encode(certImport); err != nil {
		f.Close()
		os.RemoveAll(f.Name())
		return err
	}

	f.Close()
	defer os.RemoveAll(f.Name())
	return r.Import(f.Name())
}

func (r *Runner) Import(credentialJsonPath string) error {
	_, err := r.credhubExec("import", "-f", credentialJsonPath)
	if err != nil {
		return fmt.Errorf("could not overwrite values in credhub: %w", err)
	}
	return nil
}

func (r *Runner) credhubCLI(args ...string) ([]byte, error) {
	cmd := exec.Command("credhub", args...)
	cmd.Env = r.env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("could not execute credhub command: credhub %s: %w\n%s",
			strings.Join(args, " "), err, string(output))
	}
	return output, nil
}
