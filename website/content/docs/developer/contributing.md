---
title: "Contributing"
weight: 62
---

First of all, gdg contains two binaries.  `gdg` which manages grafana entities and provide some tooling facilities
and `gdg-generate` a tool to help generate dashboards from templates to allow for some more flexible management.
gdg-generate is fairly new, but the same quality of PRs would be nice to have.

## GDG

There two types of features that get added to gdg, a `tool` or `backup` feature.

### Backup Feature

A backup feature is some entity that you want to add to GDG that it will track.

For example, if you want to add support for managing a playlist.

Typically, we want to be able to:

1. List all entities in the current namespace (or Org)
2. Download all entities
3. Upload all entities

In order to add a new feature you will need to:

0. Create an issue to track this work and explain the feature being added/requested.
1. Create a new CLI subcommand for playlist. under the backup command.
2. Extend the service to be able to list/download/upload etc the entities.
3. Write a unit test for the given entities. If need be add some seed data under test/data/
4. Update the documentation accordingly to reflect the new changes. All docs live under
   the [website/content](https://github.com/esnet/gdg/tree/main/website/content/docs). All files are in markdown. If
   you wish you can load the website locally by running: `npm install && hugo serve`

#### Testing

There are multiple types of tests.

1. `Integration tests`.  The most common one, an instance of grafana will be available and will connect to it to create/update/delete entities.
2. Unit tests can be created for any unit of work, and will always be executed.
3. CLI tooling tests.  use `task mocks` to generate mocks for your tests.  The purpose of this test is to validate the CLI parsing behavior rather than the service functionality.  Ensure that filtering works, stdout is redirected to the test and can validate the output matches the expectations.

### Tools Feature

This area is more about managing grafana entities.  The tools would provide way for creating service accounts, listing tokens, adding user to orgs, etc.

Testing such features can be a bit more difficult, but we can see enough data to validate the behavior that's always great and makes for a much more stable feature long term.

### Enterprise features

{{< callout context="caution" title="Caution" icon="alert-triangle" >}}
There are a few enterprise features that GDG supports, but unfortunately as there is no enterprise version we can access in CICD testing is very limited.
{{< /callout >}}
