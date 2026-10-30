---
title: "Structure"
description: "Overview of GDG configuration file structure, YAML layout, environment variable overrides, and config file discovery paths."
weight: 100
---

GDG Configuration is based on a YAML configuration.  It also has the option to override configuration values via environmental variables.

The default config file is gdg.yml which should exist in $CWD/config, $CWD or /etc/gdg.  If no valid file can be found, the application will fail.  An example config file can be found in [github](https://github.com/esnet/gdg/blob/main/config/gdg-example.yml) under the config folder.

The configuration has a few sections of note:
1. context_name: the selected context
2. Storage -- defines the storage engine if a cloud provider is used
3. contexts: defines every context you operate on, think of it as the definition of a grafana instance and how to connect to it.  It also defines behavior for what folders to inspect, authentication mechanism, etc.
4. global: defines global behavior for all contexts.  These settings are not context specific.  They tend to be things like enabling debugging, retry logic, api debugging etc.

## Environment Overrides

{{% notice style="warning" title="Caution" %}}
Complex data type is not supported. If the value is an array it can't be currently set, via ENV overrides.

{{% /notice %}}

If you wish to override certain value via the environment, like credentials and such you can do so.

The pattern for GDG's is as follows:  `GDG_SECTION__SECTION__keyname`

For example if I want to set the context name to a different value I can use:

```sh
GDG_CONTEXT_NAME="testing" gdg tools ctx show ## Which will override the value from the context file.
GDG_CONTEXTS__TESTING__URL="www.google.com" Will override the URL with the one provided.
 ```

## Supported Environment Variables

### Grafana Authentication (v0.10.0+: env always takes precedence)

| Environment Variable | Description |
|---|---|
| `GDG_CONTEXTS__<CTX>__TOKEN` | API token / service account token for context `<CTX>` |
| `GDG_CONTEXTS__<CTX>__PASSWORD` | Basic auth password for context `<CTX>` |

`<CTX>` must be the context name in **upper case**, e.g. `TESTING`, `PRODUCTION`.

These variables are always evaluated and **win over the secure auth file** on a
per-field basis. See [Getting Started](../getting_started) for full auth precedence details.

### Cloud / Custom S3 Storage Credentials (v0.10.0+)

| Environment Variable | Description |
|---|---|
| `GDG_STORAGE_ENGINE__<STORAGE>__ACCESS_ID` | Access ID for custom S3-compatible storage named `<STORAGE>` |
| `GDG_STORAGE_ENGINE__<STORAGE>__SECRET_KEY` | Secret key for custom S3-compatible storage named `<STORAGE>` |

`<STORAGE>` must be the storage name in **upper case**, matching the key under
`storage_engine` in your `gdg.yml`.

```sh
# Example: storage named "super_cloud"
export GDG_STORAGE_ENGINE__SUPER_CLOUD__ACCESS_ID=myaccesskey
export GDG_STORAGE_ENGINE__SUPER_CLOUD__SECRET_KEY=mysecretkey
```

These apply only to `cloud_type: custom` (MinIO / custom S3-compatible endpoints).
Native cloud providers (AWS S3, GCS, Azure Blob) use their own SDK credential
chains and are not affected.

Standard AWS SDK variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`) for native AWS S3 are unaffected by GDG and continue to work as expected by the AWS SDK.
