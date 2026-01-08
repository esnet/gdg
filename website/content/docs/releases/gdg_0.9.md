---
title: "version v0.9"
description: "Release Notes for v0.9"
date: 2026-01-11T00:00:00
draft: false
images: [ ]
weight: 1
toc: true
---

## Release Notes for v0.9.0

**Release Date: 01/09/2026**

### Min Recommended Grafana Versions:
 - Grafana 11+

### Release Notes:

This was a pretty big change with it has a good bit of breaking changes. The idea of the `secure` folder was added
a while ago that allowed for separation of sensitive data out of the main config. This released has removed all sensitive
data from the main config. That includes passwords, tokens, cloud keys etc.

for k8s and docker setup this change enable secrets to be mounted from your secret manager and it allows a clear separation between
secrets and config.

Plugin System:

State: Beta (might still have some breaking changes in future releases)

[extism](https://extism.org/) plugin system was incorporated into gdg. The plugin system is disabled by default. There is a performance
hit when using them, so if you don't need them it's better to keep the feature off. The main reason this was added was because
sensitive data is pulled in with the alerting contact points. If you use that feature and store your backups in git it's highly
encouraged to use a cipher plugin.

There are two working currently available at: [gdg-plugins](https://github.com/esnet/gdg-plugins)

1. ansible vault cipher
2. aes-256-gcm

You can also easily write your own. The cipher plugin takes in a string and returns a string. It's likely one of the simplest ones that will be supported.

More detailed docs will be available with the official [docs](https://software.es.net/gdg/docs/gdg/configuration/plugins/).

{{< callout context="caution" title="Caution" icon="alert-triangle" >}}
Not all data in the `secure` folder currently supports encryption. Datasource auth is still in plaintext. The next release should
address this. issued [#524](https://github.com/esnet/gdg/issues/524), will address this change.
{{< /callout >}}

### Breaking Changes
  - [#502](https://github.com/esnet/gdg/pull/502) Password and Tokens in config file has been deprecated.
  - [#510](https://github.com/esnet/gdg/pull/510) Renamed default config to gdg.(yaml|yml).
      importer.yml will still work but a warning will be printed. importer.yml will be dropped in 0.10.x
  - [#513](https://github.com/esnet/gdg/pull/513) Changed location of S3 secure auth.
    - Additionally:
      - `custom`: `true` flag has been removed. It was deprecated 0.8.x. Please use `cloud_type`: `custom`
        instead.
      - Behavioral change. AWS_ACCESS_KEY and AWS_SECRET_KEY will now override config values. This is now consistent with
        how the rest of the GDG config handled env overrides.
  - [#504](https://github.com/esnet/gdg/pull/504) Changing behavior of alert rules. Since they are tied to a given folder, the
     rules will be saved in the given folder. Additionally, folder filtering has been added to allow a user to only
     include rules they are interested in.

### Changes
  - [#520](https://github.com/esnet/gdg/pull/520) Adding plugin support.
  - [#515](https://github.com/esnet/gdg/pull/515) Add support for Alerting Timings
  - [#519](https://github.com/esnet/gdg/pull/519) Add support for library elements pagination (more than 100 elements)

### Bug/Security Fixes
  - [#521](https://github.com/esnet/gdg/pull/521) pnpm security update JS
  -


### Tech Updates
 - [#497](https://github.com/esnet/gdg/pull/497) Refactoring of buildConfigSeachPath
 - [#503](https://github.com/esnet/gdg/pull/503) Ensuring only one CI job run for a PR
 - [#509](https://github.com/esnet/gdg/pull/509) Switching to use OpenAPI endpoint for HeathEndpoint
 - [#512](https://github.com/esnet/gdg/pull/512) Changing naming convention to auth_context

#### Contributors:
  - [PavelsDenisovs](https://github.com/PavelsDenisovs)
  - [safaci2000](https://github.com/safaci2000)
