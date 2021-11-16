---
title: "Configuration"
weight: 14
---
## Getting started

This project requires Go to be installed. On OS X with Homebrew you can just run `brew install go`.





make a copy of [conf/importer-example.yml](https://github.com/netsage-project/gdg/blob/master/conf/importer-example.yml) and name it `conf/importer.yml` You'll need GRAFANA ADMINISTRATIVE privileges to proceed.


### Authentication

#### Authentication Token

You can use an Authentication Token / API Key to authenticate with the Grafana API, which can be generated in your Grafana Web UI => Configuration => API Keys. You can then use it in your configuration file (eg. `importer.yml`).
```
context_name: main

contexts:
  main:
    url: https://grafana.example.org
    token: "<<Grafana API Token>>"
    output_path: "myFolder"
    ignore_filters: False  # When set to true all Watched filtered folders will be ignored and ALL folders will be acted on
    watched:
      - Example
    datasources:
      default:
        user: admin
        password: secret
        url_regex:    ## set to pattern to match as well as the name.
      misc:
        user: admin
        password: secret
        url_regex: .*esproxy2*      

global:
  debug: true
  ignore_ssl_errors: false
```

#### Username / Password

You can also use username/password credentials of an admin user to authenticate with the Grafana API. You can specify them in your configuration file (eg. `importer.yml`).
```
context_name: main

contexts:
  main:
    url: https://grafana.example.org
    user_name: <your username>
    password: <your password>
    output_path: "myFolder"
    watched:
      - Example

global:
  debug: true
  ignore_ssl_errors: false
```

### DataSource Credentials.

If the datasource has BasicAuth enabled, then we'll attempt to set the auth with the following precedence on matches:

1. Match of DS credentials based on DS name.
2. Match URL regex for the DS if regex specified.
3. Use Default Credentials if the above two both failed.

#### Notes

If you configure both, Auth Token and Username/Password, then the Token is given priority.

Watched folders under grafana is a white list of folders that are being managed by the tool.  By default only "General" is managed.  

env.output defines where the files will be saved and imported from.

### Global Flags

`globals.debug` when set will print a more verbose output (Development In Progress)
`globals.ignore_ssl_errors` when set will disregard any SSL errors and proceed as expected


### Building/Running the app

Running it then should be as simple as:

```bash
$ make build
$ ./bin/gdg
```
