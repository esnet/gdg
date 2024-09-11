---
title: "Working with Nested Folders"
weight: 4
---

Starting with GDG 0.7, support for nested folders has been added.  This feature requires grafana 11+.  You can watch a Intro [video](https://www.youtube.com/watch?v=R9mehA0EssU) or read the offical annoucements [here](https://grafana.com/docs/grafana-cloud/whats-new/2024-02-27-subfolders/).

It is current behind a feature toggle.  You will need to set the folliwing value in your grafana.ini

```toml
[feature_toggles]
enable = nestedFolders,...
```

or have the following ENV variable set

```env
GF_FEATURE_TOGGLES_ENABLE=nestedFolders
```

Additionaly GDG configuration needs to have the behavior enabled.

```yaml
dashboard_settings:
    nested_folders: true
```

Once enabled, the behavior for Dashboards and folders should reflect that.

## Dashboards

For example:

`gdg backup dashboard list`

```sh
┌────┬───────────────────────────────────┬─────────────────────────────┬────────────┬──────────────┬────────────────┬───────────────────────────────┬────────────────────────────────────────────────────────────────────┐
│ ID │ TITLE                             │ SLUG                        │ FOLDER     │ NESTEDPATH   │ UID            │ TAGS                          │ URL                                                                │
├────┼───────────────────────────────────┼─────────────────────────────┼────────────┼──────────────┼────────────────┼───────────────────────────────┼────────────────────────────────────────────────────────────────────┤
│ 21 │ RabbitMQ-Overview                 │ rabbitmq-overview           │ General    │ General             │ Kn5xm-gZk      │ ["rabbitmq-prometheus"]       │ http://localhost:3000/d/Kn5xm-gZk/rabbitmq-overview                │
│ 24 │ Node Exporter Full                │ node-exporter-full          │ dummy      │ Others/dummy        │ rYdddlPWk      │ ["linux"]                     │ http://localhost:3000/d/rYdddlPWk/node-exporter-full               │
│ 26 │ K8s / Storage / Volumes / Cluster │ k8s-storage-volumes-cluster │ someFolder │ Others/someFolder   │ bdx48n30kejuoa │ ["k8s","openshift","storage"] │ http://localhost:3000/d/bdx48n30kejuoa/k8s-storage-volumes-cluster │
└────┴───────────────────────────────────┴─────────────────────────────┴────────────┴──────────────┴────────────────┴───────────────────────────────┴────────────────────────────────────────────────────────────────────┘
```

Note the folder of `Node Exporter Full` is now `Others/dummy`, the watched_folders would also need to be updated as it does a substring match, but it might give you plenty of false positives.

Example: filter on 'dummy' folder also matches /dummy and /a/b/c/d/dummy and /a/dummy/ etc.  It's better to be explicit or have a regex Patern

```yaml
watched:
  - Others/*
```

OR

```yaml
watched:
  - Others/dummy
  - Others/someFolder
```

## Folders

`gdg backup folders list `

```sh
┌────────────────┬──────────────┬────────────┐
│ UID            │ TITLE      │ NESTEDPATH          │
├────────────────┼──────────────┼────────────┤
│ ddxll3n7dse80d │ dummy      │ Others/dummy        │
│ edx4a6qbjt5hcd │ dummy      │ dummy               │
│ fdxll3n62cbnkf │ Others     │ Others              │
│ fdxll3nd7jv9cc │ someFolder │ Others/someFolder   │
└────────────────┴────────────┴──────────────┘
```
