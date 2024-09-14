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
┌────┬───────────┬───────────────┬───────────────┬─────────────────────────────────┬─────────┬──────────────────────────────────────────────────────────────┐
│ ID │ UID       │ NAME          │ SLUG          │ TYPE                            │ DEFAULT │ URL                                                          │
├────┼───────────┼───────────────┼───────────────┼─────────────────────────────────┼─────────┼──────────────────────────────────────────────────────────────┤
│  1 │ uL86Byf4k │ Google Sheets │ google-sheets │ grafana-googlesheets-datasource │ false   │ http://localhost:3000/connections/datasources/edit/uL86Byf4k │
└────┴───────────┴───────────────┴───────────────┴─────────────────────────────────┴─────────┴──────────────────────────────────────────────────────────────┘
╔════════════════╦════════════════════╦═════════════════╦════════════════════╗
║ CONNECTION UID ║ PERMISSION GRANTED ║ PERMISSION TYPE ║ PERMISSION GRANTEE ║
╠════════════════╬════════════════════╬═════════════════╬════════════════════╣
║ uL86Byf4k      ║ Admin              ║ User            ║ user:admin         ║
║ uL86Byf4k      ║ Admin              ║ User            ║ user:tux           ║
║ uL86Byf4k      ║ Edit               ║ User            ║ user:bob           ║
║ uL86Byf4k      ║ Query              ║ Team            ║ team:musicians     ║
║ uL86Byf4k      ║ Query              ║ BuiltinRole     ║ builtInRole:Viewer ║
║ uL86Byf4k      ║ Query              ║ BuiltinRole     ║ builtInRole:Editor ║
║ uL86Byf4k      ║ Admin              ║ BuiltinRole     ║ builtInRole:Admin  ║
╚════════════════╩════════════════════╩═════════════════╩════════════════════╝
┌────┬───────────┬─────────┬─────────┬───────────────┬─────────┬──────────────────────────────────────────────────────────────┐
│ ID │ UID       │ NAME    │ SLUG    │ TYPE          │ DEFAULT │ URL                                                          │
├────┼───────────┼─────────┼─────────┼───────────────┼─────────┼──────────────────────────────────────────────────────────────┤
│  3 │ rg9qPqP7z │ netsage │ netsage │ elasticsearch │ true    │ http://localhost:3000/connections/datasources/edit/rg9qPqP7z │
└────┴───────────┴─────────┴─────────┴───────────────┴─────────┴──────────────────────────────────────────────────────────────┘
╔════════════════╦════════════════════╦═════════════════╦════════════════════╗
║ CONNECTION UID ║ PERMISSION GRANTED ║ PERMISSION TYPE ║ PERMISSION GRANTEE ║
╠════════════════╬════════════════════╬═════════════════╬════════════════════╣
║ rg9qPqP7z      ║ Admin              ║ User            ║ user:admin         ║
║ rg9qPqP7z      ║ Admin              ║ BuiltinRole     ║ builtInRole:Admin  ║
║ rg9qPqP7z      ║ Query              ║ BuiltinRole     ║ builtInRole:Viewer ║
║ rg9qPqP7z      ║ Query              ║ BuiltinRole     ║ builtInRole:Editor ║
╚════════════════╩════════════════════╩═════════════════╩════════════════════╝
```
{{< /details >}}


