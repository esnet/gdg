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


### Changes
  - #159 Due to confusion that has been generated with using import/export.  The action verbs were replaced with download/upload with the previous cmds still left in as functional elements.
      - All 'import' has been replaced with 'download' action.
      - All 'export' has been replaced with an 'upload' action.
  - #160 Removed deprecated configuration patterns.  Removed `datasources.credentials` and  `datasources.filters`
  - #167 Adding support for Folder Permissions
  - #170 OS level characters are no longer supported in folders.  For example '/' and '\' will not be support in any folder GDG backs up.  The behavior combined with the mkdir / path command is too buggy to really
    allow such characters in the names.  The complexity in code needed to support it vs just disallowing it isn't worth it.


### Bug Fixes
  - Bug #156 fixed.  When gdg binary and config are in completely different paths, gdg is unable to load the configuration file and fallsback on the default config instead.
  - BUG #170 fixed.  Added disallowed characters.  For example "/" and "\" will not be supported in folder names
  - Some calls failed with invalid SSL.  Fixed secondary code path to also support unsigned SSL

