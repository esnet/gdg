---
title: "Backup Guide"
description: "GDG backup and restore guide covering list, download, upload, and delete operations for dashboards, alerts, folders, and connections."
weight: 16
---

Every namespace supporting CRUD operations has the functions: list, download, upload, clear operating on only the monitored folders.

### Alerting

Alerting is made up of several type of entities: ContactPoints, Alert Rules, Notification Policy and finally Templates.

Some entities have dependencies on one another.

Example: Alert Rules need the contact points to exist in older to be created. They also need the folders or dashboards that they
operate on to exist.

{{< callout context="caution" title="Caution" icon="alert-triangle" >}}
Unlike most other entities that GDG operates on, alerting entities are global to the Grafana organization and ignore the watched folder list. This applies to notification policies, mute timings, and alert templates (when no `--filter` flag is passed).

Contact points and templates support their own dedicated filtering — see the sections below.
{{< /callout >}}

**Alert Rules are the exception: they are tied to a folder and respect folder filters.**

#### Contact Points

{{< callout note >}} Grafana has a contact point type named 'grafana-default-email' that has an inconsistent behavior.
Unless it has been modified, GDG will ignore it on listing, download and upload.
If it has been modified, it will not be able to clear it due to grafana restriction and show an error for that particular
entity.{{< /callout >}}

```sh
gdg backup alerting contactpoints list -- Lists all current contact points
gdg backup alerting contactpoints download  -- Download all known contact points
gdg backup alerting contactpoints upload -- Upload all contact points
gdg backup alerting contactpoints clear -- Clear all contact points
```

{{< details "Example Output:" >}}
```
┌────────────────┬─────────┬─────────┬───────────────────────────────────────────────────┐
│ UID            │ NAME    │ TYPE    │ SETTINGS                                          │
├────────────────┼─────────┼─────────┼───────────────────────────────────────────────────┤
│ fdxmqkyb5gl4xb │ discord │ discord │ {"url":"[REDACTED]","use_discord_username":false} │
│ aeov0rrgij7r4a │ slack   │ slack   │ {"recipient":"testing","token":"[REDACTED]"}      │
└────────────────┴─────────┴─────────┴───────────────────────────────────────────────────┘
```
{{< /details >}}

#### Filtering

All four contact point operations respect `alert_settings.contact_points.filters` in your context config. Filters match a [gjson](https://github.com/tidwall/gjson) path on the contact point JSON with a regular expression. Set `inclusive: true` to keep only matches (allowlist); omit it or set `false` to drop matches (denylist). Multiple filters are evaluated in order and the first match wins. Rules where the field path does not exist in the payload are silently skipped.

```yaml
alert_settings:
  contact_points:
    filters:
      - field: name
        regex: "discord"        # exclude contact points named "discord"
      - field: receivers.#.type
        regex: "slack"
        inclusive: true         # keep ONLY contact points with a slack receiver
```

Nested array paths such as `receivers.#.type` and `receivers.#.settings.recipient` are supported. See [Contexts]({{< ref "contexts" >}}) for the full filter reference.

#### Notifications

```sh
gdg backup alerting notifications list -- Lists all current contact points
gdg backup alerting notifications download  -- Download all known contact points
gdg backup alerting notifications upload -- Upload all contact points
gdg backup alerting notifications clear -- Clear all contact points
```
{{< details "Example Output:" >}}
```
┌───────────────────────┬──────────────┐
│ UID                   │              │
│ RECEIVER              │ MATCHERS     │
├───────────────────────┼──────────────┤
│ grafana-default-email │ [[foo = 22]] │
│ slack                 │ [[moo = 23]] │
└───────────────────────┴──────────────┘
```
{{< /details >}}

#### Rules

Rules will use watched folders to list act on. If you want to look at all filter across all orgs use: `--no-filters`

**Note:**  This cli option is temporary, it will likely go away on the next release. More robust filtering will be added.

```sh
gdg backup alerting rules list -- Lists all rules
gdg backup alerting rules download  -- Download all known rules
gdg backup alerting rules upload -- Upload all rules
gdg backup alerting rules clear -- Clear all rules
```

**Available Options:**
  - `--folder string`            filter by folder
  - `--ignore-watched-folders`   Default to false, but if passed then will only operate on the list of folders listed in the configuration file
  - `--label stringArray`        Filter by label name value pair. (Additive behavior dashboard includes: label1 AND label2).  ex --label env=staging
  - `--uid string`               filter by rule uid


{{< details "Example Output:" >}}
```
┌──────┬────────────────┬────────────────┬───────────┬───────┐
│ NAME │ UID            │ FOLDERUID      │ RULEGROUP │   FOR │
├──────┼────────────────┼────────────────┼───────────┼───────┤
│ boom │ aeozpk1wn93b4b │ aen349iiivdhcf │ L2        │ 10m0s │
│ moo  │ ceozp0ovszy80c │ den349iklsbuoc │ L1        │  1m0s │
└──────┴────────────────┴────────────────┴───────────┴───────┘
```
{{< /details >}}

#### Timed Intervals

Timed intervals are time window that can be used in conjunction with notification policies.

```sh
gdg backup alerting mute-timings clear       Delete all alert timings for the given Organization
gdg backup alerting mute-timings download    Download all alert timings for the given Organization
gdg backup alerting mute-timings list        List all alert timings for the given Organization
gdg backup alerting mute-timings upload      Upload all alert timings for the given Organization
```

{{< details "Example Output:" >}}
```
┌─────────────┬────────────────┐
│ NAME        │ INTERVAL COUNT │
├─────────────┼────────────────┤
│ after-hours │              2 │
└─────────────┴────────────────┘
╔═══════════╦═══════════════════════╦═══════════╦═══════════════════════════╦════════════════╦═══════════════╗
║ DAYS      ║ LOCATION              ║ MONTHS    ║ TIMES                     ║ WEEKDAYS       ║ YEARS         ║
╠═══════════╬═══════════════════════╬═══════════╬═══════════════════════════╬════════════════╬═══════════════╣
║ [         ║ America/New_York      ║ [         ║ [                         ║ [              ║ [             ║
║   "7:31"  ║                       ║   "1:11"  ║   {                       ║   "monday",    ║   "2021:2031" ║
║ ]         ║                       ║ ]         ║     "end_time": "23:59",  ║   "tuesday",   ║ ]             ║
║           ║                       ║           ║     "start_time": "17:00" ║   "wednesday", ║               ║
║           ║                       ║           ║   },                      ║   "thursday",  ║               ║
║           ║                       ║           ║   {                       ║   "friday"     ║               ║
║           ║                       ║           ║     "end_time": "09:00",  ║ ]              ║               ║
║           ║                       ║           ║     "start_time": "01:00" ║                ║               ║
║           ║                       ║           ║   }                       ║                ║               ║
║           ║                       ║           ║ ]                         ║                ║               ║
║ [ ║ Pacific/Auckland       ║ [         ║ [                         ║ [              ║ [             ║
║   "15:31" ║                       ║   "11:12" ║   {                       ║   "friday",    ║   "1900:2700" ║
║ ]         ║                       ║ ]         ║     "end_time": "10:00",  ║   "thursday",  ║ ]             ║
║           ║                       ║           ║     "start_time": "08:00" ║   "wednesday"  ║               ║
║           ║                       ║           ║   }                       ║ ]              ║               ║
║           ║                       ║           ║ ]                         ║                ║               ║
╚═══════════╩═══════════════════════╩═══════════╩═══════════════════════════╩════════════════╩═══════════════╝
```
{{< /details >}}




#### Templates


```sh
gdg backup alerting templates list -- Lists all templates
gdg backup alerting templates download  -- Download all templates
gdg backup alerting templates upload -- Upload all contact templates
gdg backup alerting templates clear -- Clear all templates
```

All four commands accept a `--filter` flag for name-based regex filtering. An empty filter (the default) matches all templates.

```sh
gdg backup alerting templates list --filter "^prod_.*"     # list only templates whose name starts with "prod_"
gdg backup alerting templates download --filter "critical"  # download only templates whose name contains "critical"
```

{{< details "Example Output:" >}}
```
┌───────────┬────────────┬───────────────────────────────────────────────────────┬──────────────────┐
│ NAME      │ PROVENANCE │ TEMPLATE SNIPPET                                      │ VERSION          │
├───────────┼────────────┼───────────────────────────────────────────────────────┼──────────────────┤
│ test_tpl1 │ api        │ {{- /* This is a copy of the "default.message" tem... │ ea62014659bb56f7 │
│ tpl2_test │ api        │ {{- /* Example displaying additional information, ... │ 53e8e4dd5634e38a │
└───────────┴────────────┴───────────────────────────────────────────────────────┴──────────────────┘
```
{{< /details >}}




### Connections

{{< callout note >}} Starting with v0.4.6 "Datasources" was renamed to connections. {{< /callout >}}

Connections credentials are keyed by the name of the DataSource.  See [config example](https://github.com/esnet/gdg/blob/main/config/gdg-example.yml).  If the connection JSON doesn't have auth enabled, the credentials are ignored.  If Credentials are missing, we'll fall back on default credentials if any exist.  The password is set as a value for basicAuthPassword in the API payload.
Datasources are imported or exported from _organization_ specified in configuration file otherwise current organization user is used.


All commands can use `connection` or `c` to manage datasources.

```sh
gdg backup c list -- Lists all current connections
gdg backup c download -- Import all connections from grafana to local file system
gdg backup c upload -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
gdg backup c clear -- Deletes all connections
```

#### Connection Permissions

{{< callout context="caution" title="Enterprise required" icon="alert-triangle" >}}
Connection permissions require **Grafana Enterprise**. See the [Enterprise Guide]({{< ref "enterprise_guide" >}}) for full details on prerequisites and troubleshooting steps.
{{< /callout >}}

```sh
gdg backup c permission list     -- Lists all current connection permissions
gdg backup c permission download -- Download all connection permissions to local filesystem
gdg backup c permission upload   -- Upload connection permissions from local filesystem to Grafana
gdg backup c permission clear    -- Clear all explicit connection permissions (leaves system defaults)
```


### Dashboards

Dashboards are imported or exported from _organization_ specified in configuration file otherwise current organization user is used.

All commands can use `dashboards` or `dash` to manage dashboards

```sh
gdg backup dash list -- Lists all current dashboards
gdg backup dash download -- Import all dashboards from grafana to local file system
gdg backup dash upload -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
gdg backup dash clear -- Deletes all dashboards
```

You can also use filtering options to list or import your dashboard by folder or by tags.

```sh
gdg backup dash download -f myFolder
gdg backup dash download -t myTag
gdg backup dash download -t tagA -t tagB  -t complex,tagC
```
The command above will return any dashboard that is tagged with `tagA` or `tagB` or `complex,tagC`


**NOTE**: Starting with v0.5.2 full crud support for tag filtering.  You can list,upload,clear,download dashboards using tag filters.  Keep in mind the tag filtering on any matching tags.  ie.  Any dashboard that has tagA or tagB or complex,tagC will be listed,uploaded, etc.

### Dashboard Permissions

{{< callout context="note" title="Note" icon="info-circle" >}}
Available with +v0.7.2. Dashboard permissions are supported in both Grafana OSS and Enterprise.

Starting with Grafana v13, GDG automatically uses the RBAC access-control API
(`/api/access-control/dashboards/{uid}`) instead of the legacy ACL API
(`/api/dashboards/uid/{uid}/permissions`). Downloaded permission files are tagged
with a `gdg_api_version` field so GDG can detect and handle format differences at
upload time — preventing cross-version mismatches.
{{< /callout >}}

All commands can use `permission` or `p` to manage dashboard permissions.

```sh
gdg backup dash permission list -- Lists all current dashboard permissions
gdg backup dash permission download -- Download all dashboard permissions from Grafana
gdg backup dash permission upload -- Upload all dashboard permissions from local filesystem to Grafana
gdg backup dash permission clear -- Clear all dashboard permissions (leaves default values)
```

You can additionally filter by dashboard slug to operate on a single dashboard:

```sh
gdg backup dash permission list --dashboard bandwidth-dashboard
```

{{< details "Example Output (v13+):" >}}
```
┌─────────────────────┬───────────┐
│ DASHBOARD NAME      │ UID       │
├─────────────────────┼───────────┤
│ Bandwidth Dashboard │ 000000003 │
└─────────────────────┴───────────┘
╔═══════════════╦═════════════════════╦════════════╦═══════════╦═════════════╦════════════╗
║ DASHBOARD UID ║ DASHBOARD TITLE     ║ USERLOGIN  ║ TEAM      ║ BUILTIN ROLE║ PERMISSION ║
╠═══════════════╬═════════════════════╬════════════╬═══════════╬═════════════╬════════════╣
║ 000000003     ║ Bandwidth Dashboard ║ bob        ║           ║             ║ Edit       ║
║ 000000003     ║ Bandwidth Dashboard ║            ║ musicians ║             ║ Admin      ║
║ 000000003     ║ Bandwidth Dashboard ║            ║           ║ Editor      ║ Edit       ║
║ 000000003     ║ Bandwidth Dashboard ║            ║           ║ Viewer      ║ View       ║
╚═══════════════╩═════════════════════╩════════════╩═══════════╩═════════════╩════════════╝
```
{{< /details >}}

### Folders

Mostly optional as Dashboards will create/delete these are needed but if there is additional metadata you wish to persist you can use this to manage them.

```sh
gdg backup folders list -- Lists all current folders
gdg backup folders download -- Import all folders from grafana to local file system
gdg backup folders upload -- Exports all folders from local filesystem
gdg backup folders clear -- Deletes all folders
```

### Folder Permissions

This CRUD allows you to import / export folder permissions.  Initial release will be part of v0.4.4.  There are a lot of nested relationship that go with this.

Expectations:
  - Users have to already exist.
  - Teams (if used) need to already exist.
  - Folders also need to already exist.

The Folder Permissions will list, import and re-apply permissions.  But the expectations is that all other entities are already there.  Next few iteration will try to add more concurrency for
this tool and more error checking when entities that don't exist are being referenced.

**NOTE:** Unlike other command, permissions does not have a `clear` function.  Theoretically you could have a folder name with an emtpy array under folder-permissions to clear all known permissions to the folder, but otherwise
clearing permissions from all folders seems too destructive to really be a useful function.

```sh
gdg backup folders list -- Lists all current folder permissions
gdg backup folders download -- Retrieve all folders permissions from Grafana
gdg backup folders upload -- Exports all folders from local filesystem
```

```
┌───────────┬──────────────────────────────────────┬───────────────────────────────────────────────────────────────────────────────────┬─────────────┬────────────────────────────────┬────────┬─────────────────┐
│ FOLDER ID │ FOLDERUID                            │ FOLDER NAME                                                                       │ USERID      │ TEAM NAME                      │ ROLE   │ PERMISSION NAME │
├───────────┼──────────────────────────────────────┼───────────────────────────────────────────────────────────────────────────────────┼─────────────┼────────────────────────────────┼────────┼─────────────────┤
│ 2272      │ dfba969d-565b-481e-a930-53aa5684992c │ sub-flow                                                                          │             │                                │        │                 │
│                                                  │     PERMISSION--->                                                                │ admin       │                                         │ Admin           │
│ 520       │ GPmSOQNnk                            │ EngageMap (internal beta)                                                         │             │                                │        │                 │
│                                                  │     PERMISSION--->                                                                │                                              │ Admin  │ Edit            │
│                                                  │     PERMISSION--->                                                                │                                              │ Editor │ Edit            │
│                                                  │     PERMISSION--->                                                                │                                              │ Viewer │ View            │
│ 2031      │ n3xS8TwVk                            │ Team CMS - US dumb dumb                                                           │             │                                │        │                 │
│                                                  │     PERMISSION--->                                                                │             │ authscope_team_cms             │        │ Edit            │
│ 1746      │ pASPyoQVk                            │ Team DOE-IN-PNNL - DOE-IN Pacific Northwest National Laboratory                   │             │                                │        │                 │
└──────────────────────────────────────────────────┴───────────────────────────────────────────────────────────────────────────────────┴─────────────┴────────────────────────────────┴────────┴─────────────────┘
```

The listing includes the folder name, followed by several lines with "PERMISSION--->" which will each list a permission.  It can a user being granted access or a team being granted a role etc.



### Library Elements

Library elements are components that can be shared among multiple dashboards.  Folder matching will still be applied, so any folders not monitored will be ignored unless explicitly specified.  If wildcard flag is enabled, all elements will be acted on irrelevant of folder location

All commands can use `libraryelements` aliased to `library` and `lib` for laziness purposes.  A more extensive tutorial is available [here](https://software.es.net/gdg/docs/tutorials/library_elements/)

```sh
gdg backup lib list -- Lists all library components
gdg backup lib download -- Import all library components from grafana to local file system
gdg backup lib upload -- Exports all library components from local filesystem (matching folder filter) to Grafana
gdg backup lib clear -- Deletes all library components
gdg backup lib  list-connections <Lib Element UID> -- Will list all of the dashboards connected to the Lib Element (Coming in v0.4.2)
```



### Organizations

{{< callout context="danger" title="Danger" icon="alert-octagon" >}}
Auth:  Requires Grafana Admin

  - Tokens/service account tokens are tied to a specific org and are therefore not supported.
  - Organization Admins don't have access to list all Orgs, therefore are also not supported.

  {{< /callout >}}

Command can use `organizations` or `org` to manage organizations.


```sh
gdg backup org list -- Lists all organizations
gdg backup org upload -- Upload Orgs to grafana
gdg backup org download -- Download Orgs to grafana
```

A tutorial on working with [organizations](https://software.es.net/gdg/docs/tutorials/organization-and-authentication/) is available.

### Teams

{{< callout context="caution" title="Caution" icon="alert-triangle" >}}
Users need to be created before team export can succeed
{{< /callout >}}


```sh
gdg backup team list  -- Lists all known team members
gdg backup team download -- download all known team members
gdg backup team upload -- upload all known team members
gdg backup team clear -- Delete all known team except admin
```

{{< details "Team Listing" >}}
```

┌────┬───────────┬───────┬───────┬─────────┬─────────────┬──────────────┬───────────────────┐
│ ID │ NAME      │ EMAIL │ ORGID │ CREATED │ MEMBERCOUNT │ MEMBER LOGIN │ MEMBER PERMISSION │
├────┼───────────┼───────┼───────┼─────────┼─────────────┼──────────────┼───────────────────┤
│ 4  │ engineers │       │ 1     │ 2       │             │              │                   │
│    │           │       │       │         │ admin       │ Admin        │                   │
│    │           │       │       │         │ tux         │ Member       │                   │
│ 5  │ musicians │       │ 1     │ 1       │             │              │                   │
│    │           │       │       │         │ admin       │ Admin        │                   │
└────┴───────────┴───────┴───────┴─────────┴─────────────┴──────────────┴───────────────────┘

```
{{< /details >}}


### Users

Only supported with basic auth.  Users is the only one where basic auth is given priority.  API Auth is not supported, so will try to use basic auth if configured otherwise will warn the user and exit.

NOTE: admin user is always ignored.

```sh
gdg backup users list -- Lists all known users
gdg backup users download -- Lists all known users
gdg backup users upload -- Export all users (Not yet supported)
gdg backup users clear -- Delete all known users except admin
```

