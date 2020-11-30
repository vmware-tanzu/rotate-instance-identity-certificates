// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package om_test

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/om"
)

const uaaToken = "abcdefgh"

func TestEnsureAvailability(t *testing.T) {
	handlers := map[string]http.Handler{}
	runner := func(numRetries uint, decryptionPassphrase string, shouldFail bool, shouldBeUnauthorized bool) {
		handlers["/api/v0/unlock"] = unlockHandler(2, "test-passphrase")

		server := getServer(handlers, true)
		defer server.Close()
		t.Run(fmt.Sprintf("with %d retries", numRetries), func(t *testing.T) {
			api := om.NewAPI(server.URL, "TESTING", "TESTING", decryptionPassphrase, false, getClient())
			err := api.EnsureAvailability(numRetries)
			if shouldFail {
				if err == nil {
					t.Fatal("an error was expected but it did not occur")
				}

				if shouldBeUnauthorized && !errors.Is(err, om.ErrBadStatusCode) {
					t.Fatalf("an error occured that wasn't the expected bad status code error: %v", err)
				}

				return
			}

			if err != nil {
				t.Fatalf("an unexpected error occured: %v", err)
			}
		})
	}

	runner(1, "", true, false)
	runner(2, "foo", true, false)
	runner(3, "asdf", true, true)
	runner(3, "test-passphrase", false, false)
}

func getClient() *http.Client {
	return &http.Client{
		Transport: &teapotHandler{
			rt: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		},
	}
}

func getServer(handlers map[string]http.Handler, includeDefaultHandlers bool) *httptest.Server {
	mux := http.NewServeMux()

	if includeDefaultHandlers {
		mux.HandleFunc("/uaa/oauth/token", uaaHandlerFunc)
	}

	for pattern, handler := range handlers {
		mux.Handle(pattern, handler)
	}

	return httptest.NewTLSServer(mux)
}

func uaaHandlerFunc(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	writeString(w, fmt.Sprintf(`{"access_token":%q}`, uaaToken))
}

func unlockHandler(requiredRetries int, decryptionPassphrase string) http.Handler {
	attempts := 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if r.Body != nil {
			defer r.Body.Close()
		}

		if attempts < requiredRetries {
			w.WriteHeader(http.StatusTeapot)
			return
		}

		if decryptionPassphrase != "" {
			m := map[string]string{}
			e := json.NewDecoder(r.Body).Decode(&m)
			if e != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			if decryptionPassphrase != m["passphrase"] {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("{}"))
		if err != nil {
			panic(err)
		}
	})
}

type teapotHandler struct {
	rt http.RoundTripper
}

func (t *teapotHandler) RoundTrip(req *http.Request) (*http.Response, error) {
	r, e := t.rt.RoundTrip(req)
	if r != nil && r.StatusCode == http.StatusTeapot {
		return nil, errors.New("teapot")
	}

	return r, e
}
