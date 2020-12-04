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
	"strings"
)

// GetDirectorCredentials will return a slice of environment variables (in the form KEY=VALUE),
// suitable for use in exec.Command, that will facilitate connections to the Ops Manager's
// BOSH Director
func (a *API) GetDirectorCredentials() ([]string, error) {
	if err := a.EnsureAvailability(30); err != nil {
		return nil, err
	}

	client, err := a.getHTTPClient()
	if err != nil {
		return nil, err
	}

	respBody := map[string]string{}
	resp, err := client.Get(fmt.Sprintf("%s/api/v0/deployed/director/credentials/bosh_commandline_credentials", a.host))
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

	if err = json.NewDecoder(buf).Decode(&respBody); err != nil {
		return nil, err
	}

	var cred string
	var ok bool
	if cred, ok = respBody["credential"]; !ok {
		return nil, errors.New("empty credential field")
	}

	cred = strings.TrimSuffix(cred, " bosh ")
	result := strings.Split(cred, " ")
	l := len(result)
	for i := 0; i < l; i++ {
		parts := strings.Split(result[i], "=")
		switch parts[0] {
		case "BOSH_CLIENT":
			result = append(result, "CREDHUB_CLIENT="+parts[1])
		case "BOSH_CLIENT_SECRET":
			result = append(result, "CREDHUB_SECRET="+parts[1])
		case "BOSH_CA_CERT":
			result = append(result, "CREDHUB_CA_CERT="+parts[1])
		case "BOSH_ENVIRONMENT":
			result = append(result, "CREDHUB_SERVER=https://"+parts[1]+":8844")
		}
	}
	return result, nil
}
