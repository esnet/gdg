---
title: "Usage Guide"
weight: 14
---

Every namespace supporting CRUD operations has the functions: list, import, export, clear operating on only the monitored folders.

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

#### Dashboards

Dashboards are imported or exported from _organization_ specified in configuration file otherwise current organization user is used.

All commands can use `dashboards` or `dash` to manage dashboards

```sh
./bin/gdg dash list -- Lists all current dashboards
./bin/gdg dash import -- Import all dashboards from grafana to local file system
./bin/gdg dash export -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/gdg dash clear -- Deletes all dashboards
```

#### DataSources

DataSources credentials are keyed by the name of the DataSource.  See see [config example](https://github.com/netsage-project/gdg/blob/master/conf/importer-example.yml).  If the datasource JSON doesn't have auth enabled, the credentials are ignored.  If Credentials are missing, we'll fall back on default credentials if any exist.  The password is set as a value for basicAuthPassword in the API payload.
Datasources are imported or exported from _organization_ specified in configuration file otherwise current organization user is used.


All commands can use `datasources` or `ds` to manage datasources

```sh
./bin/gdg ds list -- Lists all current datasources
./bin/gdg ds import -- Import all datasources from grafana to local file system
./bin/gdg ds export -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/gdg ds clear -- Deletes all datasources
```

#### Organizations
Command can use `organizations` or `org` to manage organizations.

```sh
./bin/gdg org list -- Lists all organizations
```

### Users

Only supported with basic auth.  Users is the only one where basic auth is given priority.  API Auth is not supported, so will try to use basic auth if configured otherwise will warn the user and exit.

```sh
./bin/gdg users list -- Lists all known users
./bin/gdg users promote -u user@foobar.com -- promotes the user to a grafana admin
```

