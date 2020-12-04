// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package om

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mitchellh/pointerstructure"
)

func (a *API) GetBoshDirectorName() (string, error) {
	if err := a.EnsureAvailability(30); err != nil {
		return "", err
	}

	client, err := a.getHTTPClient()
	if err != nil {
		return "", err
	}

	resp, err := client.Get(fmt.Sprintf("%s/api/v0/deployed/director/manifest", a.host))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	manifest := struct {
		InstanceGroups []interface{} `json:"instance_groups"`
	}{}

	if err = json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return "", err
	}

	for _, ig := range manifest.InstanceGroups {
		if s, e := getStr(ig, "/name"); e == nil && s == "bosh" {
			return getStr(ig, "/properties/director/name")
		}
	}

	return "", fmt.Errorf("could not determine bosh director name from its manifest")
}

// GetBoshManifest returns the latest attemped OpsMan generated BOSH manifest
func (a *API) GetBoshManifest(deploymentName string) ([]byte, error) {
	if err := a.EnsureAvailability(30); err != nil {
		return nil, err
	}

	client, err := a.getHTTPClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(fmt.Sprintf("%s/api/v0/deployed/products/%s/manifest", a.host, deploymentName))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w, got %d with body %s", ErrBadStatusCode, resp.StatusCode, buf.String())
	}

	return buf.Bytes(), nil
}

func getStr(obj interface{}, query string) (string, error) {
	intf, err := pointerstructure.Get(obj, query)
	if err != nil {
		return "", err
	}

	if str, ok := intf.(string); ok {
		return str, nil
	}

	return "", fmt.Errorf("expected string at %s but it was a %T", query, intf)
}
