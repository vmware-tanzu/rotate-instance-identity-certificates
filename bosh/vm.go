// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package bosh

import (
	"encoding/json"
	"fmt"
)

// VM is a bosh instance
type VM struct {
	DeploymentName string
	Name           string
}

// newVM creats a new bosh VM instance
func newVM(deploymentName, name string) VM {
	return VM{
		DeploymentName: deploymentName,
		Name:           name,
	}
}

func loadVMs(deploymentName string, output []byte) (vms []VM, err error) {
	type boshVms struct {
		Tables []struct {
			Rows []struct {
				Instance string `json:"instance,omitempty"`
			} `json:"Rows,omitempty"`
		} `json:"Tables,omitempty"`
	}

	var v boshVms
	err = json.Unmarshal(output, &v)
	if err != nil {
		return nil, fmt.Errorf("invalid json from bosh vms: %w", err)
	}

	for _, row := range v.Tables[0].Rows {
		vm := newVM(deploymentName, row.Instance)
		vms = append(vms, vm)
	}

	return vms, nil
}
