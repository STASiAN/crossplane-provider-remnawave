# crossplane-provider-remnawave

`crossplane-provider-remnawave` is a [Crossplane](https://crossplane.io/) provider
for the [Remnawave](https://remna.st/) proxy management panel. It is built with
[Upjet](https://github.com/crossplane/upjet) code generation from the
[batonogov/remnawave](https://registry.terraform.io/providers/batonogov/remnawave/latest)
Terraform provider (v1.4.0) and exposes XRM-conformant managed resources for the
Remnawave API.

## Installation

Install the provider with Crossplane v2.x:

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: crossplane-provider-remnawave
spec:
  package: ghcr.io/stasian/crossplane-provider-remnawave:v0.1.0
```

## Configuration

Create a credentials secret and a `ProviderConfig`. The credentials JSON keys
mirror the Terraform provider attributes: `endpoint`, `api_token`, `username`,
`password`, `insecure_skip_verify`, `proxy_headers`, `custom_headers`,
`request_timeout`. `api_token` takes precedence over `username`/`password`.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: remnawave-creds
  namespace: crossplane-system
type: Opaque
stringData:
  credentials: |
    {
      "endpoint": "https://panel.example.com",
      "api_token": "<your API token>"
    }
---
apiVersion: remnawave.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: remnawave-creds
      namespace: crossplane-system
      key: credentials
```

Then create managed resources, e.g.:

```yaml
apiVersion: remnawave.crossplane.io/v1alpha1
kind: User
metadata:
  name: john-doe
spec:
  forProvider:
    username: john-doe
    trafficLimitBytes: 10737418240
    trafficLimitStrategy: MONTH
  providerConfigRef:
    name: default
```

See [`examples/`](examples/) for more.

## Resource coverage

All kinds live in the `remnawave.crossplane.io` group (cluster-scoped, legacy)
and `remnawave.m.crossplane.io` (namespaced, Crossplane v2 style):

| Kind | Terraform resource |
| --- | --- |
| `User` | `remnawave_user` |
| `Node` | `remnawave_node` |
| `Host` | `remnawave_host` |
| `ConfigProfile` | `remnawave_config_profile` |
| `ExternalSquad` | `remnawave_external_squad` |
| `InternalSquad` | `remnawave_internal_squad` |
| `SubscriptionTemplate` | `remnawave_subscription_template` |
| `SubscriptionSettings` | `remnawave_subscription_settings` |
| `SubpageConfig` | `remnawave_subpage_config` |
| `PanelSettings` | `remnawave_panel_settings` |
| `Snippet` | `remnawave_snippet` |
| `NodePlugin` | `remnawave_node_plugin` |
| `InfraProvider` | `remnawave_infra_provider` |
| `BillingNode` | `remnawave_billing_node` |
| `BillingHistory` | `remnawave_billing_history` |
| `APIToken` | `remnawave_api_token` |
| `Passkey` | `remnawave_passkey` |
| `UserMetadata` | `remnawave_user_metadata` |
| `NodeMetadata` | `remnawave_node_metadata` |
| `HwidDevice` | `remnawave_hwid_device` |

The imperative Terraform resources (`remnawave_node_action`,
`remnawave_user_action`, `remnawave_host_bulk_action`,
`remnawave_user_bulk_action`, `remnawave_node_bulk_action`,
`remnawave_drop_connections`) are intentionally not generated: one-shot actions
do not fit Crossplane's declarative reconcile loop.

## Development

```bash
make submodules   # initialize the build submodule
make generate     # regenerate APIs/controllers from the Terraform schema
make build        # build binaries and packages
make run          # run the controller out-of-cluster against $KUBECONFIG
```

The provider runs the wrapped Terraform provider via the Terraform CLI
(TF 1.5.7, plugin-framework compatible). Embedding the provider in-process
(upjet no-fork mode) is a possible future optimization.

## Releasing

Push a `v*` tag; the [`publish`](.github/workflows/publish.yml) workflow builds
the multi-arch package and pushes it to
`ghcr.io/stasian/crossplane-provider-remnawave`.

## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/STASiAN/crossplane-provider-remnawave/issues).

## Licensing

crossplane-provider-remnawave is under the Apache 2.0 license.
