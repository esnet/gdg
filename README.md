[![Build Status](https://github.com/esnet/gdg/actions/workflows/go.yml/badge.svg)](https://github.com/esnet/gdg/actions/workflows/go.yml)[![Build Status](https://github.com/esnet/gdg/actions/workflows/hugo.yml/badge.svg)](https://github.com/esnet/gdg/actions/workflows/hugo.yml)[![Go Report Card](https://goreportcard.com/badge/github.com/esnet/gdg)](https://goreportcard.com/report/github.com/esnet/gdg)[![GoDoc](https://godoc.org/github.com/esnet/gdg?status.svg)](https://godoc.org/github.com/esnet/gdg)

# Grafana dash-n-grab

Grafana Dash-n-Grab (GDG) -- Dashboard/DataSource Manager.  The purpose of this project is to provide an easy-to-use CLI to interact with the grafana API allowing you to backup and restore dashboard, connections (formerly datasources), and other entities.

The following remote backup locations are supported:
  - AWS S3
  - Google Storage
  - Azure Storage

Please find the generated documentation [here](https://software.es.net/gdg/) and the code for updating the docs is available [here](https://github.com/esnet/gdg/blob/master/documentation/content/docs/usage_guide.md)

## Release conventions.

GDG mostly follows the semver conventions with some minor modifications.

For those that are unfamiliar semver referes to X.Y.Z version patterns with 

  - X = Major version
  - Y = Minor version
  - Z= patch

Most regular releases will increment the patch number.  ie. 0.4.5 is a regular release, and next normal release would be 0.4.6.

Minor version change will likely introduce some breaking change.   For example, renaming datasources to connections or some 
configuration changes that are not backward compatible etc.  

Major version: Is a major feature set change for example, removing cloud support in the base release and introducing a plugin system 
would be 1.X release.  Splitting the GDG binary into a tools and backup cli, or introducing a diff tooling that allow you to compare 
contexts.  i.e.  `gdg diff dashboards prod staging` is a major divergences from the current expectations so it'll be a major version bump.

For more info, please see the release notes and documentation both available [here](https://software.es.net/gdg/)

## Quickstart 

![Quickstart screen](website/static/quickstart.gif)

