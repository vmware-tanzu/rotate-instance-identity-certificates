// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package validate

import (
	"testing"
	"time"
)

const certificate = `-----BEGIN CERTIFICATE-----
MIIDTTCCAjWgAwIBAgIUZRXlSOK+t7BSp10Knjic/K1J++YwDQYJKoZIhvcNAQEL
BQAwKjEoMCYGA1UEAxMfRGllZ28gSW5zdGFuY2UgSWRlbnRpdHkgUm9vdCBDQTAe
Fw0yMDA5MTYyMTIyMzlaFw0yMjA5MTYyMTIyMzlaMDIxMDAuBgNVBAMTJ0RpZWdv
IEluc3RhbmNlIElkZW50aXR5IEludGVybWVkaWF0ZSBDQTCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBALTpovpgQALuVoaiFk/ULivM106S7MclTilnIeoq
JtlSOxSz2PMARQMWYhYs4Hux6ICABsMvxUyeldItNvXIUC3Rc9KlnHHF/O284qjG
blBwuOxuXaAQmnH9FdtZxC34lgrZDIozRGslRGY9KQFmp8QdBcTbY7jLNDEM/wR3
o5gxQry8X4oCLyzqq+TYWuB2HJDfa3vZWA/6N9TLDE6fmPC3OWbZxBNl5n2cVPgb
vlfOBz6e7x/zibGCJ3fL5ziCvsuh7hRny9NC3PuH2AkDtnx+bHdOkzFb0ydBOFxN
bU03JZ6z88olYYWcn9luyHCei2uC5uKfGzLUBtjaDg/7CekCAwEAAaNjMGEwHQYD
VR0OBBYEFDFS2cKdvZv7+79EE01vJrPWSOMTMA4GA1UdDwEB/wQEAwICBDAfBgNV
HSMEGDAWgBStC6JGt9uuXLfPae86vkcnxZ7VujAPBgNVHRMBAf8EBTADAQH/MA0G
CSqGSIb3DQEBCwUAA4IBAQC0N7dEQK/q78yVe143aNsRl68ToKo+JRw/XnxS4KRY
Ds5U9A9DV5cb4lvVSiXbHyILyiPMKwHIKSBNvrz99AMak+ji/obEZcyc9tb6quzQ
3l+ItKuhHmNC2MEXQrpm4PbFq8YBvTA+5zzhQpf/5HaQdlXDmS/o9krCX/X1yGGm
8wcrODy11aeVhVNc827GP88sG1rPW95R6aTNF3M3iZhGK64TQoLWF4RbnDzVDsXY
RfKXlnXLh9e1vBUYOELTtqcnlRuqQgL9o7n1fvE0xybKypBq+nosYQPnQ6ObnCZ2
lObshT64lGwUPYqX/8N/8VIQKEMrqD+an5kf6nB4wTz7
-----END CERTIFICATE-----
`

func TestGetExpiration(t *testing.T) {
	expiration, err := getExpiration(certificate)
	if err != nil {
		t.Error(err)
	}

	year := expiration.Year()
	month := expiration.Month()
	day := expiration.Day()
	if year != 2022 || month != time.September || day != 16 {
		t.Errorf("Expected cert to expire on 9/16/22, but it expires on %s", expiration)
	}
}
