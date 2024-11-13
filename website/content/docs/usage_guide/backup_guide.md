---
title: "Backup Guide"
weight: 16
---

Every namespace supporting CRUD operations has the functions: list, download, upload, clear operating on only the monitored folders.

### Alerting

Alerting is made up of several type of entities: ContactPoints, Alert Rules, Notification Policy and finally Templates.
Currently only Contact Points is supported.

#### Contact Points

{{< callout note >}} Grafana has a contact point type named 'grafana-default-email' that has an inconsistent behavior.
Unless it has been modified, GDG will ignore it on listing, download and upload.  If it has been modified, it will not be able to clear it due to grafana restriction  {{< /callout >}}

```sh
./bin/gdg backup alerting contactpoints list -- Lists all current contact points
./bin/gdg backup alerting contactpoints download  -- Download all known contact points
./bin/gdg backup alerting contactpoints upload -- Upload all contact points
./bin/gdg backup alerting contactpoints clear -- Clear all contact points
```

### Connections

{{< callout note >}} Starting with v0.4.6 "Datasources" was renamed to connections. {{< /callout >}}

Connections credentials are keyed by the name of the DataSource.  See [config example](https://github.com/esnet/gdg/blob/main/config/importer-example.yml).  If the connection JSON doesn't have auth enabled, the credentials are ignored.  If Credentials are missing, we'll fall back on default credentials if any exist.  The password is set as a value for basicAuthPassword in the API payload.
Datasources are imported or exported from _organization_ specified in configuration file otherwise current organization user is used.


All commands can use `connection` or `c` to manage datasources.

```sh
./bin/gdg backup c list -- Lists all current connections
./bin/gdg backup c download -- Import all connections from grafana to local file system
./bin/gdg backup c upload -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/gdg backup c clear -- Deletes all connections
```


### Dashboards

Dashboards are imported or exported from _organization_ specified in configuration file otherwise current organization user is used.

All commands can use `dashboards` or `dash` to manage dashboards

```sh
./bin/gdg backup dash list -- Lists all current dashboards
./bin/gdg backup dash download -- Import all dashboards from grafana to local file system
./bin/gdg backup dash upload -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/gdg backup dash clear -- Deletes all dashboards
```

You can also use filtering options to list or import your dashboard by folder or by tags.

```sh
./bin/gdg backup dash download -f myFolder
./bin/gdg backup dash download -t myTag
./bin/gdg backup dash download -t tagA -t tagB  -t complex,tagC
```
The command above will return any dashboard that is tagged with `tagA` or `tagB` or `complex,tagC`


**NOTE**: Starting with v0.5.2 full crud support for tag filtering.  You can list,upload,clear,download dashboards using tag filters.  Keep in mind the tag filtering on any matching tags.  ie.  Any dashboard that has tagA or tagB or complex,tagC will be listed,uploaded, etc.

### Folders

Mostly optional as Dashboards will create/delete these are needed but if there is additional metadata you wish to persist you can use this to manage them.

```sh
./bin/gdg backup folders list -- Lists all current folders
./bin/gdg backup folders download -- Import all folders from grafana to local file system
./bin/gdg backup folders upload -- Exports all folders from local filesystem
./bin/gdg backup folders clear -- Deletes all folders
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
./bin/gdg backup folders list -- Lists all current folder permissions
./bin/gdg backup folders download -- Retrieve all folders permissions from Grafana
./bin/gdg backup folders upload -- Exports all folders from local filesystem
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
./bin/gdg backup lib list -- Lists all library components
./bin/gdg backup lib download -- Import all library components from grafana to local file system
./bin/gdg backup lib upload -- Exports all library components from local filesystem (matching folder filter) to Grafana
./bin/gdg backup lib clear -- Deletes all library components
./bin/gdg backup lib  list-connections <Lib Element UID> -- Will list all of the dashboards connected to the Lib Element (Coming in v0.4.2)
```



### Organizations

{{< callout context="danger" title="Danger" icon="alert-octagon" >}}
Auth:  Requires Grafana Admin

  - Tokens/service account tokens are tied to a specific org and are therefore not supported.
  - Organization Admins don't have access to list all Orgs, therefore are also not supported.

  {{< /callout >}}

Command can use `organizations` or `org` to manage organizations.


```sh
./bin/gdg backup org list -- Lists all organizations
./bin/gdg backup org upload -- Upload Orgs to grafana
./bin/gdg backup org download -- Download Orgs to grafana
```

A tutorial on working with [organizations](https://software.es.net/gdg/docs/tutorials/organization-and-authentication/) is available.

### Teams

{{< callout context="caution" title="Caution" icon="alert-triangle" >}}
Users need to be created before team export can succeed
{{< /callout >}}


```sh
./bin/gdg backup team list  -- Lists all known team members
./bin/gdg backup team download -- download all known team members
./bin/gdg backup team upload -- upload all known team members
./bin/gdg backup team clear -- Delete all known team except admin
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
./bin/gdg backup users list -- Lists all known users
./bin/gdg backup users download -- Lists all known users
./bin/gdg backup users upload -- Export all users (Not yet supported)
./bin/gdg backup users clear -- Delete all known users except admin
```

