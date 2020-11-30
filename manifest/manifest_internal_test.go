// Copyright 2020 VMware, Inc.
// SPDC-License-Identifier: Apache-2.0

package manifest

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestCloneVariable(t *testing.T) {
	m, err := NewManifest("p-bosh", "testdata/cf-manifest.yml")
	if err != nil {
		t.Fatal(err)
	}

	l := len(m.Content["variables"].([]interface{}))

	_, err = m.cloneVariable("autoscale-db-credentials", "new-autoscale-db-credentials")
	if err != nil {
		t.Fatal(err)
	}

	if l2 := len(m.Content["variables"].([]interface{})); l2 != l+1 {
		t.Errorf("expected updated manifest to have %d variables but found %d", l+1, l2)
	}

	found := false
	for _, v := range m.Content["variables"].([]interface{}) {
		name := v.(map[interface{}]interface{})["name"]
		if name == "new-autoscale-db-credentials" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("did not find new variable")
	}
}

func TestVerifyManifest(t *testing.T) {
	om, err := exec.LookPath("om")
	if err != nil {
		t.Skip("[WARN] Skipping manifest validation test as om is not installed.")
		return
	}

	withIntermediate, err := ioutil.TempFile("", "cf-with-intermediate-*.yml")
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		withIntermediate.Close()
		os.Remove(withIntermediate.Name())
	})

	m, err := NewManifest("p-bosh", "testdata/cf-manifest.yml")
	if err != nil {
		t.Fatal(err)
	}

	if err := m.Update(withIntermediate); err != nil {
		t.Fatal(err)
	}

	withIntermediate.Close()

	getPath := func(file, path string) ([]byte, error) {
		t.Helper()
		cmd := exec.Command(om, "interpolate",
			"--config", file,
			"--path", path,
			"--skip-missing",
		)
		return cmd.CombinedOutput()
	}

	// note: this test won't catch all mistakes, as some of these paths require
	// the new variable to be added as a new array element, and others just expect
	// string concatenation - we're simply checking for the presence of the variable
	// at the right location here
	t.Run("root", func(t *testing.T) {
		checkPath := func(path string) {
			t.Helper()
			out, err := getPath(withIntermediate.Name(), path)
			if err != nil {
				t.Errorf("couldn't check path %s: %v\n%s", path, err, out)
				return
			}

			if !bytes.Contains(out, []byte(RootCertRegenVariable)) {
				t.Errorf("expected path %v to contain root regen var", path)
			}
		}
		checkPath("/instance_groups/name=router/jobs/name=gorouter/properties/router/ca_certs")
		checkPath("/instance_groups/name=credhub/jobs/name=credhub/properties/credhub/authentication/mutual_tls/trusted_cas")
		checkPath("/instance_groups/name=diego_cell/jobs/name=rep/properties/containers/trusted_ca_certificates")
		checkPath("/instance_groups/name=diego_cell/jobs/name=cflinuxfs2-rootfs-setup/properties/cflinuxfs2-rootfs/trusted_certs")
		checkPath("/instance_groups/name=diego_cell/jobs/name=cflinuxfs3-rootfs-setup/properties/cflinuxfs3-rootfs/trusted_certs")
		checkPath("/instance_groups/name=diego_brain/jobs/name=ssh_proxy/properties/diego/ssh_proxy/bbs/ca_cert")

		proxyEnabledTLS, err := getPath(withIntermediate.Name(), "/instance_groups/name=diego_brain/jobs/name=ssh_proxy/properties/backends/tls/enabled")
		if err == nil && strings.TrimSpace(string(proxyEnabledTLS)) != "false" {
			checkPath("/instance_groups/name=diego_brain/jobs/name=ssh_proxy/properties/backends/tls/ca_certificates")
		}
	})

	t.Run("intermediate", func(t *testing.T) {
		check := func(path, want string) {
			t.Helper()
			b, err := getPath(withIntermediate.Name(), path)
			if err != nil {
				t.Error(err)
				return
			}
			if s := strings.TrimSpace(string(b)); s != want {
				t.Errorf("expected %q at path %v, got %q", want, path, s)
			}
		}

		check("/instance_groups/name=diego_cell/jobs/name=rep/properties/diego/executor/instance_identity_ca_cert", IntermediateCertRegenVariable)
		check("/instance_groups/name=diego_cell/jobs/name=rep/properties/diego/executor/instance_identity_key", IntermediatePrivateKeyRegenVariable)

		check("/variables/name="+IntermediateCertRegenName+"/ca", RootCertRegenName)
		check("/variables/name="+IntermediateCertRegenName+"/options/ca", RootCertRegenName)
	})

}
