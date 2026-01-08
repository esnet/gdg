---
title: "Legacy Versions"
description: "Release Notes for Legacy Versions"
date: 2023-09-01T00:00:00
draft: false
images: [ ]
weight: 5
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

---

##  Release Notes for v0.4.5
**Release Date: 07/13/2023**

### Changes:
- Fixing broken CICD release process

---
##  Release Notes for v0.4.4
**Release Date: 07/13/2023**


### Changes
- [#159](https://github.com/esnet/gdg/issues/159) Due to confusion that has been generated with using import/export.  The action verbs were replaced with download/upload with the previous cmds still left in as functional elements.
  - All 'import' has been replaced with 'download' action.
  - All 'export' has been replaced with an 'upload' action.
- [#160](https://github.com/esnet/gdg/issues/160) Removed deprecated configuration patterns.  Removed `datasources.credentials` and  `datasources.filters`
- [#167](https://github.com/esnet/gdg/issues/167) Adding support for Folder Permissions
- [#170](https://github.com/esnet/gdg/issues/170) OS level characters are no longer supported in folders.  For example '/' and '\' will not be support in any folder GDG backs up.  The behavior combined with the mkdir / path command is too buggy to really
  allow such characters in the names.  The complexity in code needed to support it vs just disallowing it isn't worth it.

### Bug Fixes
- Bug [#156](https://github.com/esnet/gdg/issues/156) fixed.  When gdg binary and config are in completely different paths, gdg is unable to load the configuration file and fallsback on the default config instead.
- BUG #170 fixed.  Added disallowed characters.  For example "/" and "\" will not be supported in folder names
- Some calls failed with invalid SSL.  Fixed secondary code path to also support unsigned SSL


---
##  Release Notes for v0.4.3
**Release Date: 04/14/2023**

### New Features
- Team CRUD support, allows full CRUD on all team and members. Fixes [#127](https://github.com/esnet/gdg/issues/127) and [#147](https://github.com/esnet/gdg/issues/147)
  - Known Bug:  Permissioning not persisted.  All users are added as a member. See issue [149](https://github.com/esnet/gdg/issues/149)
- CLI Tooling introduced to faciliate very basic service management, and token creations for both services and API tokens.
- Improved Credential mapping and filtering introduced.  Allows filtering and credential mapping to be based on any JSON field and regex.


###  Configuration Changes
- DataSource has had a configuration overhaul.  It is technically backward compatible, all previous tests work, with the previous config, but I would highly encourage people to migrate.  Next feature I will drop the backward support.
- URLMatching for Credentials will not work (legacy pattern) if the URL AND the datasource do not match.  If you need URL matching with variable datasource names, you will need to migrate to the new [configuration](https://software.es.net/gdg/docs/gdg/configuration/#datasource).

---
##  Release Notes for v0.4.2

Issue with release, failed CI, so skipping version.

##  Release Notes for v0.4.1
**Release Date: 04/01/2023**

### New Features

#### Library Elements Connections

- Added support for libraryelement connections.  This option allows you to see what dashboards are using the given library.
  - note: You won't be able to delete the library element while there are dashboards utilizing it.
### Bug Fixes
- FIXED: Addressing Login issue when Basic Auth is omitted. #144


---

## Release Notes for v0.4.0
**Release Date: 03/31/2023**

This is a major change from the previous one, I'll likely cut the 1.x soon and start following the more typical Semver conventions.  Aka Major version is a breaking change, Minor is just that, patches for previous versions.

Please see the API Changes notes [below](https://software.es.net/gdg/docs/releases/gdg_0.4.0/#api-sdk-changes).

### New Features

#### Wild card flag

You can now set a flag under each context that will ignore Watched Folders and retrieve all dashboards.

 ```yaml
   context_name:
     dashboard_settings:
       ignore_filters: false #
```
#### LibraryElements support added.

Please see the usage guide [here](https://software.es.net/gdg/docs/gdg/usage_guide/#library-elements) and a brief tutorial available [here](https://software.es.net/gdg/docs/tutorials/library_elements/)

#### Folders Update

Introducing a --use-filters.  When enabled will only operate on folders configured.  Default is to create/update/delete all folders in the grafana instance.

### Breaking Changes:

##### SFTP support dropped.

See the Cloud [configuration](https://software.es.net/gdg/docs/gdg/cloud_configuration/) section.  Switched out the library we relied on, which means the auth has moved out of GDG config and relies on the system config.

#### API SDK Changes:

I have been trying to find a proper library to use so I'm not re-writing and reinventing the wheel so to speak.

For reference, here are all the current "active" (active can be a relative term for some of these project) development I'm aware of.

- [Grafana Tools SDK](https://github.com/grafana-tools/sdk) Initial version of GDG was based on this project.  It mostly works but getting any PRS accepted can be tedious and it's needs some help.
- [Grafana API Go Client](https://github.com/grafana/grafana-api-golang-client) Owned by the Grafana Org which is nice, but it has a slightly different goal.  It's primary goal is to support the terraform provider for Grafana.  I also found some endpoints missing very early on.  So decided not to go with it.
- Swagger Based:  There's a branch that I've been keeping an eye on.  https://github.com/grafana/grafana-api-golang-client/tree/papagian/generate-client-from-swagger which makes an effort to generate code based on the swagger manifest that's available from Grafana.  It's a mostly automated code that pulls data from the [Schema](https://github.com/grafana/grafana/blob/main/public/api-merged.json) and generates the underlying code.  It hasn't had much traction of late so I ended up forking the project currently available [here](https://github.com/esnet/grafana-swagger-api-golang)

#### Final Choice:

Although the Swagger/OpenAPI based version is not great, I've even ran into a few issues where the documented response
does not match the result, it's a lot more encompassing and allows further development without being as limited on upstream changes.

#### DataModel Changes

I've tried to utilize mostly the same endpoints to recreate the same behavior for all the various entities, but there
is are some changes.  For most use cases this shouldn't matter.  But you have been officially warned.

#### Cloud Support

The previous abstraction library used to provide S3, GS, SFTP has limited activity and introduced some security vulnerabilities.  0.4.X also
changes some of the cloud behavior.  It relies on the system authentication rather than having the auth in the config file.

Please see the related docs on how to configure your environment.

As the Stow library was removed, SFTP has been dropped.  The list of current supported cloud providers are: S3, GS, Azure.
