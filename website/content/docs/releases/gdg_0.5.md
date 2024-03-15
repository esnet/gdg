---
title: "version v0.5"
description: "Release Notes for v0.5"
date: 2023-09-01T00:00:00
draft: false
images: []
weight: 198
toc: true
---

##  Release Notes for v0.5.2
### Changes
- [#229](https://github.com/esnet/gdg/issues/229) Datasource auth has been moved to a file based configuration under secure/.  This allows for any number of secure values to be passed in.  Using the wizard for initial config is recommended, or see test data for some examples.
  - [#168](https://github.com/esnet/gdg/issues/168) Introduced a new tool called gdg-generate which allows for templating of dashboards using go.tmpl syntax.
  - gdg context has been moved under tools.  ie. `gdg tools ctx` instead of `gdg ctx`
  - [#221](https://github.com/esnet/gdg/issues/221) Version check no longer requires a valid configuration
  - [#236](https://github.com/esnet/gdg/issues/236) Dashboard filter by tag support.  Allows a user to only list,delete,upload dashboards that match a set of given tags.

### Bug Fixes
- [#235](https://github.com/esnet/gdg/issues/235) Fixed a bug that prevented proxy grafana instances from working correctly. ie. someURL/grafana/ would not work since it expected grafana to hosted on slash (/).

### Developer Changes
  - Migrated to Office Grafana GoLang API
  - refactored packages, moving cmd-> cli, and created cmd/ to allow for multiple binaries to be generated.


##  Release Notes for v0.5.1

**Release Date: 11/03/2023**


### Changes
  - TechDebt: Rewriting the CLI flag parsing to allow for easier testing patterns.  Should mostly be transparent to the user.
  - OrgWatchedFolders added a way to override watched folders for a given organization
  - [#93](https://github.com/esnet/gdg/issues/93) Homebrew support added in.  First pass at having a homebrew release.

### Bug Fixes
  - Tiny patch to fix website documentation navigatioin
  - [#205](https://github.com/esnet/gdg/issues/205) fixes invalid cross-link device when symlink exists to /tmp filesystem.
  - [#206](https://github.com/esnet/gdg/issues/206) fixed behavior issue

### Developer Changes
  - Replaced Makefile with Taskfiles.
  - Added dockertest functionality.  Allows for a consistent testing pattern on dev and CI.
  - postcss security bug.
  - Added a new integration pattern to allow all tests to be executed with tokens and basicauth to ensure behavior is consistent when expected


## Notes on 0.5.x

This is going to be a fairly big release and changing several of the expectations that GDG had before.

The main push for this was to support organizations a bit better, and the only way to really do this correctly was to change the destination path of where the orgs are being saved.  Every entity that supports organization will now be namespace by the org it belongs to.  This will now allow GDG to manage connections and dashboards across multiple organizations.

The other big change, is that most feature are now namespaced under either 'backup' or 'tools' with the exception context which a GDG concept.  The intent of the CLI was getting a bit murky.  There is functionality to create a service account, modify a user permission and so on which is a good bit different from the initial intent of GDG which was to simply manage entities.  Any additional features beyond the crud are under tools.  This might be split into two different binaries later down the line but the separation helps clarify the intent.

Datasources have also been deprecated in favor of 'Connections' to match the Grafana naming convention changes.


##  Release Notes for v0.5.0
**Release Date: 09/01/2023**

### Changes
  - Adding support for Basic CRU for Orgs  [#179](https://github.com/esnet/gdg/issues/179)
  - Renamed 'DataSources' command to 'Connections' to match Grafana's naming convention.
  - Connection Permissions are now supported.  This is an enterprise features and will only function if you have an enterprise version of grafana.  Enterprise features are enabled by setting `enterprise_support: true` for a given context. [#166](https://github.com/esnet/gdg/issues/166)
  - Namespacing all supported entities by organization.
  - Add support for custom S3 Provider (ie. enables ceph, minio and other S3 compatible providers to work with GDG), related [discussion](https://github.com/esnet/gdg/discussions/190)

#### Technical Debt
  - Misc dependencies updates for website and gdg dependencies.
  - Clean up of the Storage interface
  - Updated CICD to only pushed documentation changes on tag release.

### Bug Fixes
  - Fixed issue with import team member with elevated permissions. [#149](https://github.com/esnet/gdg/issues/149)


### Breaking Changes
  - datasources have been renamed as connections.  If you have an existing backup, simply rename the folder to 'connections' and everything should continue working.
  - All Orgs namespaced backups (ie. everything except users and orgs) need to be moved under their respective org folder.  ie.  `org_1` where the given Org has an ID of 1.
  - All commands have now been moved under 'backup' or 'tools' to better reflect their functionality. [#183](https://github.com/esnet/gdg/issues/183)
  - `organization` config is deprecated in favor of `organization_id`.
