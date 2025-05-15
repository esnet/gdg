---
title: "version v0.8"
description: "Release Notes for v0.8"
date: 2025-05-15T00:00:00
draft: false
images: [ ]
weight: 8
toc: true
---

## Release Notes for v0.8.0
**Release Date: 05/11/2025**

Major features in this release are:
- Dropped the configuration flag for nested folders (`nested_folders`) as it is now the default behavior.
- Dropped the configuration flag to ignore bad folders (`nested_folders`), instead special characters are handled by
URL encoding the output.

aka. Folder named `/t/'n / r'/booh/k & r` will be stored locally as: `t/n+%2F+r/booh/n+%2F+r.json`

This should allow us to support any folder name but if you have a folder with special characters in its name any regex you
are using should be updated accordingly.


### Breaking Changes


### Changes
- [#408](https://github.com/esnet/gdg/issues/408) Nested Folder support added. (Only available in grafana +v11)

### Bug/Security Fixes
- [#425](https://github.com/esnet/gdg/pull/425) Fixing behavior with missing trailing slash

### Developer Changes
- [#411](https://github.com/esnet/gdg/pull/411) [TechDebt] Removing references to InitTestLegacy (#411)
- Upgraded to latest grafana openapi client.



