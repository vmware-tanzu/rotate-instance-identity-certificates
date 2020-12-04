// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package om

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CheckPendingChanges will return true if an apply
// changes would actually do anything (config changes,
// version updates, stemcell updates, etc)
func (a *API) CheckPendingChanges() (bool, error) {
	err := a.EnsureAvailability(30)
	if err != nil {
		return false, err
	}

	client, err := a.getHTTPClient()
	if err != nil {
		return false, err
	}

	resp, err := client.Get(fmt.Sprintf("%s/api/v0/staged/pending_changes", a.host))
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return false, err
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("%w, got %d with body %s", ErrBadStatusCode, resp.StatusCode, buf.String())
	}

	pendingChanges := struct {
		Products []struct {
			GUID   string `json:"guid"`
			Action string `json:"action"`
		} `json:"product_changes"`
	}{}

	err = json.NewDecoder(buf).Decode(&pendingChanges)
	if err != nil {
		return false, err
	}

	for _, p := range pendingChanges.Products {
		if p.Action != "unchanged" {
			return true, nil
		}
	}

	return false, nil
}
