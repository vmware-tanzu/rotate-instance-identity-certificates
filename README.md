# rotate-instance-identity-certificates

Tooling to rotate the Diego Instance Identity Certificate Authorities for Tanzu
Application Service (TAS) 2.4, 2.5, and 2.6.

Running TAS and think this might help? Check out the
[user docs site](https://vmware-tanzu.github.io/rotate-instance-identity-certificates).

## For Developers

This project uses Go 1.15+ and Go modules. Clone the repo to any directory.

`go test ./...` to run the tests.

Build the project by running `go build` in the project root.

To cross compile, set the `GOOS` and `GOARCH` environment variables.

For example, to compile a Linux binary run:

```
$ GOOS=linux GOARCH=amd64 go build
```

### Releases

You can view release notes and download artifacts from the project's
[releases](https://github.com/vmware-tanzu/rotate-instance-identity-certificates/releases)
page.

To create a release, simply create and push a new semver tag starting with `v`:

```
$ git tag v0.2.0
$ git push --tags
```

A Github Action will pick up this tag and create a draft release. Navigate to
the releases page, find the draft, enter any release notes, and publish!

### Measuring Uptime

Since one of the goals for this tool is to be able to rotate certificates without
downtime, we must prove that any changes made to the tool don't negatively
impact running apps or the cloud controller API.

To measure the uptime of these components we use a tool called
[Uptimer](https://github.com/cloudfoundry/uptimer). Since this tool must be run
on the Operations Manager VM you'll need to run uptimer from another machine. To
do this likely need to run this tool separately from uptimer and use uptimer
with a configuration that waits.

```json
{
  "while": [{
    "command": "sleep",
    "command_args": ["7200"]
    ...
```

Windows diego cell support is in the works for measuring application uptime.
