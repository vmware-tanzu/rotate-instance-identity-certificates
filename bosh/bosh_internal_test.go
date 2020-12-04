// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package bosh

import (
	"io/ioutil"
	"testing"
)

func TestFilterDiegoDeployments(t *testing.T) {
	dd := []string{"mysql-guid", "pas-windows-f9239c09b3772fdf6a12", "p-isolation-segment-56025cc76d5913ca1a69", "p-isolation-segment-94221dd76d5964cc3b44", "cf-3e6b71ab5a6736db362b"}

	deployments := filterDiegoDeployments(dd)
	if len(deployments) != 4 {
		t.Fatalf("Expected 4 diego deployments, but found %d", len(deployments))
	}

	if !containsDeployment(deployments, "pas-windows-f9239c09b3772fdf6a12") {
		t.Fatal("Expected diego deployments to contain pas-windows-f9239c09b3772fdf6a12")
	}
	if !containsDeployment(deployments, "p-isolation-segment-56025cc76d5913ca1a69") {
		t.Fatal("Expected diego deployments to contain p-isolation-segment-56025cc76d5913ca1a69")
	}
	if !containsDeployment(deployments, "p-isolation-segment-94221dd76d5964cc3b44") {
		t.Fatal("Expected diego deployments to contain p-isolation-segment-94221dd76d5964cc3b44")
	}
	if !containsDeployment(deployments, "cf-3e6b71ab5a6736db362b") {
		t.Fatal("Expected diego deployments to contain cf-3e6b71ab5a6736db362b")
	}
}

func TestLoadDeployments(t *testing.T) {
	f, err := ioutil.ReadFile("testdata/deployments.json")
	if err != nil {
		t.Fatalf("Failed to read test data deployments.json: %s", err)
	}

	deployments, err := loadDeployments(f)
	if err != nil {
		t.Fatalf("Failed to parse deployments from deployments.json: %s", err)
	}

	if len(deployments) != 3 {
		t.Fatalf("Expected 3 deployments but got %d", len(deployments))
	}
	if !containsDeployment(deployments, "cf-3e6b71ab5a6736db362b") {
		t.Fatal("Expected deployments to contain cf-3e6b71ab5a6736db362b")
	}
	if !containsDeployment(deployments, "p-isolation-segment-56025cc76d5913ca1a69") {
		t.Fatal("Expected deployments to contain p-isolation-segment-56025cc76d5913ca1a69")
	}
	if !containsDeployment(deployments, "pas-windows-f9239c09b3772fdf6a12") {
		t.Fatal("Expected deployments to contain pas-windows-f9239c09b3772fdf6a12")
	}
}

func containsDeployment(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
