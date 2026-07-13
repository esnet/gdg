---
title: "Enterprise Guide"
description: "GDG enterprise features guide covering data source permissions, Fine-Grained Access Control (FGAC), and other Grafana Enterprise-only operations."
weight: 18
---

The features listed below are for the enterprise edition of Grafana only. They will not work on the OSS version.

## Prerequisites

To use any enterprise feature you need a running Grafana Enterprise instance with a valid JWT license.

For a Docker setup, pass the license as an environment variable:

```sh
GF_ENTERPRISE_LICENSE_TEXT='<your jwt token>'
```

GDG detects enterprise status automatically by calling Grafana's licensing API — no GDG configuration key is needed. If the connected instance is not Enterprise, GDG will refuse enterprise-only commands with a clear error message.

---

## Fine-Grained Access Control (FGAC) and Datasource Permissions

### What is FGAC?

Fine-Grained Access Control (FGAC) is an **Enterprise license tier feature** that enables restricting datasource query access to specific **users**, **service accounts**, and **teams**. Without FGAC, all users in an organization can query any datasource — FGAC lets you lock down which users or teams can access which datasources.

Grafana's official documentation covers this in two places:

- [Data source permissions](https://grafana.com/docs/grafana/latest/administration/data-source-management/#data-source-permissions) — what permissions mean and how to manage them in the Grafana UI.
- [Role-based access control (RBAC) overview](https://grafana.com/docs/grafana/latest/administration/roles-and-permissions/access-control/) — the broader RBAC system that underpins FGAC, including fixed roles and custom roles.

{{< callout context="caution" title="Two separate Enterprise gates" icon="alert-triangle" >}}
Running Grafana Enterprise is not sufficient on its own. FGAC is a separate license tier on top of Enterprise.

- **Gate 1 — Enterprise license**: The Grafana instance must be running with a valid `GF_ENTERPRISE_LICENSE_TEXT` JWT. GDG checks this via `IsEnterprise()`.
- **Gate 2 — FGAC license tier**: The JWT must include the Fine-Grained Access Control feature. GDG probes this at runtime — if the Grafana instance returns `403 Unlicensed` on a permission write, FGAC is unavailable and GDG skips or reports accordingly.

There is **no GDG configuration key** to enable or disable FGAC. Enablement is determined entirely by what license the Grafana server reports over HTTP.
{{< /callout >}}

### Permission types managed by GDG

When FGAC is available, GDG manages the following permission grant types per datasource:

| Type | Description |
|---|---|
| **User** | Grants a specific Grafana user `Query`, `Edit`, or `Admin` access to the datasource |
| **Team** | Grants all members of a Grafana team `Query`, `Edit`, or `Admin` access |
| **Built-in role** | Grants the `Viewer`, `Editor`, or `Admin` built-in org role a baseline permission |

{{< callout context="note" title="Built-in role grants are system-managed" icon="info-circle" >}}
Built-in role grants (`Viewer`, `Editor`, `Admin`) are managed by Grafana itself and **cannot be removed or modified** via the access-control API — Grafana enforces them as defaults. GDG records these entries in downloaded permission files for visibility, but skips them during upload and clear operations to avoid API errors.
{{< /callout >}}

### How GDG detects FGAC availability

When you run a connection permission command, GDG first calls `IsDataSourcePermissionsEnabled()`, which sends a lightweight probe request to the Grafana RBAC API:

```
POST /api/access-control/datasources/__probe__/users/0
```

Grafana evaluates the license **before** looking up the resource, so:

- `403 Unlicensed` → FGAC is **not** available on this license tier. GDG logs a warning and skips the operation.
- `404 Not Found` / `400 Bad Request` / `200 OK` → License check passed. FGAC is **active** (the sentinel resource simply doesn't exist, which is expected).

This probe is transparent — no datasource is modified.

---

## Connections Permissions

{{< callout context="note" title="Note" icon="info-circle" >}}
Available with +v0.4.6. Requires Grafana Enterprise with FGAC. See [FGAC prerequisites](#fine-grained-access-control-fgac-and-datasource-permissions) above.

Requires Grafana version: +v10.2.3
{{< /callout >}}

Connection permissions let you control which users, service accounts, and teams can query each datasource. All commands are a subset of the connection command and can use `permission` or `p`.

```sh
gdg c permission list       -- Lists all current connection permissions
gdg c permission download   -- Download all connection permissions to local filesystem
gdg c permission upload     -- Upload connection permissions from local filesystem to Grafana
gdg c permission clear      -- Clear all explicit connection permissions (leaves system defaults)
```

You can filter by connection slug to operate on a single datasource:

```sh
gdg c permission list --connection my-elastic-connection
```

### Upload behaviour

When uploading, GDG applies the following rules to each permission entry in the stored file:

1. **Built-in role entries** (`Viewer`, `Editor`, `Admin`) are **skipped** — these are system-managed and cannot be set via the API.
2. **Admin user grants** are **skipped** — Grafana always grants the admin user full access and this cannot be modified.
3. **User grants** (`userId` is resolved by login name at upload time, not the stored integer ID — so the upload is safe across different Grafana instances where auto-increment IDs may differ).
4. **Team grants** (`teamId` is resolved by team name at upload time for the same reason).

### Clear behaviour

`gdg c permission clear` removes all **explicit** user and team grants from every monitored datasource. It does not remove built-in role grants or the admin user grant — those are system defaults that Grafana enforces regardless.

{{< details "Example: Permission Listing" >}}
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

---

## Troubleshooting

### `Requires Enterprise to be enabled`

GDG could not confirm that the connected Grafana instance is running an Enterprise license. Verify that `GF_ENTERPRISE_LICENSE_TEXT` is set correctly in your Grafana container or process environment.

### `403 Unlicensed` on permission writes

The Grafana instance is Enterprise but the license JWT does not include the Fine-Grained Access Control (FGAC) tier. Datasource per-user and per-team permissions require FGAC — see [Grafana's data source permissions docs](https://grafana.com/docs/grafana/latest/administration/data-source-management/#data-source-permissions) for a description of what this feature covers. Contact your Grafana account team to upgrade the license tier.

### Built-in role permissions not applied after upload

This is expected behaviour. Built-in role grants (`Viewer`, `Editor`, `Admin`) on datasources are enforced by Grafana as system defaults and cannot be set or removed via the access-control API. GDG stores them in downloaded files for completeness, but deliberately skips them during upload and clear.

### Datasource permissions for third-party plugin types

Some datasource plugin types (e.g. `grafana-googlesheets-datasource`) may cause the Grafana RBAC API to reject permission writes if the plugin is not installed in the running Grafana instance. Use built-in datasource types (Elasticsearch, Prometheus, etc.) for permission fixtures that must be portable across environments.
