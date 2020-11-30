// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

// Package om provides methods to interact with an Ops Manager via its API.
package om

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// API will allow access to an Ops Manager's HTTP API
type API struct {
	host                 string
	username             string
	password             string
	decryptionPassphrase string
	useClientCredentials bool
	ctx                  context.Context
}

// ErrBadStatusCode is an error that is thrown when a non-200 status code is returned from an HTTP call. It can be read with
//     errors.Is(err, om.ErrBadStatusCode)
var ErrBadStatusCode = errors.New("expected 2xx error code")

// NewAPI will create a new API object. If you want to use client id and client secret instead of username and password, set
// useClientCredentials to true. Note that host must begin with the scheme (https://example.com, not example.com)
//
// If you pass a nil http.Client, http.DefaultClient will be used for all calls. If you require skipping TLS validation or using
// custom certs, you must create a Client that does those things and pass it in.
func NewAPI(host string, username string, password string, decryptionPassphrase string, useClientCredentials bool, hc *http.Client) *API {
	a := &API{
		host:                 host,
		username:             username,
		password:             password,
		decryptionPassphrase: decryptionPassphrase,
		useClientCredentials: useClientCredentials,
	}

	if hc == nil {
		hc = http.DefaultClient
	}

	a.ctx = context.WithValue(context.Background(), oauth2.HTTPClient, hc)
	return a
}

func (a *API) getHTTPClient() (*http.Client, error) {
	tokenURL := fmt.Sprintf("%s/uaa/oauth/token", a.host)

	if a.useClientCredentials {
		config := &clientcredentials.Config{
			ClientID:     a.username,
			ClientSecret: a.password,
			Scopes:       []string{"opsman.admin"},
			TokenURL:     tokenURL,
		}

		return config.Client(a.ctx), nil
	}

	// opsman is an implicit client with no secret, used to get password tokens
	config := oauth2.Config{
		ClientID:     "opsman",
		ClientSecret: "",
		Endpoint: oauth2.Endpoint{
			TokenURL: tokenURL,
		},
	}

	token, err := config.PasswordCredentialsToken(a.ctx, a.username, a.password)
	if err != nil {
		return nil, err
	}

	return config.Client(a.ctx, token), nil
}

// EnsureAvailability will attempt up to numRetries times to unlock the API
// with the decryption passphrase provided in NewAPI.
func (a *API) EnsureAvailability(numRetries uint) error {
	var i uint = 0
	for ; i < numRetries; i++ {
		err := a.unlock()
		if err == nil {
			return nil
		}

		// only retry if it timed out, not if a non-200 status was returned
		if errors.Is(err, ErrBadStatusCode) {
			return err
		}

		if a.username != "TESTING" && a.password != "TESTING" {
			time.Sleep(10 * time.Second)
		}
	}

	return fmt.Errorf("failed to unlock after %d tries", numRetries+1)
}

func (a *API) unlock() error {
	body := bytes.NewBufferString(fmt.Sprintf(`{"passphrase":%q}`, a.decryptionPassphrase))

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v0/unlock", a.host), body)
	if err != nil {
		return err
	}

	var retrieveError *oauth2.RetrieveError

	client, err := a.getHTTPClient() // this makes an OpsManager UAA call, and should be tested for unauthorized
	if err != nil {
		if errors.As(err, &retrieveError) {
			if retrieveError.Response.StatusCode != http.StatusOK {
				return fmt.Errorf("%w, got %d with body %s", ErrBadStatusCode, retrieveError.Response.StatusCode, string(retrieveError.Body))
			}
		}
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.As(err, &retrieveError) {
			if retrieveError.Response.StatusCode != http.StatusOK {
				return fmt.Errorf("%w, got %d with body %s", ErrBadStatusCode, retrieveError.Response.StatusCode, string(retrieveError.Body))
			}
		}

		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := &bytes.Buffer{}
		_, err = io.Copy(buf, resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("%w, got %d with body %s", ErrBadStatusCode, resp.StatusCode, buf.String())
	}

	return nil
}
