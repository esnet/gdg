---
title: "version v0.9"
description: "Release Notes for v0.9"
date: 2025-05-15T00:00:00
draft: false
images: [ ]
weight: 9
toc: true
---

## Release Notes for v0.9.0

**Release Date: 12/31/2025**

### Min Recommended Grafana Versions:
 - Grafana 11+

### Breaking Changes
  - [#502](https://github.com/esnet/gdg/pull/502) Password and Tokens in config file has been deprecated.
  - [#510](https://github.com/esnet/gdg/pull/510) Renamed default config to gdg.(yaml|yml)
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
  - [#515](https://github.com/esnet/gdg/pull/515) Add support for Alerting Timings
  - [#519](https://github.com/esnet/gdg/pull/519) Add support for library elements pagination (more than 100 elements)

### Bug/Security Fixes
-

### Developer Changes
-

### Tech Updates
 - [#497](https://github.com/esnet/gdg/pull/497) Refactoring of buildConfigSeachPath
 - [#503](https://github.com/esnet/gdg/pull/503) Ensuring only one CI job run for a PR
 - [#509](https://github.com/esnet/gdg/pull/509) Switching to use OpenAPI endpoint for HeathEndpoint
 - [#512](https://github.com/esnet/gdg/pull/512) Changing naming convention to auth_context

#### Contributors:
  - [PavelsDenisovs](https://github.com/PavelsDenisovs)
  - [safaci2000](https://github.com/safaci2000)
