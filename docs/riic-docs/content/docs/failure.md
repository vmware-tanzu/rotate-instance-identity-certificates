---
title: "Handling Failures"
date: 2021-01-19T10:06:00-07:00
draft: false
weight: 300
---

In some cases deployments may fail, leaving the certificate rotation in a partially
completed state. It's safe to re-run riic from the beginning after you've addressed
whatever failure may have occurred. In some situations you may not want to wait for
all of the riic steps to re-run. In these scenarios, you can have it pick up at a
specific step using a purposely hidden command-line flag.

{{% notice warning %}}
Starting at the wrong step can cause a broken deployment and downtime for applications.
{{% /notice %}}

To start at a specific step, run the tool with the `--start-phase` argument. The
default start phase is `bosh` - which starts by doing a direct bosh deploy of the
freshly-generated certificates.

For example:

```
$ riic --username=admin rotate --start-phase=credhub
```

Valid start phases are:

- `bosh`: start by modifying the manifests and performing a BOSH deployment
- `credhub`: start by updating the original certs in Credhub with the newly generated ones
- `apply`: start with the Operations Manager apply changes step
- `cleanup`: simply remove the duplicate certificate reference from Credhub and exit

Specifying anything other than one of these values is equivalent to passing
`bosh`, which will start the process from the beginning. If you're unsure of
which step to start at, it's always safe to start at the first step `bosh`.

{{% notice note %}}
Rerunning successful steps is safe. Skipping steps that have failed is _not_ safe.
{{% /notice %}}

Any failure past the initial bosh deploy step should generally be started at the
`credhub` step. You can tell if the bosh deploy step succeeded by looking for the
log message "Rotating identity certs in Credhub", if you see that you know you've
succesfully gotten past the bosh step.

You may need to start at the cleanup step in cases where you've decided to apply changes directly via Operations Manager.
