// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package om_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/mitchellh/pointerstructure"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/om"
)

var streamCalled = false

func TestApplyChanges(t *testing.T) {
	requestBody := make(chan io.Reader, 1)

	handlers := map[string]http.Handler{
		"/api/v0/unlock":                    unlockHandler(0, ""),
		"/api/v0/installations":             startInstallationHandler(t, requestBody),
		"/api/v0/staged/products/":          http.HandlerFunc(errandConfigHandler),
		"/api/v0/deployed/products":         http.HandlerFunc(getDeployedProductsHandler),
		"/api/v0/installations/current_log": http.HandlerFunc(currentLogHandler),
	}

	server := getServer(handlers, true)
	defer server.Close()

	api := om.NewAPI(server.URL, "", "", "", true, getClient())

	t.Run("director changes only without streaming", func(t *testing.T) {
		streamCalled = false
		err := api.ApplyChanges(nil, true)
		if err != nil {
			t.Fatalf("unexpected error occured: %s", err)
		}

		if streamCalled {
			t.Fatalf("logs were streaming and shouldn't have been")
		}

		reqBody := <-requestBody

		var request interface{}
		err = json.NewDecoder(reqBody).Decode(&request)
		if err != nil {
			t.Fatalf("got invalid JSON: %v", err)
		}

		deployVal, err := pointerstructure.Get(request, "/deploy_products")
		if err != nil {
			t.Fatalf("unexpected error occured: %s", err)
		}

		if !reflect.DeepEqual(deployVal, "none") {
			t.Fatalf("expected deploy_products to be none, but it was %v", deployVal)
		}

		ignoreWarnings, err := pointerstructure.Get(request, "/ignore_warnings")
		if err != nil {
			t.Fatalf("unexpected error occured: %s", err)
		}
		if ignoreWarnings.(string) != "true" {
			t.Fatalf("expected ignore_warnings to be true, got %v", ignoreWarnings)
		}

		_, err = pointerstructure.Get(request, "/errands")
		if err == nil || !strings.Contains(err.Error(), "couldn't find key") {
			t.Fatalf("the expected error did not occur. the error that did occur is %v", err)
		}
	})

	t.Run("apply all changes without streaming", func(t *testing.T) {
		streamCalled = false
		err := api.ApplyChanges(nil, false, "all")
		if err != nil {
			t.Fatalf("unexpected error occured: %s", err)
		}

		if streamCalled {
			t.Fatalf("logs were streaming and shouldn't have been")
		}

		reqBody := <-requestBody

		var request interface{}
		err = json.NewDecoder(reqBody).Decode(&request)
		if err != nil {
			t.Fatalf("got invalid JSON: %v", err)
		}

		deployVal, err := pointerstructure.Get(request, "/deploy_products")
		if err != nil {
			t.Fatalf("unexpected error occured: %s", err)
		}

		if !reflect.DeepEqual(deployVal, "all") {
			t.Fatalf("expected deploy_products to be none, but it was %v", deployVal)
		}

		errand, err := pointerstructure.Get(request, "/errands/component-type1-guid/run_post_deploy/errand-1")
		if err != nil {
			t.Fatalf("unexpected error occured: %v", err)
		}

		val, ok := errand.(bool)
		if !ok {
			t.Fatalf("expected value to be bool, but it was %T", errand)
		}

		if val {
			t.Fatal("expected value to be false, but it was true")
		}
	})

	t.Run("apply product changes with streaming", func(t *testing.T) {
		buf := &bytes.Buffer{}

		streamCalled = false
		if err := api.ApplyChanges(buf, false, "component-type1"); err != nil {
			t.Fatalf("unexpected error occured: %v", err)
		}
		if !streamCalled {
			t.Fatal("should have streamed logs, but they weren't")
		}

		reqBody := <-requestBody

		var request interface{}
		if err := json.NewDecoder(reqBody).Decode(&request); err != nil {
			t.Fatalf("unexpected error occured: %v", err)
		}

		prods, err := pointerstructure.Get(request, "/deploy_products")
		if err != nil {
			t.Fatalf("unexpected error occured: %v", err)
		}

		if !reflect.DeepEqual(prods, []interface{}{"component-type1-guid"}) {
			t.Fatalf("expected deploy_products to be [component-type1-guid] but it was %v", prods)
		}

		lines := strings.Split(buf.String(), "\n")
		if len(lines) != 52 {
			t.Fatalf("expected 52 lines of logs, got %d", len(lines))
		}
	})
}

func TestApplyChangesWithError(t *testing.T) {
	requestBody := make(chan io.Reader, 1)

	handlers := map[string]http.Handler{
		"/api/v0/unlock":                    unlockHandler(0, ""),
		"/api/v0/installations":             startInstallationHandler(t, requestBody),
		"/api/v0/staged/products/":          http.HandlerFunc(errandConfigHandler),
		"/api/v0/deployed/products":         http.HandlerFunc(getDeployedProductsHandler),
		"/api/v0/installations/current_log": http.HandlerFunc(currentLogWithErrorHandler),
	}

	server := getServer(handlers, true)
	defer server.Close()

	api := om.NewAPI(server.URL, "", "", "", true, getClient())
	err := api.ApplyChanges(ioutil.Discard, false)
	if err == nil || err.Error() != "installation failed with code 1" {
		t.Fatalf("an expected error did not occur. the error that did occur was %v", err)
	}
}

func startInstallationHandler(t *testing.T, requestBody chan io.Reader) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusOK)

		defer r.Body.Close()
		buf := &bytes.Buffer{}
		_, err := io.Copy(buf, r.Body)
		if err != nil {
			panic(err)
		}

		requestBody <- buf

		writeString(w, `{"install":{"id":1}}`)
	})
}

func currentLogHandler(w http.ResponseWriter, r *http.Request) {
	streamCalled = true
	eventCount := 5
	dataPerEvent := 10
	lineTime := 5 * time.Millisecond

	w.Header().Set("content-type", "text/event-stream")
	w.WriteHeader(http.StatusOK)

	for e := 0; e < eventCount; e++ {
		writeString(w, fmt.Sprintf("event:step-%d\n", e))
		for d := 0; d < dataPerEvent; d++ {
			writeString(w, fmt.Sprintf("data:substep %d of step %d\n", d, e))
			time.Sleep(lineTime)
		}
	}

	writeString(w, fmt.Sprintln("event:exit"))
	writeString(w, fmt.Sprintln(`data:{"type":"exit","code":0}`))
}

func currentLogWithErrorHandler(w http.ResponseWriter, r *http.Request) {
	streamCalled = true
	eventCount := 1
	dataPerEvent := 3
	lineTime := 5 * time.Millisecond

	w.Header().Set("content-type", "text/event-stream")
	w.WriteHeader(http.StatusOK)

	for e := 0; e < eventCount; e++ {
		writeString(w, fmt.Sprintf("event:step-%d\n", e))
		for d := 0; d < dataPerEvent; d++ {
			writeString(w, fmt.Sprintf("data:substep %d of step %d\n", d, e))
			time.Sleep(lineTime)
		}
	}

	writeString(w, fmt.Sprintln("event:exit"))
	writeString(w, fmt.Sprintln(`data:{"type":"exit","code":1}`))
}

func errandConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)

	writeString(w, `{
	"errands": [
	{
		"name": "errand-1",
		"post_deploy": false,
		"label": "Errand 1 Label"
	},
	{
		"name": "errand-2",
		"pre_delete": true,
		"label": "Errand 2 Label"
	},
	{
		"name": "shared-errand",
		"post_deploy": false,
		"pre_delete": true,
		"label": "Shared Errand Label"
	}
	]
}`)

}

func writeString(w io.Writer, s string) {
	_, err := io.WriteString(w, s)
	if err != nil {
		panic(err)
	}
}
