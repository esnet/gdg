---
title: "version v0.7"
description: "Release Notes for v0.7"
date: 2024-09-11T00:00:00
draft: false
images: [ ]
weight: 196
toc: true
---


## Release Notes for v0.7.2
**Release Date: TBD**

### Breaking Changes


### Changes
 - [#274](https://github.com/esnet/gdg/pull/274) Adding Dashboard Permissions, enterprise feature.
 -

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
