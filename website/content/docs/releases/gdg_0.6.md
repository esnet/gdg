---
title: "version v0.6"
description: "Release Notes for v0.6"
date: 2023-09-01T00:00:00
draft: false
images: [ ]
weight: 4
toc: true
---

## Release Notes for v0.6.0

**Release Date: 03/11/2024**

### Breaking Changes

- This is a major release, so once again doing some cleanups and introducing some breaking changes. Please use version
  v0.5.2 if you wish to maintain backward compatibility. Previous version organized downloads by org_id. Example an
  export for dashboards would be stored in `export/org_1/dashboards`. The org ID is inconsistent and something that
  cannot be guaranteed across installations of grafana. Instead, we've switched over to using a slug of the OrgName. The
  default dashboard backup will go in: `export/org_main-org/dashboards`.

  {{< callout context="caution" title="Caution" icon="alert-triangle" >}}
 If you renamed the default organization to something else besides `Main Org.`, be sure to set `organization_name` in your config, otherwise nothing will work.
  {{< /callout >}}
- The import/export dashboards keyword provided confusion. It has been phased out bit by bit.  In version 0.6 all references have now been removed. Import has been renamed to `download`, and export is now `upload`.
- `organization_id` is deprecated in the importer config in favor of `organization_name`.

### New Features

- Retry logic. You can now set a number of `retry_count` and `retry_delay` in the configuration that retries failed API calls
- User upload now has the ability to generate random passwords.  (Please be aware that those values can't be recreated)
- JSON output has been added.  This can be somewhat unstructured but is available to the user. `--output json` with the default rendering being table format.
- Linter support! (Beta) Grafana official linter has been added to GDG `tools dashboard lint`.
- `gdg-generate` CLI behavior has been updated to better mimic what `gdg` is already doing.  Restructuring into subcommands.
- Org Preferences can now be retrieved when listing Orgs. `--with-preferences`
  {{< callout context="caution" title="Caution" icon="alert-triangle" >}}
This is a heavy call currently till this [issue](https://github.com/grafana/grafana/issues/84309) is resolved.  Use with caution if you have many Organizations in your grafana instance.
  {{< /callout >}}

### Changes

- [#192](https://github.com/esnet/gdg/issues/192) Dropping support for various entities. import/export no longer
  supported. Removed warning for datasources (Deprecated config). Removed AlertNotification as it's been deprecated from
  grafana for a while now.
- [#254](https://github.com/esnet/gdg/issues/254) [#258](https://github.com/esnet/gdg/issues/258) org_id usage has been
  deprecated. Switching mainly using orgName / SlugName to allow for a more consistent experience between grafana
  installations. Change affects gdg and gdg-generate
- User import now has support for random password generator. Only printed upon import.
- [#259](https://github.com/esnet/gdg/issues/259) Adding support for Org Properties. Allowing a user to update a given
  orgs properties. Data also added to org Listing.
- [#251](https://github.com/esnet/gdg/issues/251) Adding a dashboard linter tool. The official grafana is recommended,
  but GDG will provide similar functionality.

### Bug Fixes
- [#253](https://github.com/esnet/gdg/issues/253) In order to manage orgs, the grafana admin that is configured needs to be a part of all organizations.  This ticket adds a sanity check to ensure that the configured Grafana Admin is part of all known organizations.  It then programmatically adds the user (if user confirms and the feature is supported), otherwise gdg will list all orgs that the user needs to be added.
-

### Developer Changes

- Upgraded to go 1.22
- Updated documentation instructions relating to install
- Website theme upgraded to latest version
