---
title: "Frequently Asked Questions"
weight: 13
---

### Can I use GDG to backup grafana?

yes and no.  GDG is not a full backup solution.  If you are migrating from one server to another, doing disaster recovery gdg is not the right tool for the job.  Grafana already has some really nice [guides](https://grafana.com/docs/grafana/latest/administration/back-up-grafana/) for how to do that.

GDG is used to backup and recreate specific entities.  Usually it is used to allow for a consistent way to move say a dashboard, connections, alerts etc from one installation of grafana to the next.

The typical use case can be to manage release cycles.  Example:

Create dashboard on dev, test everything out, pull into on a feature branch, deploy to staging environment.  Validate everything looks good and promote it to production.

### Why does GDG not list all my dashboards?

By default, GDG works on a select list of `watched` or monitored folders.  If none are specified, it will limit itself to only operating on `General`.

The folders that it DOES monitor, it is assumed that gdg has full control over. Meaning, anything in those folders is under its pervue, so it may delete all connections and replace them with the ones in its exports.  As far as it knows, nobody else should be writing to it.  The upload operation implies that you want to sync the export data with the data found in the backup local, cloud, or otherwise.

PS. if you want to list/import etc all dashboards, you can set the following config for your context.

```yaml
    dashboard_settings:
      ignore_filters: false # When set to true all Watched filtered folders will be ignored and ALL folders will be acted on

```

### I need feature X, can you please add that in?

Maybe? If there's enough cycles, it could benefit others, and the feature makes sense I'd be happy to.  It is also an OSS project, so contributions are always appreciated and welcome. See [contributing](https://software.es.net/gdg/docs/developer/contributing/) for more info.

### I need help, where do I go?

There is a "Discussion" area on github where you can start a [conversation](https://github.com/esnet/gdg/discussions) and ask questions.  If you think this is a bug in GDG itself, then please file an issue [here](https://github.com/esnet/gdg/issues).  There is a small slack on channel titled `#gdg` in the grafana slack, you're free to join [here](https://slack.grafana.com/)

### I don't like GDG because of: A, B, C... what else can I use

 - [Grafanactl](https://github.com/grafana/grafanactl) Recently Grafana released an official tool to manage resources.
 - [Grizzly](https://github.com/grafana/grizzly/) Prior version of GrafanaCTL

Know of any others? Create a PR and add it to the list
