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
