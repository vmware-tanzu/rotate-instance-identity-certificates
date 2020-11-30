// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package om_test

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/vmware-tanzu/rotate-instance-identity-certificates/om"
)

func TestPendingChanges(t *testing.T) {
	handlers := map[string]http.Handler{}

	runner := func(output string, requiredToken string, expectedResult bool, shouldFail bool) {
		handlers["/api/v0/unlock"] = unlockHandler(0, "")
		handlers["/api/v0/staged/pending_changes"] = pendingChanges(output, requiredToken, shouldFail)
		server := getServer(handlers, true)
		defer server.Close()

		t.Run("", func(t *testing.T) {
			api := om.NewAPI(server.URL, "", "", "", true, getClient())
			changes, err := api.CheckPendingChanges()
			if shouldFail {
				if err == nil {
					t.Fatal("an error was expected but it did not occur")
				}

				if !errors.Is(err, om.ErrBadStatusCode) {
					t.Fatal("only http status code errors are expected here")
				}

				return
			}

			if err != nil {
				t.Fatalf("an unexpected error occured: %v", err)
			}

			if changes != expectedResult {
				t.Fatalf("expected return value %t, but got %t", expectedResult, changes)
			}
		})
	}

	runner(noProducts, uaaToken, false, false)
	runner(noChanges, uaaToken, false, false)
	runner(changes, uaaToken, true, false)
	runner(noProducts, "bad-token", false, true)
}

func pendingChanges(output string, requiredToken string, non200Return bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			writeString(w, `{"message":"unauthorized"}`)
			return
		}

		if authHeader[len("bearer "):] != requiredToken {
			w.WriteHeader(http.StatusUnauthorized)
			writeString(w, `{"message":"unauthorized"}`)
			return
		}

		if non200Return {
			w.WriteHeader(http.StatusInternalServerError)
			writeString(w, `{"message":"internal server error"}`)
			return
		}

		w.WriteHeader(http.StatusOK)
		writeString(w, output)
	})
}

const noProducts = `{"product_changes": []}`
const noChanges = `{"product_changes":[{"guid": "product1-abdc", "action": "unchanged"}]}`
const changes = `{"product_changes":[{"guid": "product1-abdc", "action": "unchanged"},{"guid": "product2-efgh", "action": "update"}]}`
