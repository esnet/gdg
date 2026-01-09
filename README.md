<div  align="center">

Grafana Dash-n-Grab (GDG) -- Resource Manager.  The purpose of this project is to provide an easy-to-use CLI to interact with the grafana API allowing you to backup and restore dashboard, connections, and other resources.

[![Build Status](https://github.com/esnet/gdg/actions/workflows/go.yml/badge.svg)](https://github.com/esnet/gdg/actions/workflows/go.yml)
[![Build Status](https://github.com/esnet/gdg/actions/workflows/hugo.yml/badge.svg)](https://github.com/esnet/gdg/actions/workflows/hugo.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/esnet/gdg)](https://goreportcard.com/report/github.com/esnet/gdg)
[![GoDoc](https://godoc.org/github.com/esnet/gdg?status.svg)](https://godoc.org/github.com/esnet/gdg)
[![codecov](https://codecov.io/gh/esnet/gdg/graph/badge.svg?token=sVEr9Zj6f3)](https://codecov.io/gh/esnet/gdg)
[![GitHub release](https://img.shields.io/github/release/esnet/gdg.svg)](https://github.com/esnet/gdg/releases)


<img src="website/static/logo.png" width="841" height="951">

</div>



The following remote backup locations are supported:
  - AWS S3
  - Google Storage
  - Azure Storage
  - S3 Compatible Storage (Minio, Ceph, etc)

Please find the generated documentation [here](https://software.es.net/gdg/) and the code for updating the docs is available [here](https://github.com/esnet/gdg/blob/main/documentation/content/docs/usage_guide.md)

## Quickstart

[<img src="https://raw.githubusercontent.com/esnet/gdg/main/website/static/quickstart.gif" alt="Quickstart screen">](https://raw.githubusercontent.com/esnet/gdg/main/website/static/quickstart.gif)

## Supported Versions

GDG community will try to support the last 2 major version of grafana.  Though there is nothing preventing you from using it with any version of grafana you are dependent on what the API supports/doesn't support and changes that have been added since then.

New features particularly related to Orgs, ACLs, roles etc are far less likely to work the older your version is.

Current Entities supported (See official docs for more details)


| Resource                | sub-component | Status    | Regex Filtering | Authorization       | Enterprise Only |
|-------------------------|---------------|-----------|-----------------|---------------------|-----------------|
| Connections             |               | Supported | Available       | Token/Basic         |                 |
| Dashboards              |               | Supported | Available       | Token/Basic         |                 |
| Folders                 |               | Supported | Optional        | Token/Basic         |                 |
| Organization            |               | Supported | N/A             | Basic Grafana Admin |                 |
| Teams                   |               | Supported | N/A             | Token/Basic         |                 |
| Users                   |               | Supported | N/A             | Basic               |                 |
| Library Elements        |               | Supported | Available       | Token/Basic         |                 |
| Connections Permissions |               | Supported |                 | Token/Basic         | X               |
| Alerting                | contacts      | Supported | N/A             | Token/Basic         |                 |
| Alerting                | mute-timings  | Supported | N/A             | Token/Basic         |                 |
| Alerting                | policies      | Supported | N/A             | Token/Basic         |                 |
| Alerting                | rules         | Supported | Available       | Token/Basic         |                 |
| Alerting                | templates     | Supported | N/A             | Token/Basic         |                 |


## Release conventions.

GDG mostly follows the semver conventions with some minor modifications.

For those that are unfamiliar semver references to X.Y.Z version patterns with

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

---
GDG is powered by the [Grafana OpenAPI Client](https://github.com/grafana/grafana-openapi-client-go). Any feature exposed by the API could
be added to GDG if desired, feel free to fill out a feature request on our GitHub issue tracker.
