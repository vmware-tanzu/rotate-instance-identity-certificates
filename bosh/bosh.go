// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package bosh

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// diegoDeploymentPrefixes contains all the known TAS bosh deployment string prefixes
var diegoDeploymentPrefixes = []string{"cf-", "p-isolation-segment-", "pas-windows-"}

// Runner is a bosh command runner
type Runner struct {
	env []string
}

// NewRunner creates a new bosh runner instance, env should contain the
// environment variables necessary to connect to the bosh instance.
func NewRunner(env []string) *Runner {
	return &Runner{
		env: env,
	}
}

// Deploy executes the specified bosh deployment with the specified
// manifest on disk.
func (r Runner) Deploy(deploymentName string, manifestFilename string) error {
	return r.DeployWithFlags(deploymentName, manifestFilename)
}

func (r Runner) DeployWithFlags(deploymentName string, manifestFilename string, flags ...string) error {
	args := []string{
		"deploy",
		manifestFilename,
		"--deployment", deploymentName,
		"--non-interactive",
	}

	args = append(args, flags...)

	log.Println("running command: bosh", strings.Join(args, " "))
	cmd := exec.Command("bosh", args...)
	cmd.Env = r.env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// GetDeploymentManifest returns the bosh deployment manifest yaml for
// the specified deployment.
func (r Runner) GetDeploymentManifest(deploymentName string) ([]byte, error) {
	output, err := r.boshExec("-d", deploymentName, "manifest")
	if err != nil {
		return nil, fmt.Errorf("retrieving bosh manifest failed: %w", err)
	}
	return output, nil
}

// GetDiegoDeployments returns only the bosh deployment names of deployments
// with diego cells: TAS, TASW, ISO segments.
func (r Runner) GetDiegoDeployments() ([]string, error) {
	allDeployments, err := r.GetDeployments()
	if err != nil {
		return nil, err
	}
	return filterDiegoDeployments(allDeployments), nil
}

// GetDeployments gets a bosh deployment names.
func (r Runner) GetDeployments() (deployments []string, err error) {
	output, err := r.boshExec("deployments", "--json")
	if err != nil {
		return nil, fmt.Errorf("retrieving bosh deployments failed: %w", err)
	}
	return loadDeployments(output)
}

// GetDeploymentVMs gets all the VMs from the specified deployment.
func (r Runner) GetDeploymentVMs(deploymentName string) (vms []VM, err error) {
	output, err := r.boshExec("vms", "-d", deploymentName, "--json")
	if err != nil {
		return nil, fmt.Errorf("retrieving bosh vms failed: %w", err)
	}
	return loadVMs(deploymentName, output)
}

// ScpFile copies a file using bosh scp
func (r Runner) ScpFile(deploymentName, source, target string) error {
	output, err := r.boshExec("-d", deploymentName, "scp", source, target)
	if err != nil {
		return fmt.Errorf("failed to SCP file from %s to %s (%s): %w", source, target, output, err)
	}
	return nil
}

// filterDiegoDeployments filter out deployments that don't contain diego cells
func filterDiegoDeployments(deployments []string) (diegoDeployments []string) {
	for _, d := range deployments {
		for _, p := range diegoDeploymentPrefixes {
			if strings.HasPrefix(d, p) {
				diegoDeployments = append(diegoDeployments, d)
			}
		}
	}
	return diegoDeployments
}

func (r *Runner) boshExec(args ...string) ([]byte, error) {
	cmd := exec.Command("bosh", args...)
	cmd.Env = r.env
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("could not execute bosh command: bosh %s: %w\n%s",
			strings.Join(args, " "), err, string(output))
	}
	return output, nil
}

func loadDeployments(output []byte) (deployments []string, err error) {
	type boshDeployments struct {
		Tables []struct {
			Rows []struct {
				Name string `json:"name,omitempty"`
			} `json:"Rows,omitempty"`
		} `json:"Tables,omitempty"`
	}

	var d boshDeployments
	err = json.Unmarshal(output, &d)
	if err != nil {
		return nil, fmt.Errorf("invalid json from bosh deployments: %w", err)
	}
	for _, row := range d.Tables[0].Rows {
		deployments = append(deployments, row.Name)
	}

	return deployments, nil
}
