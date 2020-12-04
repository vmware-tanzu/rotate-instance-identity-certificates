// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package om

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type errandConfig struct {
	RunPostDeploy map[string]bool `json:"run_post_deploy"`
}

type applyChangesRequest struct {
	IgnoreWarnings string                  `json:"ignore_warnings"`
	DeployProducts interface{}             `json:"deploy_products"`
	Errands        map[string]errandConfig `json:"errands,omitempty"`
}

// ApplyChanges will attempt to trigger the installation/update of the
// requested products. How it behaves depends on the value of products:
//
// * if products is omitted, it will attempt only to deploy the BOSH director
//
// * if products exists and the first element is "all", it will deploy all
// staged products. Otherwise, only the BOSH director and the specified
// products will deploy. In these cases, no errands will be run. These should
// be specified as product names (cf) instead of GUIDs (cf-1234asdf)
//
// if output is nil, the operation will return immediately and is effectively
// asynchronous. If not, the logs of the just-trigged installation will be
// streamed to output.
func (a *API) ApplyChanges(output io.Writer, ignoreWarnings bool, products ...string) error {
	deployedProducts, err := a.getDeployedProducts()
	if err != nil {
		return err
	}

	var toDeploy interface{}

	if len(products) > 0 {
		if products[0] == "all" {
			toDeploy = "all"
		} else {
			guids := []string{}
			for _, p := range products {
				found := false
				for _, d := range deployedProducts {
					if p == d.Name {
						found = true
						guids = append(guids, d.GUID)
						break
					}
				}

				if !found {
					return fmt.Errorf("unknown product %s", p)
				}
			}
			toDeploy = guids
		}
	}

	if len(products) == 0 {
		toDeploy = "none"
	}

	productGUIDs := make([]string, 0, len(deployedProducts))
	if toDeploy == "all" {
		for _, d := range deployedProducts {
			if d.Name == "p-bosh" {
				continue
			}

			productGUIDs = append(productGUIDs, d.GUID)
		}
	} else {
		for _, p := range products {
			for _, d := range deployedProducts {
				if p == d.Name {
					productGUIDs = append(productGUIDs, d.GUID)
					break
				}
			}
		}
	}

	body := applyChangesRequest{
		DeployProducts: toDeploy,
		IgnoreWarnings: strconv.FormatBool(ignoreWarnings),
	}

	errands := map[string]errandConfig{}

	for _, guid := range productGUIDs {
		ec, err := a.getErrandConfig(guid)
		if err != nil {
			return err
		}

		if len(ec.RunPostDeploy) > 0 {
			errands[guid] = ec
		}
	}

	if len(errands) > 0 {
		body.Errands = errands
	}

	bodyReader := &bytes.Buffer{}
	err = json.NewEncoder(bodyReader).Encode(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v0/installations", a.host), bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")

	client, err := a.getHTTPClient()
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w, got %d with body %s", ErrBadStatusCode, resp.StatusCode, buf.String())
	}

	acResponseBody := struct {
		Install struct {
			ID int `json:"id"`
		} `json:"install"`
	}{}

	if err = json.NewDecoder(buf).Decode(&acResponseBody); err != nil {
		return err
	}

	return a.streamLog(output)
}

func (a *API) streamLog(output io.Writer) error {
	if output == nil {
		return nil
	}

	client, err := a.getHTTPClient()
	if err != nil {
		return err
	}

	resp, err := client.Get(fmt.Sprintf("%s/api/v0/installations/current_log", a.host))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w, got %d instead", ErrBadStatusCode, resp.StatusCode)
	}

	lineReader := bufio.NewReader(resp.Body)
	gotExitEvent := false
	for {
		line, err := lineReader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				if strings.TrimSpace(line) == "" {
					return nil
				}

				parts := strings.SplitN(line, ":", 2)
				_, err = io.WriteString(output, parts[len(parts)-1])
				if err != nil {
					return err
				}
				return nil
			}

			return err
		}

		if strings.HasPrefix(line, "data:") {
			line := line[len("data:"):]
			_, err = io.WriteString(output, line)
			if err != nil {
				return err
			}

			if gotExitEvent {
				data := map[string]interface{}{}
				err = json.NewDecoder(strings.NewReader(line)).Decode(&data)
				if err != nil {
					return err
				}

				if code, ok := data["code"].(float64); ok && code != 0 {
					return fmt.Errorf("installation failed with code %d", int(code))
				}

				return nil
			}
		}

		if strings.TrimSpace(line) == "event:exit" {
			gotExitEvent = true
		}
	}
}

func (a *API) getErrandConfig(guid string) (errandConfig, error) {
	errandBody := struct {
		Errands []struct {
			Name       string `json:"name"`
			PostDeploy *bool  `json:"post_deploy,omitempty"`
		} `json:"errands"`
	}{}

	client, err := a.getHTTPClient()
	if err != nil {
		return errandConfig{}, err
	}

	resp, err := client.Get(fmt.Sprintf("%s/api/v0/staged/products/%s/errands", a.host, guid))
	if err != nil {
		return errandConfig{}, err
	}
	defer resp.Body.Close()

	buf := &bytes.Buffer{}
	_, err = io.Copy(buf, resp.Body)
	if err != nil {
		return errandConfig{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return errandConfig{}, fmt.Errorf("%w, got %d with body %s", ErrBadStatusCode, resp.StatusCode, buf.String())
	}

	err = json.NewDecoder(buf).Decode(&errandBody)
	if err != nil {
		return errandConfig{}, err
	}

	errandMap := map[string]bool{}
	for _, errand := range errandBody.Errands {
		if errand.PostDeploy != nil {
			errandMap[errand.Name] = false
		}
	}

	return errandConfig{
		RunPostDeploy: errandMap,
	}, nil
}
