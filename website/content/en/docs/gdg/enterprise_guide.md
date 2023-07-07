---
title: "Enterprise User Guide"
weight: 17
---
The features listed below are for the enterprise edition of Grafana only.  They will not work on the OSS version.

In order to use these features you need.

1. Update your context to enable enterprise features.  Simply add the following flag to your context.

`enterprise_support: true`

2. Have a running Enterprise version of grafana, I'll defer to the grafana community on instructions on how to set this up.  

### Connections Permissions 

Note:  Available with +v0.4.6.  All of these commands are a subset of the connection command.

All commands can use `permission` or `p` to manage connection permissions.  

```sh
./bin/gdg c permission list -- Lists all current connections permissions
./bin/gdg c permission download -- Download all connections from grafana to local file system
./bin/gdg c permission upload -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/gdg c permission clear -- Deletes all connections Permissions (Leaving only the default values)
```

You can additionally filter by connection slug in order to only operate on a single connection.

`./bin/gdg c permission list --connection my-elastic-connection `

```
┌─────┬───────────┬──────────────────────────────────────────┬───────────────────────────────────────────────┬───────────────────────┬────────────────────────────────────────────┐
│  ID │ UID       │ NAME                                     │ SLUG                     │ TYPE               │ DEFAULT               │ URL                                        │
├─────┼───────────┼──────────────────────────────────────────┼──────────────────────────┼────────────────────┼───────────────────────┼────────────────────────────────────────────┤
│ 712 │ t5xBsTQ4k │ My Elastic Connection                    │ my-elastic-connection    │  elasticsearch     │ false                 │ http://localhost:3000//datasource/edit/712 │
│ 712 │ t5xBsTQ4k │     PERMISSION-->                        │ Edit                     │                    │ sa-gdg-authscope-test │                                            │
│ 712 │ t5xBsTQ4k │     PERMISSION-->                        │ Query                    │                    │ authscope_team_arm    │                                            │
└─────┴───────────┴──────────────────────────────────────────┴──────────────────────────┴────────────────────┴───────────────────────┴────────────────────────────────────────────┘
````