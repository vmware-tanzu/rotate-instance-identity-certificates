// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

// Command riic rotates the Diego Instance Identity Certificate.
// It is intended to run from the Operations Manager VM.
package main

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/mattn/go-isatty"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/bosh"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/credhub"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/manifest"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/om"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/rotate"
	"github.com/vmware-tanzu/rotate-instance-identity-certificates/validate"
	"golang.org/x/crypto/ssh/terminal"
)

const dateFormat = "01/02/2006"

var Version = "0.0.0-dev"

var cli struct {
	Username             string `short:"u" env:"RIIC_USERNAME" help:"The Operations Manager Username"`
	UseClientSecret      bool   `short:"c" env:"RIIC_USE_CLIENT_SECRET" help:"Use client ID/secret instead of password auth"`
	RunOutsideOpsManager bool   `hidden:"" env:"RIIC_RUN_EXTERNALLY" short:"x" help:"Bypass checks that verify we're running on Operations Manager"`
	Interactive          bool   `short:"i" help:"Set or update required values from the console"`

	Version kong.VersionFlag `short:"v" help:"Show the version and exit"`

	Password             string `kong:"-"`
	DecryptionPassphrase string `kong:"-"`

	CheckExpiry struct{} `cmd:"" help:"Check the certificate expiration date"`
	Rotate      struct {
		StartPhase string `hidden:"" default:"bosh" help:"Specify the starting point (bosh|credhub|apply|cleanup)"`
	} `cmd:"" help:"Perform the certificate rotation"`
	Validate struct{} `cmd:"" help:"Validate that the certs in Credhub match what's deployed to VMs"`
}

var stdin = bufio.NewReader(os.Stdin)

func main() {
	log.SetPrefix("RIIC - ")
	ctx, err := buildContext()
	if err != nil {
		log.Fatal(err)
	}

	if !cli.RunOutsideOpsManager {
		if _, err := os.Stat("/var/tempest/workspaces"); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "This tool must run on the Operations Manager VM.\n")
			os.Exit(1)
		}
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	om := om.NewAPI("https://127.0.0.1", cli.Username, cli.Password, cli.DecryptionPassphrase, cli.UseClientSecret, client)
	err = ValidateVersion(om)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	env, err := om.GetDirectorCredentials()
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not get director credentials: %v", err)
		os.Exit(1)
	}
	env = append(env, os.Environ()...)

	boshRunner := bosh.NewRunner(env)
	credhubRunner := credhub.NewRunner(env)
	manifestLoader := manifest.NewLoader(om, boshRunner)
	certExpirationValidator := validate.NewCertExpiration(credhubRunner)
	diegoValidator := validate.NewDiego(boshRunner, credhubRunner)
	routerValidator := validate.NewRouter(boshRunner, credhubRunner)

	switch ctx.Command() {
	case "check-expiry":
		expired := func(name string, t time.Time) bool {
			if time.Now().After(t) {
				fmt.Printf("❌ %s cert expired on: %s\n", name, t.Format(dateFormat))
				return true
			}
			return false
		}
		check := func(name string, t time.Time) {
			timeLeft := time.Until(t)
			days := int(timeLeft.Hours() / 24)
			s := "✅"
			if days <= 90 {
				s = "⚠️"
			}
			fmt.Printf("%s %s cert is valid for %d more days, it expires on: %s\n",
				s, name, days, t.Format(dateFormat))
		}

		// check the shared root CA cert
		rootExpiration, err := certExpirationValidator.CheckRootCertExpiration()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
		if !expired("Root", rootExpiration) {
			check("Root", rootExpiration)
		}

		// check each deployment's intermediate cert
		manifests, err := manifestLoader.GetAllManifestsWithDiegoCells()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		for _, m := range manifests {
			intermediateExpiration, err := certExpirationValidator.CheckIntermediateCertExpiration(&m)
			if err != nil {
				fmt.Fprintf(os.Stderr, "could not check expiration date: %v\n", err)
				os.Exit(1)
			}
			n := fmt.Sprintf("%s intermediate", m.DeploymentName)
			if !expired(n, intermediateExpiration) {
				check(n, intermediateExpiration)
			}
		}

	case "rotate":
		printBanner()
		u, err := user.Current()
		if err != nil {
			fmt.Fprintf(os.Stderr, "could not get the current user: %v\n", err)
			os.Exit(1)
		}
		if u.Username != "tempest-web" {
			fmt.Fprintf(os.Stderr, "Cannot proceed as %s, expected to be running under tempest-web user\n", u.Username)
			os.Exit(1)
		}

		rotator := rotate.NewCertRotator(om, boshRunner, credhubRunner, manifestLoader, diegoValidator, routerValidator)
		err = rotator.RotateCerts(cli.Rotate.StartPhase)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Rotation Failed, exiting due to error: %s\n", err)
			os.Exit(1)
		}
		fmt.Print("\n\nFinished rotating certs\n\n")

	case "validate":
		manifests, err := manifestLoader.GetAllManifestsWithDiegoCells()
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}

		for _, m := range manifests {
			err = diegoValidator.ValidateCerts(&m, validate.AllInstancesFilter)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
			err = routerValidator.ValidateCerts(&m, validate.AllInstancesFilter)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s\n", err)
				os.Exit(1)
			}
		}
	}
}

func ValidateVersion(om *om.API) error {
	cfVersion, err := om.GetDeployedProductVersion("cf")
	if err != nil {
		return fmt.Errorf("couldn't check cf version: %w", err)
	}

	parts := strings.Split(cfVersion, ".")

	if parts[0] != "2" {
		return fmt.Errorf("invalid cf version %s", cfVersion)
	}

	switch parts[1] {
	case "4", "5", "6":
		return nil
	default:
		return fmt.Errorf("invalid cf version %s", cfVersion)
	}
}

func buildContext() (*kong.Context, error) {
	ctx := kong.Parse(&cli,
		kong.Name("riic"),
		kong.Description(`
riic rotates Diego instance-identity certificates

To run non-interactively set the $RIIC_PASSWORD and $RIIC_DECRYPTION_PASSPHRASE environment variables
`),
		kong.Vars{
			"version": Version,
		},
	)

	inTTY := canBeInteractive()
	if !inTTY && cli.Interactive {
		return nil, errors.New("cannot specify --interactive and run outside of a TTY (i.e. via nohup)")
	}

	cli.Password = os.Getenv("RIIC_PASSWORD")
	cli.DecryptionPassphrase = os.Getenv("RIIC_DECRYPTION_PASSPHRASE")

	var err error
	if cli.Username == "" || cli.Interactive {
		if err = handleInput("Operations Manager Username", "--username", "RIIC_USERNAME", &cli.Username, inTTY, false, true); err != nil {
			return nil, fmt.Errorf("could not get username from console: %w", err)
		}
	}

	if cli.Password == "" || cli.Interactive {
		if err = handleInput("Operations Manager Password", "no flag available", "RIIC_PASSWORD", &cli.Password, inTTY, true, true); err != nil {
			return nil, fmt.Errorf("could not get password from console: %w", err)
		}
	}

	if cli.DecryptionPassphrase == "" || cli.Interactive {
		if err = handleInput("Operations Manager Decryption Passphrase", "no flag available", "RIIC_DECRYPTION_PASSPHRASE", &cli.DecryptionPassphrase, inTTY, true, false); err != nil {
			return nil, fmt.Errorf("could not get decryption passphrase from console: %w", err)
		}
	}

	return ctx, nil
}

func handleInput(prompt string, flag string, envar string, value *string, inTTY bool, isSecret bool, isRequired bool) error {
	if value == nil {
		return errors.New("cannot pass a nil value pointer")
	}

	if !inTTY && (*value == "") {
		if isRequired {
			return fmt.Errorf("the %s is not set and cannot be input interactively. Please make sure to set this value via the appropriate flag (%s) or environment variable (%s)", prompt, flag, envar)
		}

		return nil
	}

	showDefVal := *value
	if isSecret && showDefVal != "" {
		showDefVal = "*****"
	}

	var (
		inputVal []byte
		err      error
	)

	for {
		fmt.Printf("%s (ENTER to accept current value) [%s]: ", prompt, showDefVal)

		if isSecret {
			inputVal, err = terminal.ReadPassword(int(os.Stdin.Fd()))
			fmt.Println()
		} else {
			inputVal, err = stdin.ReadBytes('\n')
			if err == nil {
				inputVal = inputVal[0 : len(inputVal)-1] // strip the newline that terminal.ReadPassword doesn't inclue
			}
		}

		if err != nil {
			return fmt.Errorf("could not accept input value: %w", err)
		}

		if len(inputVal) > 0 {
			*value = string(inputVal)
		}

		if isRequired && strings.TrimSpace(*value) == "" {
			fmt.Println("This field is required and must not be blank. Please try again.")
			continue
		}

		break
	}

	return nil
}

func canBeInteractive() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

const banner = `       O O
  _ __ _ _  ___
 | '__| | |/ __|
 | |  | | | (__
 |_|  |_|_|\___|
`

func printBanner() {
	fmt.Println(banner)
}
