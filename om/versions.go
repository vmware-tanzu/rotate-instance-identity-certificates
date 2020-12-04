// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package om

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type omVersion struct {
	Info struct {
		Version string `json:"version"`
	} `json:"info"`
}

type deployedProduct struct {
	GUID    string `json:"guid"`
	Name    string `json:"type"`
	Version string `json:"product_version"`
}

type deployedProducts []deployedProduct

// GetOpsManagerVersion returns the version of Ops Manager as reported by the /api/v0/info endpoint
func (a *API) GetOpsManagerVersion() (string, error) {
	client, err := a.getHTTPClient()
	if err != nil {
		return "", err
	}

	resp, err := client.Get(fmt.Sprintf("%s/api/v0/info", a.host))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w, got %d with body %s", ErrBadStatusCode, resp.StatusCode, buf.String())
	}

	body := omVersion{}
	err = json.NewDecoder(buf).Decode(&body)
	if err != nil {
		return "", fmt.Errorf("got unexpected JSON from Ops Manager: %w", err)
	}

	return body.Info.Version, nil
}

// GetDeployedProductVersion will return the version of the requested product that
// is currently deployed. It will error if the product is p-bosh. For that, you should use
// GetOpsManagerVersion. It will also error if the product is not found.
func (a *API) GetDeployedProductVersion(productName string) (string, error) {
	if productName == "p-bosh" {
		return "", errors.New("this method cannot return the version of the BOSH director")
	}

	products, err := a.getDeployedProducts()
	if err != nil {
		return "", err
	}

	for _, product := range products {
		if product.Name == productName {
			return product.Version, nil
		}
	}

	return "", fmt.Errorf("unable to find deployed product with name %s", productName)
}

func (a *API) getDeployedProducts() (deployedProducts, error) {
	err := a.EnsureAvailability(30)
	if err != nil {
		return nil, err
	}

	client, err := a.getHTTPClient()
	if err != nil {
		return nil, err
	}

	resp, err := client.Get(fmt.Sprintf("%s/api/v0/deployed/products", a.host))
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

	var products deployedProducts
	err = json.NewDecoder(buf).Decode(&products)
	if err != nil {
		return nil, err
	}

	return products, nil
}
