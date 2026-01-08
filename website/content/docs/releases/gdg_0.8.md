---
title: "version v0.8"
description: "Release Notes for v0.8"
date: 2025-05-15T00:00:00
draft: false
images: [ ]
weight: 2
toc: true
---

## Release Notes for v0.8.1

**Release Date: 06/30/2025**

### BugFix:
  - [#456](https://github.com/esnet/gdg/issues/456) Issued with folders containing spaces

### Changes:
  - [#450](https://github.com/esnet/gdg/issues/450) Update GoReleaser configurations (#453)
      Changes the patterns for brew installs from Formula to Cask: https://goreleaser.com/deprecations/#brews
  - [#454](https://github.com/esnet/gdg/pull/454) Adding a logo GDG logo (#454)
  - [#445](https://github.com/esnet/gdg/issues/445) Added a global --context to easy switch without needing a config change.

## Release Notes for v0.8.0
**Release Date: 06/25/2025**

Major features in this release are:
- Dropped the configuration flag for nested folders (`nested_folders`) as it is now the default behavior.
- Dropped the configuration flag to ignore bad folders (`nested_folders`), instead special characters are handled by
URL encoding the output.

aka. Folder named `/t/'n / r'/booh/k & r` will be stored locally as: `t/n+%2F+r/booh/n+%2F+r.json`

This should allow us to support any folder name but if you have a folder with special characters in its name any regex you
are using should be updated accordingly.

### Min Recommended Grafana Versions:

While most behavior should be backward compatible gdg v0.8.x is tested with grafana 11 and 12. Anything older use at your own
risk. Please use grafana +11.

### Breaking Changes
  - [#374](https://github.com/esnet/gdg/pull/374) Removed Tooling around creating a token, service account has replaced this feature.
  - [#412](https://github.com/esnet/gdg/issues/412) Updates to library elements introducing a new data model. Previous backups will not be compatible with v0.8

### Changes
- [#408](https://github.com/esnet/gdg/issues/408) Nested Folder support added as a default behavior
- [#134](https://github.com/esnet/gdg/issues/134) Adding support for Alerting entities. (rules, contact points, templates, policies)
- [#421](https://github.com/esnet/gdg/issues/421) Added support for an auth file as well as secure location override.

### Bug/Security Fixes
- [#425](https://github.com/esnet/gdg/pull/425) Fixing behavior with missing trailing slash

### Developer Changes
- [#411](https://github.com/esnet/gdg/pull/411) [TechDebt] Removing references to InitTestLegacy (#411)
- Upgraded to latest grafana openapi client.
- [#427](https://github.com/esnet/gdg/pull/427) Re-enabling code coverage report uploading to cobertura

### Tech Updates
  - updated to latest grafana-api client.
  - various golang/npm updates
  - removed and updated tests to no longer use a deprecated pattern.
  - gopls modernize tool updates

