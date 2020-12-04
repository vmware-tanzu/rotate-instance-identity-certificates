// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package om_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/om"
)

func TestGetOpsManagerVersion(t *testing.T) {
	handlers := map[string]http.Handler{}

	handlers["/api/v0/info"] = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		writeString(w, `{"info":{"version":"2.4"}}`)
	})
	server := getServer(handlers, true)
	defer server.Close()

	api := om.NewAPI(server.URL, "", "", "", true, getClient())
	version, err := api.GetOpsManagerVersion()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if version != "2.4" {
		t.Fatalf("expected version to be 2.4, but it was %s", version)
	}
}

func TestGetDeployedProductVersion(t *testing.T) {
	handlers := map[string]http.Handler{}

	handlers["/api/v0/unlock"] = unlockHandler(0, "")
	handlers["/api/v0/deployed/products"] = http.HandlerFunc(getDeployedProductsHandler)

	server := getServer(handlers, true)
	defer server.Close()

	api := om.NewAPI(server.URL, "", "", "", true, getClient())

	runner := func(productName string, expectedVersion string, shouldFail bool) {
		t.Run(fmt.Sprintf("test-%s", productName), func(t *testing.T) {
			version, err := api.GetDeployedProductVersion(productName)
			if shouldFail {
				if err == nil {
					t.Fatalf("an expected error did not occur")
				}

				return
			}

			if err != nil {
				t.Fatalf("an unexpected error occured: %s", err)
			}

			if version != expectedVersion {
				t.Fatalf("expected version %s, got version %s", expectedVersion, version)
			}
		})
	}

	runner("p-bosh", "", true)
	runner("component-type1", "1.0", false)
	runner("cf", "2.4.27", true)
}

func getDeployedProductsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	writeString(w, deployedProducts)
}

const deployedProducts = `
[
  {
    "installation_name": "component-type1-installation-name",
    "guid": "component-type1-guid",
    "type": "component-type1",
    "product_version": "1.0",
    "stale": {
      "parent_products_deployed_more_recently": ["p-bosh-guid"]
    }
  },
  {
    "installation_name": "p-bosh-installation-name",
    "guid": "p-bosh-guid",
    "type": "p-bosh",
    "stale": {
      "parent_products_deployed_more_recently": []
    }
  }
]
`
