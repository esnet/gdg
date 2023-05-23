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

##  Release Notes for v0.4.4
**Release Date: 04/14/2023**


### New Features
  - #159 Due to confusion that has been generated with using import/export.  The action verbs were replaced with download/upload with the previous cmds still left in as functional elements.
      - All 'import' has been replaced with 'download' action.
      - All 'export' has been replaced with an 'upload' action.

### Bug Fixes
  - Bug #156 fixed.  When gdg binary and config are in completely different paths, gdg is unable to load the configuration file and fallsback on the default config instead.
