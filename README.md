# grafana-dashboard-manager

Grafana Dashboard Manager

## Getting started

This project requires Go to be installed. On OS X with Homebrew you can just run `brew install go`.

### Configuring Auth

make a copy of `conf/importer.yml.default` and name it `conf/importer.yml` You'll need administrative privileges to proceed.

You can use either an Auth Token or username/password credentials.  If you configure both then the Token is given priority.

Watched folders under grafana is a white list of folders that are being managed by the tool.  By default only "General" is managed.  

env.output defines where the files will be saved and imported from.

### Running the app

Running it then should be as simple as:

```console
$ make build
$ ./bin/grafana-dashboard-manager
```

Every namespace has three functions: list, import, export, clear operating on only the monitored folders.

#### Dashboards

All commands can use `dashboards` or `dash` to manage dashboards

```sh
./bin/grafana-dashboard-manager dash list -- Lists all current dashboards
./bin/grafana-dashboard-manager dash import -- Import all dashboards from grafana to local file system
./bin/grafana-dashboard-manager dash export -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/grafana-dashboard-manager dash clear -- Deletes all dashboards
```

#### DataSources

Currently only one set of credentials is supported.  grafana.datasource defines the default credentials.  The password is set as a value for 
basicAuthPassword.  


All commands can use `datasources` or `ds` to manage datasources

```sh
./bin/grafana-dashboard-manager ds list -- Lists all current datasources
./bin/grafana-dashboard-manager ds import -- Import all datasources from grafana to local file system
./bin/grafana-dashboard-manager ds export -- Exports all dashboard from local filesystem (matching folder filter) to Grafana
./bin/grafana-dashboard-manager ds clear -- Deletes all datasources
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
