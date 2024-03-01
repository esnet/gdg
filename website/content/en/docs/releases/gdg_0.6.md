---
title: "version v0.6"
description: "Release Notes for v0.6"
date: 2023-09-01T00:00:00
draft: false
images: []
weight: 197
toc: true
---

##  Release Notes for v0.6.0
**Release Date: 03/11/2024**

### Breaking Changes

 - This is a major release, so once again doing some cleanups and introducing some breaking changes.  Please use version v0.5.2 if you wish to maintain backward compatibility.  Previous version organized downloads by org_id.  Example an export for dashboards would be stored in `export/org_1/dashboards`.  The org ID is inconsistent and something that cannot be guaranteed across deployment of grafana.  Instead, we've switched over to using a slug of the OrgName.  The default dashboard backup will go in: `export/org_main-org/dashboards`.

 - This is mentioned below, but `datasources` config key was deprecated and replaced with `connections`.  Previous version would warn about the change, that functionality has been removed.  Please Update to using connections, if you haven't already done so.

 - The import/export dashboards keyword provided confusion.  It has been phased out bit by bit, but all references to it should now be fully removed.  Import is now 'download', and export is now 'upload'.



### Changes
- [#192](https://github.com/esnet/gdg/issues/192) Dropping support for various entities.  import/export no longer supported.  Removed warning for datasources (Deprecated config).  Removed AlertNotification as it's been deprecated from grafana for a while now.
- [#254](https://github.com/esnet/gdg/issues/254) [#258](https://github.com/esnet/gdg/issues/258) org_id usage has been deprecated.  Switching mainly using orgName / SlugName to allow for a more consistent experience between grafana installations.  Change affects gdg and gdg-generate
- User import now has support for random password generator.  Only printed upon import.
-  [#259](https://github.com/esnet/gdg/issues/259) Adding support for Org Properties.  Allowing a user to update a given orgs properties.  Data also added to org Listing.
-  [#251](https://github.com/esnet/gdg/issues/251) Adding a dashboard linter tool. The official grafana is recommended, but GDG will provide similar functionality.


### Bug Fixes
-

### Developer Changes
  - Upgraded to go 1.22
  - Updated documentation instructions relating to install
