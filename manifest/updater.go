// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"regexp"
	"strings"
)

type Updater interface {
	useNewIntermediateCert() error
	useNewRootCert() error
	opsmanProductName() string
}

func replaceInvalidNameChars(isoName string) string {
	re := regexp.MustCompile("[-_ ]")
	return strings.ToLower(string(re.ReplaceAllLiteralString(isoName, "_")))
}
