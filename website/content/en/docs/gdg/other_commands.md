---
title: "Other Commands"
weight: 18
---

These are miscellaneous commands that don't fit under any category.

### Contexts

Starting with version 0.1.4 contexts are now supported.  Your config can contain one or multiple contexts which are essentially a grafana server configuration.

ctx is shorthand for context and basic CRUD is supported which is mainly tooling to make it easier to avoid updating the yaml file manually

```sh
./bin/gdg ctx list -- Lists all known contexts
./bin/gdg ctx show qa -- shows the configuration for the selected context
./bin/gdg ctx set production -- updates the active config and sets it to the request value.
./bin/gdg ctx delete qa -- Deletes the QA context
./bin/gdg ctx cp qa staging -- copies the qa context to staging and sets it as active
./bin/gdg ctx clear -- Will delete all active contexts leaving only a single example entry
```

### Version

Print the applications release version

```sh
./bin/gdg version
```


```
Build Date: 2022-05-05-13:27:08
Git Commit: 34cc84b3d80080aa93e74ed37739bddc3638997c+CHANGES
Version: 0.1.11
Go Version: go1.18
OS / Arch: darwin amd64

```
