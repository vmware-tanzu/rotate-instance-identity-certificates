// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package credhub

// Certificate is a credhub certificate
type Certificate struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Value struct {
		Certificate string `yaml:"certificate,omitempty"`
		PrivateKey  string `yaml:"private_key,omitempty"`
		CA          string `yaml:"ca,omitempty"`
	} `yaml:"value"`
}

func NewCertificate(name, certType, certificate, privateKey, CA string) *Certificate {
	c := &Certificate{
		Name: name,
		Type: certType,
	}
	c.Value.Certificate = certificate
	c.Value.CA = CA
	c.Value.PrivateKey = privateKey
	return c
}
