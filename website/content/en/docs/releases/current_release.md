---
title: "Current Release Notes"
description: "Release Notes for Current Version"
date: 2023-03-31T15:21:01+02:00
lastmod: 2023-04-14T19:25:12+02:00
draft: true
images: []
weight: 199
toc: true
---

##  Release Notes for v0.5.1
**Release Date: TBD 07/13/2023**


### Changes
  - TechDebt: Rewriting the CLI flag parsing to allow for easier testing patterns.  Should mostly be transparent to the user.
  - OrgWatchedFolders added a way to override watched folders for a given organization
  - [#93](https://github.com/esnet/gdg/issues/205) Homebrew support added in.  First pass at having a homebrew release.

### Bug Fixes
  - Tiny patch to fix website documentation navigatioin
  - [#205](https://github.com/esnet/gdg/issues/205)  fixes invalid cross-link device when symlink exists to /tmp filesystem.
  - [#206](https://github.com/esnet/gdg/issues/206) fixed behavior issue

### Developer Changes
  - Replaced Makefile with Taskfiles.
  - Added dockertest functionality.  Allows for a consistent testing pattern on dev and CI.
  - postcss security bug.
  - Added a new integration pattern to allow all tests to be executed with tokens and basicauth to ensure behavior is consistent when expected

