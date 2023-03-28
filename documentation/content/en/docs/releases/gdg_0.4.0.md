---
title: "0.4.0"
description: "Release Notes for v0.4.0"
#lead: "Doks comes with commands for common tasks."
date: 2020-10-13T15:21:01+02:00
lastmod: 2020-10-13T15:21:01+02:00
draft: false
images: []
#menu:
#  docs:
#    parent: "prologue"
weight: 130
toc: true
---
# DRAFT: Release Notes for v0.4.0

This is a major change from the previous one, I'll likely cut the 1.x soon and start following the more typical Semver conventions.  Aka Major version is a breaking change, Minor is just that, patches for previous versions.  

Please see the API Changes notes [below](https://software.es.net/gdg/docs/releases/gdg_0.4.0/#api-sdk-changes).

### New Features

#### Wild card flag
 
 You can now set a flag under each context that will ignore Watched Folders and retrieve all dashboards.

 ```yaml
   context_name:
     filter_override:
      ignore_dashboard_filters: false # 
```      
#### LibraryElements support added.  

Please see the usage guide [here](https://software.es.net/gdg/docs/gdg/usage_guide/#library-elements) and a brief tutorial available [here](https://software.es.net/gdg/docs/tutorials/library_elements/)

#### Folders Update

Introducing a --use-filters.  When enabled will only operate on folders configured.  Default is to create/update/delete all folders in the grafana instance.

### Breaking Changes:

#### SFTP support dropped.  

See the Cloud [configuration](https://software.es.net/gdg/docs/gdg/cloud_configuration/) section.  Switched out the library we relied on, which means the auth has moved out of GDG config and relies on the system config.

## API SDK Changes:

I have been trying to find a proper library to use so I'm not re-writing and reinventing the wheel so to speak. 

For reference, here are all the current "active" (active can be a relative term for some of these project) development I'm aware of.

  - [Grafana Tools SDK](https://github.com/grafana-tools/sdk) Initial version of GDG was based on this project.  It mostly works but getting any PRS accepted can be tedious and it's needs some help.  
  - [Grafana API Go Client](https://github.com/grafana/grafana-api-golang-client) Owned by the Grafana Org which is nice, but it has a slightly different goal.  It's primary goal is to support the terraform provider for Grafana.  I also found some endpoints missing very early on.  So decided not to go with it.
  - Swagger Based:  There's a branch that I've been keeping an eye on.  https://github.com/grafana/grafana-api-golang-client/tree/papagian/generate-client-from-swagger which makes an effort to generate code based on the swagger manifest that's available from Grafana.  It's a mostly automated code that pulls data from the [Schema](https://github.com/grafana/grafana/blob/main/public/api-merged.json) and generates the underlying code.  It hasn't had much traction of late so I ended up forking the project currently available [here](https://github.com/esnet/grafana-swagger-api-golang)

#### Final Choice:

Although the Swagger/OpenAPI based version is not great, I've even ran into a few issues where the documented response 
does not match the result, it's a lot more encompassing and allows further development without being as limited on upstream changes.

### DataModel Changes

I've tried to utilize mostly the same endpoints to recreate the same behavior for all the various entities, but there 
is are some changes.  For most use cases this shouldn't matter.  But you have been officially warned.

## Cloud Support

The previous abstraction library used to provide S3, GS, SFTP has limited activity and introduced some security vulnerabilities.  0.4.X also 
changes some of the cloud behavior.  It relies on the system authentication rather than having the auth in the config file.

Please see the related docs on how to configure your environment.

As the Stow library was removed, SFTP has been dropped.  The list of current supported cloud providers are: S3, GS, Azure.


