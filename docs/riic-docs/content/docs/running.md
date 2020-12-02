---
title: "Running the Tool"
date: 2020-10-29T08:14:54-06:00
draft: false
weight: 200
---

This tool will automatically rotate/update the older Diego root certificate
authority and Intermediate Identity Cert used by Diego for application instance
identity. This tool is not needed if you have already upgraded to Tanzu
Application Service 2.7 or later, as that version automatically rotates those
certs on upgrade.

The tool operates on all BOSH deployments that have Diego cells, including TAS,
TASW, and Isolation Segments for TAS.

{{% notice info %}}
Note: we only support the full TAS deployment - small
footprint runtime is not currently supported.
{{% /notice %}}

The rotation process is a 2 step process:

1. Add a new secondary root CA and intermediate identity certs. Deploy these new
   certs directly using a BOSH deploy against all deployments with Diego cells.
2. Swap the new certs into place of the original certs in Credhub and then do a
   selective Apply Changes through Operations Manager for each product with
   Diego cells.

At the end of each stage the tool will select a Diego cell and a router from
each deployment and validate the certs match what is in Credhub.

# Execute the rotation

## Prerequisites

Download the latest binary from the [releases](https://github.com/vmware-tanzu/rotate-instance-identity-certificates/releases) page. This is a linux
binary that will only run on the Operations Manager VM. Copy the binary to your
Operations Manager VM:

```bash
$ scp ~/Downloads/riic -i ops-manager-private-key.pem ubuntu@ops-manager-fqdn:~/riic
```

We recommend running the tool using `nohup`. This is important because
if you run a rotate command directly from your SSH session and your connection
terminates, so does the rotation process, potentially leaving your foundation in
an indeterminite state. To start the tool in this way run the following:

```bash
$ nohup riic <subcommand> &
```

This keeps the the command running even if your SSH session terminates. The
standard error and standard out are viewable in the nohup.out log file:

```bash
$ tail -f nohup.out
```

## Passing credentials

There are no flags to pass in secret values so that they will not appear in your
shell history. You will be prompted to enter the Operations Manager admin
user's/client's password. If you prefer this to be non-interactive, set the
`RIIC_PASSWORD` environment variable.

## SAML Auth

If your Operations Manager does not use username/password authentication, simply
set the `--use-client-secret` (`-c`) flag, and put your client ID in the
`--username` argument and your client secret as the password (see above for
an explanation of password handling).

## Diego Identity Cert Expiration Check

Before attempting to rotate your instance identity certificates, it's a good
idea to check their expiration dates. This can easily be done using the
`riic check-expiry` command.

From your Operations Manager VM, run the following command, passing in your
Operations Manager credentials.

```bash
$ ./riic --username admin check-expiry
```

This will display the Diego CA and Intermediate Identity Cert expiration dates
along with a warning if they're close to expiring.

## Diego Identity Cert Rotation

Once you've checked the expiration dates and are ready to rotate certificates,
we recommend that you first ensure that all VMs in your TAS, isolation segment,
and TASW deployments are healthy. You can do this by looking at the _status_
tab for each tile in Operations Manager, or from the bosh CLI:

```bash
$ bosh -d DEPLOYMENT_NAME vms
```

If there are VMs that are unhealthy, resolve those issues _before_ proceeding with
the rotation. This may require you to recreate failing VMs.

When you've confirmed that the deployment is healthy, proceed to rotate the
certificates.

_NOTE_ - When rotating certificates, the tool must run as the `tempest-web` user
to ensure there are no permission errors during the deployment. Make sure to
switch to the correct user before proceeding with the rotation:

```bash
$ sudo su - tempest-web
```

If you forget this step, the operation will fail and ask you to switch users
before trying again.

Once you're running as `tempest-web`, run the following command to begin the
rotation:

{{% notice tip %}}
The following examples run the rotate command using `nohup` as noted above.
{{% /notice %}}

```bash
$ nohup ./riic --username admin rotate &
$ tail -f nohup.out
```

This will rotate the Diego CA and Intermediate Identity certs and update all
BOSH jobs that reference these certs.

## Diego Identity Cert Validation

As an additional check you can run the validate command when the rotation
completes. This will check all the certs on the diego cells and routers. Unlike
the validation done at the end of rotate, this will be done on _every_ single
VM, not just the first. We don't recommend running this on very large
deployments.

```bash
$ ./riic --username admin validate
```
