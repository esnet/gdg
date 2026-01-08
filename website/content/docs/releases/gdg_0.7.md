---
title: "version v0.7"
description: "Release Notes for v0.7"
date: 2024-09-11T00:00:00
draft: false
images: [ ]
weight: 3
toc: true
---


## Release Notes for v0.7.2
**Release Date: 01/17/2025**

### Breaking Changes
  - [#318](https://github.com/esnet/gdg/pull/318) Only affects those with no valid configuration present. Removed the default
    config fall back.  For backup and tools functionality a valid configuration file is now required.  A new cli parameter is introduced: `default-config`
    which will print an example configuration to stdout.
### Feature Changes
  - [#319](https://github.com/esnet/gdg/pull/319) Remove the requirement for GF_FEATURE_TOGGLES_ENABLE for nested folder as it was incorrectly required in 0.7.1
  - [#302](https://github.com/esnet/gdg/pull/302) Adding Alerting Contact Points support. (Beta Feature, API/format may change in the next release)
  - [#303](https://github.com/esnet/gdg/pull/303) Cleaning up Permission based listings. (Visualization)
  - [#274](https://github.com/esnet/gdg/pull/274) Adding Dashboard Permissions, enterprise feature.
  - [#337](https://github.com/esnet/gdg/pull/337) Fix to maintain UIDs when updating dashboards
  - [#324](https://github.com/esnet/gdg/pull/324) Introducing a new config option to disregard bad folders

Example:
```yaml
  dashboard_settings:
    ignore_bad_folders: true
  ```
### Bug Fixes
  - [#330](https://github.com/esnet/gdg/pull/330) Fix bug with leading prefix in path when using cloud storage
  - [#333](https://github.com/esnet/gdg/pull/333) Fix bug with documentation search
  - [#314](https://github.com/esnet/gdg/pull/314) Update README to reflected newly supported features

### Security Fixes / Technical Debt
  - [#326](https://github.com/esnet/gdg/pull/326) Bump golang.org/x/crypto from 0.28.0 to 0.31.0
  - [#327](https://github.com/esnet/gdg/pull/327) Security Fix: Non-linear parsing of case-insensitive content in net/html
  - [#328](https://github.com/esnet/gdg/pull/328) Security: Fixing NPM security issues. #328
  - [320](https://github.com/esnet/gdg/pull/320) Changed default branch from `master` to `main`
  - [332](https://github.com/esnet/gdg/pull/332) Added a VHS tape to programmatically re-generate quickstart.gif


## Release Notes for v0.7.1
**Release Date: 09/11/2024**

Major features in this release are:
  - Improvement in performance when dealing with multiple organizations users and preference management.
  - Support for nested folders which affects folders, folder permissions, and dashboards.  See blog post [here](https://software.es.net/gdg/docs/tutorials/working-with-nested-folders/)
  - Regex pattern matching dashboard watched folder (nested folders would require the full path name to match otherwise)

Additionally, api_debug has been introduced.  When enabled it will print every request made to grafana as well as the response recieved from the server.

### Breaking Changes
  - [#289](https://github.com/esnet/gdg/issues/289) Config: Connection settings renamed `exclude_filters` to `filters`
  - Folder Permissions are now saving as uid.json rather than folder name.  Nested folder allows for name collisions, using uids should avoid that issue.
  - Folder Permissions are now saving as slug of nested folder path rather than folder name.  Nested folder allows for name collisions, so foobar/dummy/abcd ==> foobar-dummy-abcd.json
  - Config: ignore_dashboard_filters property has been renamed to dashboard_settings.ignore_filters.

Previous behavior:
  ```yaml
  filter_override:
     ignore_dashboard_filters: true
  ```

has been deprecated in favor of

  ```yaml
  dashboard_settings:
   ignore_filters: true

  ```


### Changes
  - [#171](https://github.com/esnet/gdg/issues/171) Nested Folder support added. (Only available in grafana +v11)
  - Enterprise config flag removed, future versions will programmatically determine version of grafana.
  - [#283](https://github.com/esnet/gdg/issues/283)  Fixing small bug with library connections
  - [#288](https://github.com/esnet/gdg/pull/288) Enterprise: Connection permission will require min. v10.2.3

### Bug/Security Fixes
  - [#268](https://github.com/esnet/gdg/issues/268) Fixing some bad URLs in release
  - [#270](https://github.com/esnet/gdg/issues/270) Fixing cli docs for deletingUserFromOrg, performance tweak to org upload.
  - dependabot Bump github.com/docker/docker from 25.0.0+incompatible to 25.0.6+incompatible.
  - [#285](https://github.com/esnet/gdg/issues/285) Fixing Security issue
  - [#283](https://github.com/esnet/gdg/issues/283) Small bug with dispalying library connections data

### Developer Changes
  - Upgraded to latest grafana openapi client.
  - [#269](https://github.com/esnet/gdg/issues/269) Adding a google analytics tracking on the gdg website.


## Release Notes for v0.7.0
**Release Date: 09/11/2024**


Issues with go releaser process.  No ChangeLog
