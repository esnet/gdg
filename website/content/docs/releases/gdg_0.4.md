---
title: "version v0.4"
description: "Release Notes for v0.4"
date: 2023-03-31T15:21:01+02:00
lastmod: 2023-07-13T00:00:00
draft: false
images: []
weight: 6
toc: true
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
