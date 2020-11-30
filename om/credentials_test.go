// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package om_test

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/om"
)

func TestGetDirectorCredentials(t *testing.T) {
	handlers := map[string]http.Handler{
		"/api/v0/unlock": unlockHandler(0, ""),
		"/api/v0/deployed/director/credentials/bosh_commandline_credentials": http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusOK)

			writeString(w, `{"credential":"BOSH_CLIENT=ops_manager BOSH_CLIENT_SECRET=SHLkTXPTQwtFAKEl4icrmgVl6N7-Wmz6 BOSH_CA_CERT=/var/tempest/workspaces/default/root_ca_certificate BOSH_ENVIRONMENT=172.31.0.11 bosh "}`)
		}),
	}

	server := getServer(handlers, true)
	defer server.Close()

	api := om.NewAPI(server.URL, "", "", "", true, getClient())
	creds, err := api.GetDirectorCredentials()
	if err != nil {
		t.Fatalf("unexpected error occured: %v", err)
	}

	expectedCreds := []string{
		"BOSH_CLIENT=ops_manager",
		"BOSH_CLIENT_SECRET=SHLkTXPTQwtFAKEl4icrmgVl6N7-Wmz6",
		"BOSH_CA_CERT=/var/tempest/workspaces/default/root_ca_certificate",
		"BOSH_ENVIRONMENT=172.31.0.11",
		"CREDHUB_CLIENT=ops_manager",
		"CREDHUB_SECRET=SHLkTXPTQwtFAKEl4icrmgVl6N7-Wmz6",
		"CREDHUB_CA_CERT=/var/tempest/workspaces/default/root_ca_certificate",
		"CREDHUB_SERVER=https://172.31.0.11:8844",
	}

	if !reflect.DeepEqual(creds, expectedCreds) {
		t.Fatalf("expected %v to match %v", creds, expectedCreds)
	}
}
