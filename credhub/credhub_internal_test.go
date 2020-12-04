// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package credhub

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestImportCerts(t *testing.T) {
	certs := []Certificate{
		*NewCertificate("diego-ca", "certificate", "ca-cert", "ca-key", "ca-cert"),
		*NewCertificate("diego-intermediate", "certificate", "intermediate-cert", "intermediate-key", "ca-cert"),
	}

	r := NewRunner([]string{})
	r.credhubExec = func(args ...string) ([]byte, error) {
		if len(args) != 3 {
			t.Fatalf("Expected 3 credhub CLI flags, but got %d", len(args))
		}
		if args[0] != "import" {
			t.Fatalf("Expected import command but got %s", args[0])
		}
		if args[1] != "-f" {
			t.Fatalf("Expected -f flag but got %s", args[1])
		}

		importYaml, err := ioutil.ReadFile(args[2])
		if err != nil {
			t.Fatal(err)
		}

		y := string(importYaml)
		if strings.TrimSpace(y) != strings.TrimSpace(expectedImportContent) {
			t.Fatalf("Expected import credential file to match:\n%s\nBut got:\n%s", expectedImportContent, y)
		}

		return nil, nil
	}

	err := r.ImportCertificates(certs)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetCert(t *testing.T) {
	r := NewRunner([]string{})
	r.credhubExec = func(args ...string) ([]byte, error) {
		if len(args) != 3 {
			t.Fatalf("Expected 3 credhub CLI flags, but got %d", len(args))
		}
		if args[0] != "get" {
			t.Fatalf("Expected get command but got %s", args[0])
		}
		if args[1] != "-n" {
			t.Fatalf("Expected -n flag but got %s", args[1])
		}

		return []byte(credhubGetOutput), nil
	}

	cert, err := r.GetCertificate("/cf/diego-instance-identity-root-ca")
	if err != nil {
		t.Fatal(err)
	}

	if cert.Name != "/cf/diego-instance-identity-root-ca" {
		t.Fatalf("Expected cert name to be /cf/diego-instance-identity-root-ca but got %s", cert.Name)
	}
	if cert.Type != "certificate" {
		t.Fatalf("Expected type certificate but got %s", cert.Type)
	}
	if !strings.Contains(cert.Value.CA, "MIIDNTCCAh2gAwIBAgIUegQO0vm421OAlOYdH8iI776CkzEwDQYJKoZIhvcNAQEL") {
		t.Fatalf("Expected CA cert to match CA entry:\n%s\nbut got:\n%s", credhubGetOutput, cert.Value.CA)
	}
	if !strings.Contains(cert.Value.Certificate, "MIIDNTCCAh2gAwIBAgIUegQO0vm421OAlOYdH8iI776CkzEwDQYJKoZIhvcNAQEL") {
		t.Fatalf("Expected cert to match certificate entry:\n%s\nbut got:\n%s", credhubGetOutput, cert.Value.Certificate)
	}
	if !strings.Contains(cert.Value.PrivateKey, "MIIEpQIBAAKCAQEAzv786ToM7VYmgqiso31U2ssmV/4hkGGqt5nM8x17rQQ0jL4F") {
		t.Fatalf("Expected private key to match private_key entry:\n%s\nbut got:\n%s", credhubGetOutput, cert.Value.PrivateKey)
	}
}

const expectedImportContent = `
credentials:
- name: diego-ca
  type: certificate
  value:
    certificate: ca-cert
    private_key: ca-key
    ca: ca-cert
- name: diego-intermediate
  type: certificate
  value:
    certificate: intermediate-cert
    private_key: intermediate-key
    ca: ca-cert
`

const credhubGetOutput = `
id: dc4711e9-9429-4fee-9bff-d415573f4df2
name: /cf/diego-instance-identity-root-ca
type: certificate
value:
  ca: |
    -----BEGIN CERTIFICATE-----
    MIIDNTCCAh2gAwIBAgIUegQO0vm421OAlOYdH8iI776CkzEwDQYJKoZIhvcNAQEL
    BQAwKjEoMCYGA1UEAxMfRGllZ28gSW5zdGFuY2UgSWRlbnRpdHkgUm9vdCBDQTAe
    Fw0yMDEwMjYxNjU0MTZaFw0yMzEwMjYxNjU0MTZaMCoxKDAmBgNVBAMTH0RpZWdv
    IEluc3RhbmNlIElkZW50aXR5IFJvb3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IB
    DwAwggEKAoIBAQDO/vzpOgztViaCqKyjfVTayyZX/iGQYaq3mczzHXutBDSMvgU2
    Ig/+sxFFRD8wOTTjCf3gGiOWcLYyZcdkKPEHbeLG8eyGSwLHAhcrl8g/HUK2BT2d
    uVvCwOefARsBuULHGaeO0vIugCitnRS15+BScfWJxYSn8hxjynG3CvYpcjbq1a4O
    0575DbKYXO7e1S662A8n5QjsDb5QXDC2mDuomc5g+UHt11auXv2FUo0AwdRkT2rb
    7GJvvChek+C5WnHU6gs+ivTAbUgGwQA4xun5V8i10LDoFwB8IUq6HSodPrAvnALU
    347KV9rTjCajiBMMt4HhuTJJ9CI4f9/rLcCtAgMBAAGjUzBRMB0GA1UdDgQWBBQW
    3SPVggvcy9UPsCAHs0fAYs6D1jAfBgNVHSMEGDAWgBQW3SPVggvcy9UPsCAHs0fA
    Ys6D1jAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBl5PTtTpkx
    zlxYR1DXAaNYhEvYwVDcINDhJiSKvK6uZnnlsHrl98oyU+9c2yyZtn2eiia/Ngn0
    WEzBNw/hUxXlmdGpCS+3NW0AWaTPyxOT8yrxLkd+DBeEznzHkHI4agszHMIqvrvl
    L+Wpj2XRnl8WVDvMnFfQt3t6vfpVtWoIaBHRyDS8OF0YEV/28FN76D1JzeUss3D3
    MRo0H3ZDLakZow+aGFW3jIhU6+ZJ41J+YBHl4H8lN9L7jkIWYFx257i/iZSz775j
    LzqLdDnRFZZqgPsJA1drkIYyIiAz80+NxxUxR58Hmc90B7AuMOvjZXSFSxGdbccc
    9AM78hx9Ak6c
    -----END CERTIFICATE-----
  certificate: |
    -----BEGIN CERTIFICATE-----
    MIIDNTCCAh2gAwIBAgIUegQO0vm421OAlOYdH8iI776CkzEwDQYJKoZIhvcNAQEL
    BQAwKjEoMCYGA1UEAxMfRGllZ28gSW5zdGFuY2UgSWRlbnRpdHkgUm9vdCBDQTAe
    Fw0yMDEwMjYxNjU0MTZaFw0yMzEwMjYxNjU0MTZaMCoxKDAmBgNVBAMTH0RpZWdv
    IEluc3RhbmNlIElkZW50aXR5IFJvb3QgQ0EwggEiMA0GCSqGSIb3DQEBAQUAA4IB
    DwAwggEKAoIBAQDO/vzpOgztViaCqKyjfVTayyZX/iGQYaq3mczzHXutBDSMvgU2
    Ig/+sxFFRD8wOTTjCf3gGiOWcLYyZcdkKPEHbeLG8eyGSwLHAhcrl8g/HUK2BT2d
    uVvCwOefARsBuULHGaeO0vIugCitnRS15+BScfWJxYSn8hxjynG3CvYpcjbq1a4O
    0575DbKYXO7e1S662A8n5QjsDb5QXDC2mDuomc5g+UHt11auXv2FUo0AwdRkT2rb
    7GJvvChek+C5WnHU6gs+ivTAbUgGwQA4xun5V8i10LDoFwB8IUq6HSodPrAvnALU
    347KV9rTjCajiBMMt4HhuTJJ9CI4f9/rLcCtAgMBAAGjUzBRMB0GA1UdDgQWBBQW
    3SPVggvcy9UPsCAHs0fAYs6D1jAfBgNVHSMEGDAWgBQW3SPVggvcy9UPsCAHs0fA
    Ys6D1jAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBl5PTtTpkx
    zlxYR1DXAaNYhEvYwVDcINDhJiSKvK6uZnnlsHrl98oyU+9c2yyZtn2eiia/Ngn0
    WEzBNw/hUxXlmdGpCS+3NW0AWaTPyxOT8yrxLkd+DBeEznzHkHI4agszHMIqvrvl
    L+Wpj2XRnl8WVDvMnFfQt3t6vfpVtWoIaBHRyDS8OF0YEV/28FN76D1JzeUss3D3
    MRo0H3ZDLakZow+aGFW3jIhU6+ZJ41J+YBHl4H8lN9L7jkIWYFx257i/iZSz775j
    LzqLdDnRFZZqgPsJA1drkIYyIiAz80+NxxUxR58Hmc90B7AuMOvjZXSFSxGdbccc
    9AM78hx9Ak6c
    -----END CERTIFICATE-----
  private_key: |
    -----BEGIN RSA PRIVATE KEY-----
    MIIEpQIBAAKCAQEAzv786ToM7VYmgqiso31U2ssmV/4hkGGqt5nM8x17rQQ0jL4F
    NiIP/rMRRUQ/MDk04wn94BojlnC2MmXHZCjxB23ixvHshksCxwIXK5fIPx1CtgU9
    nblbwsDnnwEbAblCxxmnjtLyLoAorZ0UtefgUnH1icWEp/IcY8pxtwr2KXI26tWu
    DtOe+Q2ymFzu3tUuutgPJ+UI7A2+UFwwtpg7qJnOYPlB7ddWrl79hVKNAMHUZE9q
    2+xib7woXpPguVpx1OoLPor0wG1IBsEAOMbp+VfItdCw6BcAfCFKuh0qHT6wL5wC
    1N+Oylfa04wmo4gTDLeB4bkySfQiOH/f6y3ArQIDAQABAoIBAQCrAv7vsIX9jq9C
    Qxhd+a2hFTUYfVw9bHMePHKWaEVFK7Q+kr67emi8hDRAhaGutZR7/kVAYFgGchgU
    iwGwPiLjgGVa94PxbwdcYt3BpiRKAGKc/rdpFzo4LCcvtjoZsnT5CLjlxmFPCZKR
    3LS/lFI/yuaQbB6sodnSl+5ayzOUCMvd8rU3LPXVRfu/Jh5FC8916hKhxWG+AWhJ
    ElvL1pyaHiLJRvOG4EAuo4LoJgVmDwqeEJNTk4CIdThpk6oGxgzIjcvxjzQxlVLo
    D3qrFw5+EiK4QExRfJ5LoEWu2DQv5lWtjW+O73pPRpnKUVJoOgpaF8hwa5BYh9bg
    2wUBTyIBAoGBAOulNjKtp7qIItzcweuZztV6brWf+NgmeroXqlP36lYTM6qxqWY9
    84WKrB0vZf6GfPcVTvLOT5Bkz/CrWONqrZYnkNtcHt14lwKWBwDZSWQGonBQjyPi
    aCJw7Gfl9KOmT5MD97j8OseHSCY4k1RqUBDohnqGC8SaokQqaLql7dnPAoGBAODg
    QjC96ICShpf6YuSAvaCoO+tUEHxn55BGo8SyVkEKh74glvv//EtR0+2YTXnX+n9K
    CBpg8lyAYVHheKCWKgrQUwnChVS0bPHvWoOilXWB/JqdrUFJudZwlaLomATpy/xl
    Rb1ONWCvsMwg0fPXuPrIpRBluFGGDUFrDjDwH6jDAoGBAJ9MwKbh3lGrVmYYlr++
    6qRGcDE4Q/Fbkfvbo7nADxrBQFxUXkBQASB17oSMVlcKc9BVB1n9PqxOeoQoUZ7r
    rw2jEbo5PGRb8To+Ud2xBnwoQAfNbfbER8GAtVBHlGpNM94fAIh9ev8H5S5xcKfQ
    du/3QXHyzGHMZ4XNZZ9ILNLhAoGARLN+xVFflNgvEoNGbzT9ufVryOt31eoQjr1m
    DxPE0j4bVnSya+6672/iZTYghVb8iqLdcuGnaac3FELkDXuTAJbAp7yr60Lr/cX4
    SzsCmlHKEJqXcdjKU781l/2jY+zhiwyNj9Yy7IUAaHymZ+7B7qwZ8baB5zYjGpdQ
    UJcrtO0CgYEA1GPNPr3nf405Bc824yx9tkFSwH1j2hYSabkDInOhoK8u8+BJ9uNc
    9W9cPmsAlxCVOoR4w7Ls3nbeW7JDrRrVJ2s7UeVdpTOgEMrbfn39XJ2KHtivlb77
    4QdYTeAXUCGTOsAaDpqT849D6JznedLCGvgJb/1n23CFWIUSCNDnG/w=
    -----END RSA PRIVATE KEY-----
version_created_at: "2020-10-26T17:30:42Z"
`
