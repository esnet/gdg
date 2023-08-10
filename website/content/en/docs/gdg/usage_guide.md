---
title: "Usage Guide"
weight: 16
---

Every namespace supporting CRUD operations has the functions: list, download, upload, clear operating on only the monitored folders.


### Authentication Management

This is mainly added as a convenience mechanism.  It was needed to support some testing and exposing the feature is useful as a really simple CLI to create tokens / service Keys.  You probably should be using other tooling for managing all your service files and tokens.   Unlike most other entities, this is not a backup feature as much as utility.

There are two sub commands for auth, service-accounts and tokens (will be deprecated at some point).

#### Token Management


```sh
./bin/gdg auth tokens list -- list current tokens (No access to the actual token secret)
./bin/gdg auth tokens new --  Create a new token.  new <name> <role> [ttl in seconds, forever otherwise]
./bin/gdg auth tokens clear -- Deletes all tokens
```

{{< details "Token Listing" >}}
```
┌────┬─────────┬───────┬───────────────┐
│ ID │ NAME    │ ROLE  │ EXPIRATION    │
├────┼─────────┼───────┼───────────────┤
│  1 │ testing │ Admin │ No Expiration │
└────┴─────────┴───────┴───────────────┘
```
{{< /details >}}

Example of creating a new token.

```sh
./bin/gdg auth tokens new foobar Admin 3600
```

{{< details "New Token" >}}

┌────┬────────┬─────────────────────────────────────────────────────────────┐
│ ID │ NAME   │ TOKEN                                                       │
├────┼────────┼─────────────────────────────────────────────────────────────┤
│  2 │ foobar │ eyJrIjoiNzU2WVhiMEZpVWNlV3hWSUVZQTuIjoiZm9vYmFyIiwiaWQiOjF9 │
└────┴────────┴─────────────────────────────────────────────────────────────┘

{{< /details >}}


#### Service Accounts


```sh
./bin/gdg svc  clear       delete all Service Accounts
./bin/gdg svc  clearTokens delete all tokens for Service Account
./bin/gdg svc  list        list API Keys
./bin/gdg svc  newService  newService <serviceName> <role> [ttl in seconds]
./bin/gdg svc  newToken    newToken <serviceAccountID> <name> [ttl in seconds]
```

```sh
./bin/gdg auth svc newService AwesomeSauceSvc admin
```

{{< details "New Service" >}}

┌────┬─────────────────┬───────┐
│ ID │ NAME            │ ROLE  │
├────┼─────────────────┼───────┤
│  4 │ AwesomeSauceSvc │ Admin │
└────┴─────────────────┴───────┘
{{< /details >}}

```sh
./bin/gdg auth svc newToken 4 AwesomeToken
```

{{< details "New Service" >}}

┌───────────┬──────────┬──────────────┬────────────────────────────────────────────────┐
│ SERVICEID │ TOKEN_ID │ NAME         │ TOKEN                                          │
├───────────┼──────────┼──────────────┼────────────────────────────────────────────────┤
│         4 │        3 │ AwesomeToken │ glsa_a14JOaGExOkDuJHjDWScXbxjTBIXScsw_39df7bf5 │
└───────────┴──────────┴──────────────┴────────────────────────────────────────────────┘
{{< /details >}}

```sh
./bin/gdg auth svc list
```

{{< details "Service Listing" >}}

┌────┬─────────────────┬───────┬────────┬──────────┬──────────────┬───────────────┐
│ ID │ SERVICE NAME    │ ROLE  │ TOKENS │ TOKEN ID │ TOKEN NAME   │ EXPIRATION    │
├────┼─────────────────┼───────┼────────┼──────────┼──────────────┼───────────────┤
│ 4  │ AwesomeSauceSvc │ Admin │ 1      │          │              │               │
│    │                 │       │        │        3 │ AwesomeToken │ No Expiration │
└────┴─────────────────┴───────┴────────┴──────────┴──────────────┴───────────────┘
{{< /details >}}


### Alert Notifications (DEPRECATED)

This will stop working soon both as a concept in grafana and something that GDG will support.

Allows you to manage alertnotifications (an) if you have any setup

```sh
./bin/gdg an list -- Lists all alert notifications
./bin/gdg an download -- retrieve and save all alertnotifications from grafana
./bin/gdg an upload  -- writes all local alert notifications to grafana
./bin/gdg an clear -- Deletes all alert notifications
```

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


### Dashboards

Dashboards are imported or exported from _organization_ specified in configuration file otherwise current organization user is used.

All commands can use `dashboards` or `dash` to manage dashboards

```sh
./bin/gdg dash list -- Lists all current dashboards
./bin/gdg dash download -- Import all dashboards from grafana to local file system
./bin/gdg dash upload -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/gdg dash clear -- Deletes all dashboards
```

You can also use filtering options to list or import your dashboard by folder or by tags.

```sh
./bin/gdg dash download -f myFolder
./bin/gdg dash download -t myTag
./bin/gdg dash download -t tagA -t tagB -t tagC
```

### DataSources

DataSources credentials are keyed by the name of the DataSource.  See see [config example](https://github.com/esnet/gdg/blob/master/conf/importer-example.yml).  If the datasource JSON doesn't have auth enabled, the credentials are ignored.  If Credentials are missing, we'll fall back on default credentials if any exist.  The password is set as a value for basicAuthPassword in the API payload.
Datasources are imported or exported from _organization_ specified in configuration file otherwise current organization user is used.


All commands can use `datasources` or `ds` to manage datasources

```sh
./bin/gdg ds list -- Lists all current datasources
./bin/gdg ds download -- Import all datasources from grafana to local file system
./bin/gdg ds upload -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/gdg ds clear -- Deletes all datasources
```

### Devel
Some developer helper utilities


```sh
./bin/gdg devel completion  [bash|fish|powershell|zsh] --  Will generate autocompletion for GDG for your favorite shell
./bin/gdg devel srvinfo -- print grafana server info
```

### Folders

Mostly optional as Dashboards will create/delete these are needed but if there is additional metadata you wish to persist you can use this to manage them.

```sh
./bin/gdg folders list -- Lists all current folders
./bin/gdg folders download -- Import all folders from grafana to local file system
./bin/gdg folders upload -- Exports all folders from local filesystem
./bin/gdg folders clear -- Deletes all folders
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
./bin/gdg folders list -- Lists all current folder permissions
./bin/gdg folders download -- Retrieve all folders permissions from Grafana
./bin/gdg folders upload -- Exports all folders from local filesystem
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

All commands can use `libraryelements` aliased to `library` and `lib` for laziness purposes.

```sh
./bin/gdg lib list -- Lists all library components
./bin/gdg lib download -- Import all library components from grafana to local file system
./bin/gdg lib upload -- Exports all library components from local filesystem (matching folder filter) to Grafana
./bin/gdg lib clear -- Deletes all library components
./bin/gdg lib  list-connections <Lib Element UID> -- Will list all of the dashboards connected to the Lib Element (Coming in v0.4.2)
```



### Organizations
Command can use `organizations` or `org` to manage organizations.

NOTE: this only manages top level of the orgs structure. It's mainly useful to maintain consistency.

```sh
./bin/gdg org list -- Lists all organizations
./bin/gdg org upload -- Upload Orgs to grafana
./bin/gdg org download -- Download Orgs to grafana
./bin/gdg org set <orgID> -- Sets a given Org filter.  All Dashboards and Datasources etc are uploaded to the given Org only.
```

### Teams

{{< alert icon="👉" text="Admin team members are unable to be exported back.  Currently all members except the server admin will be exported as regular members" />}}

{{< alert icon="👉" text="Users need to be created before team export can succeed" />}}

```sh
./bin/gdg team list  -- Lists all known team members
./bin/gdg team download -- download all known team members
./bin/gdg team upload -- upload all known team members
./bin/gdg team clear -- Delete all known team except admin
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
./bin/gdg users list -- Lists all known users
./bin/gdg users promote -u user@foobar.com -- promotes the user to a grafana admin
./bin/gdg users download -- Lists all known users
./bin/gdg users upload -- Export all users (Not yet supported)
./bin/gdg users clear -- Delete all known users except admin
```
### Token Management (available on +v0.4.2)

Allows the user to create, delete and list current tokens.

```sh
./bin/gdg tokens list -- Lists all known tokens
./bin/gdg tokens clear -- Delete all known tokens
./bin/gdg tokens new <token_name> <role> [ttl] -- Delete all tokens
```


Creation:

```sh
./bin/gdg tokens new foobar viewer 31536000
```

{{< details "New Key" >}}
```
┌────┬────────┬──────────────────────────────────────────────────────────────────────────────────┐
│ ID │ NAME   │ TOKEN                                                                            │
├────┼────────┼──────────────────────────────────────────────────────────────────────────────────┤
│ 15 │ foobar │ eyJrIjoiQXlQY0xMTDg1N09xZkZtc3VGck1iM01YNVNyaHZwTDkiLCJuIjoiZm9vYmFyIiwiaWQiOjF9 │
└────┴────────┴──────────────────────────────────────────────────────────────────────────────────
```
{{< /details >}}


```sh
./bin/gdg tokens list
```

{{< details "Listing" >}}
```
┌────┬────────┬────────┬──────────────────────────┐
│ ID │ NAME   │ ROLE   │ EXPIRATION               │
├────┼────────┼────────┼──────────────────────────┤
│ 12 │ me     │ Admin  │ No Expiration            │
│ 13 │ booh   │ Viewer │ No Expiration            │
│ 14 │ moo    │ Editor │ No Expiration            │
│ 15 │ foobar │ Viewer │ 2024-04-03T18:08:57.000Z │
└────┴────────┴────────┴──────────────────────────┘
```
{{< /details >}}

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
