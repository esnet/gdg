# Grafana dash-n-grab 

Grafana Dash-n-Grab (GDG) -- Dashboard/DataSource Manager.  The purpose of this project is to provide an easy to use CLI to interact with the grafana API allowing you to backup and restore dashboard and datasources.

## Getting started

This project requires Go to be installed. On OS X with Homebrew you can just run `brew install go`.

### Configuring Auth

make a copy of `conf/importer-example.yml` and name it `conf/importer.yml` You'll need administrative privileges to proceed.

You can use either an Auth Token or username/password credentials.  If you configure both then the Token is given priority.

Watched folders under grafana is a white list of folders that are being managed by the tool.  By default only "General" is managed.  

env.output defines where the files will be saved and imported from.

### Contexts

Starting with version 0.1.4 contexts are now supported.  Your config can contain one or multiple contexts which are essentially a grafana server configuration.

ctx is shorthand for context

```sh
./bin/grafana-dashboard-manager ctx list -- Lists all known contexts
./bin/grafana-dashboard-manager ctx show -c qa -- shows the configuration for the selected context
./bin/grafana-dashboard-manager ctx set -c production -- updates the active config and sets it to the request value.
```

### Users

Only supported with basic auth.  Users is the only one where basic auth is given priority.  API Auth is not supported, so will try to use basic auth if configured otherwise will warn the user and exit.

```sh
./bin/grafana-dashboard-manager users list -- Lists all known users
./bin/grafana-dashboard-manager users promote -u user@foobar.com -- promotes the user to a grafana admin
```


### Running the app

Running it then should be as simple as:

```console
$ make build
$ ./bin/grafana-dashboard-manager
```

Every namespace has three functions: list, import, export, clear operating on only the monitored folders.

#### Dashboards

Dashboards are imported or exported from _organization_ specified in configuration file otherwise current organitazione user is used.

All commands can use `dashboards` or `dash` to manage dashboards

```sh
./bin/grafana-dashboard-manager dash list -- Lists all current dashboards
./bin/grafana-dashboard-manager dash import -- Import all dashboards from grafana to local file system
./bin/grafana-dashboard-manager dash export -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/grafana-dashboard-manager dash clear -- Deletes all dashboards
```

#### DataSources

DataSources credentials are keyed by the name of the DataSource.  See see [config example](https://github.com/netsage-project/grafana-dashboard-manager/blob/master/conf/importer-example.yml).  If the datasource JSON doesn't have auth enabled, the credentials are igored.  If Credentials are missing, we'll fall back on default credentials if any exist.  The password is set as a value for basicAuthPassword in the API payload.
Datasources are imported or exported from _organization_ specified in configuration file otherwise current organitazione user is used.


All commands can use `datasources` or `ds` to manage datasources

```sh
./bin/grafana-dashboard-manager ds list -- Lists all current datasources
./bin/grafana-dashboard-manager ds import -- Import all datasources from grafana to local file system
./bin/grafana-dashboard-manager ds export -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/grafana-dashboard-manager ds clear -- Deletes all datasources
```

#### Organizations
Command can use `organizations` or `org` to manage organizations.

```sh
./bin/grafana-dashboard-manager org list -- Lists all organizations
```

## Making a release

Install goreleaser.

```sh
brew install goreleaser/tap/goreleaser
brew reinstall goreleaser`
```

export your GITHUB_TOKEN.

```sh
export GITHUB_TOKEN="secret"
```

git tag v0.1.0
goreleaser release
