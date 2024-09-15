---
title: "Enterprise Guide"
weight: 18
---
The features listed below are for the enterprise edition of Grafana only.  They will not work on the OSS version.

In order to use these features you need.

1. Have a running Enterprise version of grafana, I'll defer to the grafana community on instructions on how to set this up.

For a docker setup, you need to set:

`GF_ENTERPRISE_LICENSE_TEXT='jwt token value'`

### Connections Permissions

Note:  Available with +v0.4.6.  All of these commands are a subset of the connection command.  Requires grafana version: +v10.2.3

All commands can use `permission` or `p` to manage connection permissions.

```sh
./bin/gdg c permission list -- Lists all current connections permissions
./bin/gdg c permission download -- Download all connections from grafana to local file system
./bin/gdg c permission upload -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/gdg c permission clear -- Deletes all connections Permissions (Leaving only the default values)
```

You can additionally filter by connection slug in order to only operate on a single connection.

`./bin/gdg c permission list --connection my-elastic-connection `


{{< details "Permission Listing" >}}
```
┌────┬───────────┬───────────────────┬───────────────┬─────────────────────────────────┬────────────────────────────────┬──────────────────────────────────────────────────────────────┐
│ ID │ UID       │ NAME              │ SLUG          │ TYPE                            │ DEFAULT                        │ URL                                                          │
├────┼───────────┼───────────────────┼───────────────┼─────────────────────────────────┼────────────────────────────────┼──────────────────────────────────────────────────────────────┤
│  1 │ uL86Byf4k │ Google Sheets     │ google-sheets │ grafana-googlesheets-datasource │ false                          │ http://localhost:3000/connections/datasources/edit/uL86Byf4k │
│  1 │ uL86Byf4k │     PERMISSION--> │ Admin         │ User                            │ user:admin                     │                                                              │
│  1 │ uL86Byf4k │     PERMISSION--> │ Edit          │ User                            │ user:tux                       │                                                              │
│  1 │ uL86Byf4k │     PERMISSION--> │ Query         │ User                            │ user:sa-1-test-service-account │                                                              │
│  1 │ uL86Byf4k │     PERMISSION--> │ Query         │ Team                            │ team:engineers                 │                                                              │
│  1 │ uL86Byf4k │     PERMISSION--> │ Admin         │ BuiltinRole                     │ builtInRole:Admin              │                                                              │
│  1 │ uL86Byf4k │     PERMISSION--> │ Query         │ BuiltinRole                     │ builtInRole:Viewer             │                                                              │
│  1 │ uL86Byf4k │     PERMISSION--> │ Query         │ BuiltinRole                     │ builtInRole:Editor             │                                                              │
│  3 │ rg9qPqP7z │ netsage           │ netsage       │ elasticsearch                   │ true                           │ http://localhost:3000/connections/datasources/edit/rg9qPqP7z │
│  3 │ rg9qPqP7z │     PERMISSION--> │ Admin         │ User                            │ user:admin                     │                                                              │
│  3 │ rg9qPqP7z │     PERMISSION--> │ Query         │ BuiltinRole                     │ builtInRole:Editor             │                                                              │
│  3 │ rg9qPqP7z │     PERMISSION--> │ Admin         │ BuiltinRole                     │ builtInRole:Admin              │                                                              │
│  3 │ rg9qPqP7z │     PERMISSION--> │ Query         │ BuiltinRole                     │ builtInRole:Viewer             │                                                              │
│  2 │ i6uqEqE7k │ Netsage TSDS      │ netsage-tsds  │ globalnoc-tsds-datasource       │ false                          │ http://localhost:3000/connections/datasources/edit/i6uqEqE7k │
│  2 │ i6uqEqE7k │     PERMISSION--> │ Admin         │ User                            │ user:admin                     │                                                              │
│  2 │ i6uqEqE7k │     PERMISSION--> │ Query         │ BuiltinRole                     │ builtInRole:Viewer             │                                                              │
│  2 │ i6uqEqE7k │     PERMISSION--> │ Query         │ BuiltinRole                     │ builtInRole:Editor             │                                                              │
│  2 │ i6uqEqE7k │     PERMISSION--> │ Admin         │ BuiltinRole                     │ builtInRole:Admin              │                                                              │
└────┴───────────┴───────────────────┴───────────────┴─────────────────────────────────┴────────────────────────────────┴──────────────────────────────────────────────────────────────┘
```
{{< /details >}}


### Dashboard Permissions

Note:  Available with +v0.7.2.  All of these commands are a subset of the dashboard command.

All commands can use `permission` or `p` to manage connection permissions.

```sh
./bin/gdg dash permission list -- Lists all current connections permissions
./bin/gdg dash permission download -- Download all connections from grafana to local file system
./bin/gdg dash permission upload -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/gdg dash permission clear -- Deletes all connections Permissions (Leaving only the default values)
```


You can additionally filter by dashboard slug in order to only operate on a single connection.


{{< details "Permission Listing" >}}
```
┌────┬─────────────────────┬─────────────────────┬─────────┬───────────┬───────────────────────────────────────────────────────┐
│ ID │ NAME                │ SLUG                │ TYPE    │ UID       │ URL                                                   │
├────┼─────────────────────┼─────────────────────┼─────────┼───────────┼───────────────────────────────────────────────────────┤
│  1 │ Bandwidth Dashboard │ bandwidth-dashboard │ General │ 000000003 │ http://localhost:3000/d/000000003/bandwidth-dashboard │
└────┴─────────────────────┴─────────────────────┴─────────┴───────────┴───────────────────────────────────────────────────────┘
╔═══════════════╦═════════════════════╦════════════╦═══════════╦══════════╦════════════╗
║ DASHBOARD UID ║ DASHBOARD TITLE     ║ USERLOGIN  ║ TEAM      ║ ROLENAME ║ PERMISSION ║
╠═══════════════╬═════════════════════╬════════════╬═══════════╬══════════╬════════════╣
║ 000000003     ║ Bandwidth Dashboard ║ user:admin ║           ║          ║ Admin      ║
║ 000000003     ║ Bandwidth Dashboard ║ user:bob   ║           ║          ║ Edit       ║
║ 000000003     ║ Bandwidth Dashboard ║            ║ musicians ║          ║ Admin      ║
║ 000000003     ║ Bandwidth Dashboard ║            ║           ║ Editor   ║ Edit       ║
║ 000000003     ║ Bandwidth Dashboard ║            ║           ║ Viewer   ║ View       ║
╚═══════════════╩═════════════════════╩════════════╩═══════════╩══════════╩════════════╝
┌────┬────────────────────┬────────────────────┬─────────┬───────────┬──────────────────────────────────────────────────────┐
│ ID │ NAME               │ SLUG               │ TYPE    │ UID       │ URL                                                  │
├────┼────────────────────┼────────────────────┼─────────┼───────────┼──────────────────────────────────────────────────────┤
│  2 │ Bandwidth Patterns │ bandwidth-patterns │ General │ 000000004 │ http://localhost:3000/d/000000004/bandwidth-patterns │
└────┴────────────────────┴────────────────────┴─────────┴───────────┴──────────────────────────────────────────────────────┘
╔═══════════════╦════════════════════╦════════════╦══════╦══════════╦════════════╗
║ DASHBOARD UID ║ DASHBOARD TITLE    ║ USERLOGIN  ║ TEAM ║ ROLENAME ║ PERMISSION ║
╠═══════════════╬════════════════════╬════════════╬══════╬══════════╬════════════╣
║ 000000004     ║ Bandwidth Patterns ║ user:admin ║      ║          ║ Admin      ║
║ 000000004     ║ Bandwidth Patterns ║            ║      ║ Editor   ║ Edit       ║
║ 000000004     ║ Bandwidth Patterns ║            ║      ║ Viewer   ║ View       ║
╚═══════════════╩════════════════════╩════════════╩══════╩══════════╩════════════╝
```
{{< /details >}}
