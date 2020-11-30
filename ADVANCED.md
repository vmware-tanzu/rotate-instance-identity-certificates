# Advanced

## Starting at a specific step

In some cases deployments may fail, leaving the rotation in a partially
completed state. In these scenarios, you may need to re-run the tool and have it
pick up at a specific step.

There is a hidden command-line flag that can be used to do this, but be careful.
The flag is hidden intentionally as starting at the wrong step can cause a
broken deployment and downtime for applications.

To start at a specific step, run the tool with the `--start-phase` argument. The
default start phase is `bosh` - which starts by creating a new BOSH deployment
with freshly-generated certifictes.

For example:

```
$ riic --username=admin rotate --start-phase=credhub
```

Valid start phases are:

- `bosh`: start by modifying the manifests and performing a BOSH deployment
- `credhub`: start by updating the original certs in Credhub with the newly
  generated ones
- `apply`: start with the Operations Manager apply changes step
- `cleanup`: simply remove the duplcate certificate reference from Credhub and
  exit

Specifying anything other than one of these values is equivalent to passing
`bosh`, which will start the process from the beginning.
