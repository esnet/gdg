---
title: "Current Release Notes"
description: "Release Notes for Current Version"
date: 2023-03-31T15:21:01+02:00
lastmod: 2023-04-14T19:25:12+02:00
draft: true
images: []
weight: 199
toc: true
---

##  Release Notes for v0.5.0

**Release Date: TBD 07/13/2023**


### Changes
  - Adding support for Basic CRU for Orgs
  - Renamed 'DataSources' command to 'Connections' to match Grafana's naming convention.
  - Connection Permissions are now supported.  This is an enterprise features and will only function if you have an enterprise version of grafana.  Enterprise features are enabled by setting `enterprise_support: true` for a given context.


### Bug Fixes
  -


### Breaking Changes
  - datasources have been renamed as connections.  If you have an existing backup, simply rename the folder to 'connections' and everything should continue working.
  - All commands have now been moved under 'backup' or 'tools' to better reflect their functionality.

