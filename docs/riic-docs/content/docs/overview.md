---
title: "Overview"
date: 2020-11-02T13:30:27-07:00
draft: false
weight: 100
---

## Intro

Tanzu Application Service (TAS) includes an internal certificate authority that
is used to sign instance identity certificates. These certificates are issued to
each application instance and are used to verify application identity for
routing and
[other features](https://tanzu.vmware.com/content/blog/new-in-pcf-2-1-app-container-identity-assurance-via-automatic-cert-rotation).

If these certificates expire, the platform will be unable to verify application
identity and route requests to the correct applications, resulting in downtime
for applications that are accessed through the platform routers.

These certificate authorities are not renewable via API.

{{% notice note %}}
The _preferred_ approach to ensuring that these certificates do not expire is
to upgrade TAS to version 2.7 or later, which has the effect of both renewing
the certificates and making them rotatable via common APIs. This tool is a
good option if you cannot complete the upgrade process prior to the certificate
expiration date.
{{% /notice %}}

## Procedure

The tool orchestrates a 3-step rotation process which was designed to ensure
that the platform and applications do not experience connectivity issues during
the rotation.

1. Generate and deploy new certificates (side by side with existing certificates)
1. Overwrite expiring certificates with new ones and redeploy
1. Cleanup

This rotation results in two deployments, so the total amount of time required
depends on the size of the environment. The first deployment is performed
directly with BOSH, which creates a temporary discrepancy between the BOSH state
and the Operations Manager state. The final deployment is performed via
Operations Manager in order to bring it back in sync with BOSH.

Applications that are deployed with more than one instance should not expect an
interruption in service during the rotation.
