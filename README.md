[![Build Status](https://github.com/esnet/gdg/actions/workflows/go.yml/badge.svg)](https://github.com/esnet/gdg/actions/workflows/go.yml)[![Build Status](https://github.com/esnet/gdg/actions/workflows/hugo.yml/badge.svg)](https://github.com/esnet/gdg/actions/workflows/hugo.yml)[![Go Report Card](https://goreportcard.com/badge/github.com/esnet/gdg)](https://goreportcard.com/report/github.com/esnet/gdg)[![GoDoc](https://godoc.org/github.com/esnet/gdg?status.svg)](https://godoc.org/github.com/esnet/gdg)

# Grafana dash-n-grab

Grafana Dash-n-Grab (GDG) -- Dashboard/DataSource Manager.  The purpose of this project is to provide an easy to use CLI to interact with the grafana API allowing you to backup and restore dashboard, datasources, and other entities.

The following remote backup locations are supported:
  - AWS S3
  - Google Storage
  - Azure Storage

Please find the pretty documentation [here](https://software.es.net/gdg/docs/usage_guide/) and the code for updating the docs is available [here](https://github.com/esnet/gdg/blob/master/documentation/content/docs/usage_guide.md)

## Breaking Changes

ChangeLog: 0.4.X

This is a major change from the previous one, I'll likely cut the 1.x soon and start following the more typical Semver conventions.  

### API SDK Changes:

I have been trying to find a proper library to use so I'm not re-writing and reinventing the wheel so to speak. 

For reference, here are all the current "active" (active can be a relative term for some of these project) development I'm aware of.

  - https://github.com/grafana-tools/sdk Initial version of GDG was based on this project.  It mostly works but getting any PRS accepted can be tedious and it's needs some help.  
  - https://github.com/grafana/grafana-api-golang-client Owned by the Grafana Org which is nice, but it has a slightly different goal.  It's primary goal is to support the terraform provider for Grafana.  I also found some endpoints missing very early on.  So decided not to go with it.
  - Swagger Based:  There's a branch that I've been keeping an eye on.  https://github.com/grafana/grafana-api-golang-client/tree/papagian/generate-client-from-swagger which makes an effort to generate code based on the swagger manifest that's available from Grafana.  It's a mostly automated code that pulls data from the [Schema](https://github.com/grafana/grafana/blob/main/public/api-merged.json) and generates the underlying code.  It hasn't had much traction of late so I ended up forking the project currently available [here](https://github.com/esnet/grafana-swagger-api-golang)

Final Choice:

Although the Swagger/OpenAPI based version is not great, I've even ran into a few issues where the documented response 
does not match the result, it's a lot more encompassing and allows further development without being as limited on upstream changes.

## DataModel Changes

I've tried to utilize mostly the same endpoints to recreate the same behavior for all the various entities, but there 
is are some changes.  For most use cases this shouldn't matter.  But you have been officially warned.

## Cloud Support

The previous abstraction library used to provide S3, GS, SFTP has limited activity and introduced some security vulnerabilities.  0.4.X also 
changes some of the cloud behavior.  It relies on the system authentication rather than having the auth in the config file.

Please see the related docs on how to configure your environment.

As the Stow library was removed, SFTP has been dropped.  The list of current supported cloud providers are: S3, GS, Azure.

## Quickstart 

![Quickstart screen](assets/quickstart.gif)

