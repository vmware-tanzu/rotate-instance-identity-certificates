---
title: "Frequently Asked Questions"
date: 2020-11-02T13:42:03-07:00
draft: false
weight: 999
---

## Why must the tool run on the Operations Manager VM?

Under the hood, we rely on the `bosh` and `credhub` CLIs. These programs are
already guaranteed to be present on the Operations Manager VM and are known to
be compatible with the currently deployed version of the platform, so we require
the tooling to run in this environment to prevent errors due to incompatible
versions or missing tools.

## How long does the rotation take?

The rotation consists of 2 deployments, each of which must update the following instance groups:

- Diego Cells
- Routers
- Diego Brain
- Credhub

The number of cells and routers vary widely across installations, but you can
expect that an environment with a large number of cells and routers will take
much longer to rotate than a smaller environment.

## Should I expect downtime during the rotation?

The process is designed to maintain platform functionality.

The diego cells and routers do need to be updated during the rotation, so you
should ensure that any applications that cannot tolerate downtime are deployed
with more than one instance, and that your load balancer health checks are
correctly configured to detect unhealthy routers.

## Does this tool support the Small Footprint Runtime?

No.

## Does this tool support isolation segments or TAS for Windows?

Yes. The tool will automatically detect whether isolation segments or Windows
cells are installed, and will perform the appropriate actions to rotate the
certificates on these environments as well.

## Any tips for dealing with the verbosity of the logs?

Yes. Output coming directly from the tool (and not from BOSH) is prefixed with "RIIC"
for easy grepping:

```
$ grep RIIC nohup.out
```

## Why does riic say there are pending changes?

It's important that any pending changes staged in Operations Manager are applied or reverted before attempting to rotate certificates with riic. Doing otherwise may lead to downtime. You can check for pending changes in the Operations Manager UI, however some changes like stemcell updates won't show up here.

To find all pending changes open the URL `https://<your-opsman.fqdn>/api/v0/staged/pending_changes` from your logged in browser. In the json response look for any products that do not have an `action` of `unchanged`. That should give a clue as to what changes are actually pending and keeping riic from running. For example a stemcell update would show up with an `action` of `update`.

## Can I restart a cert rotation at a specific step?

In some cases deployments may fail, leaving the rotation in a partially
completed state. In these scenarios, you may need to re-run the tool and have it
pick up at a specific step.

There is a hidden command-line flag that can be used to do this, but be careful.
The flag is hidden intentionally as starting at the wrong step can cause a broken
deployment and downtime for applications.

To start at a specific step, run the tool with the `--start-phase` argument. The
default start phase is `bosh` - which starts by creating a new BOSH deployment
with freshly-generated certificates.

For example:

```
$ riic --username=admin rotate --start-phase=credhub
```

Valid start phases are:

- `bosh`: start by modifying the manifests and performing a BOSH deployment
- `credhub`: start by updating the original certs in Credhub with the newly generated ones
- `apply`: start with the Operations Manager apply changes step
- `cleanup`: simply remove the duplcate certificate reference from Credhub and exit

Specifying anything other than one of these values is equivalent to passing
`bosh`, which will start the process from the beginning.
