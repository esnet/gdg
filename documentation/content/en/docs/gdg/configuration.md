---
title: "Configuration"
weight: 14
---
## Getting started

This project requires Go to be installed. On OS X with Homebrew you can just run `brew install go`.





make a copy of [conf/importer-example.yml](https://github.com/esnet/gdg/blob/master/conf/importer-example.yml) and name it `conf/importer.yml` You'll need GRAFANA ADMINISTRATIVE privileges to proceed.


### Authentication

#### Authentication Token

You can use an Authentication Token / API Key to authenticate with the Grafana API, which can be generated in your Grafana Web UI => Configuration => API Keys. You can then use it in your configuration file (eg. `importer.yml`).

WARNING: gdg is currently using viper to read in the config.  Since viper makes all keys lowercase, we also have the same limitation.  Camel case will be read in but be aware that a section named fooBar == Foobar == foobar etc.


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
      credentials:
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

### DataSource 

#### DataSource Credentials

If the datasource has BasicAuth enabled, then we'll attempt to set the auth with the following precedence on matches:

1. Match of DS credentials based on DS name.
2. Match URL regex for the DS if regex specified.
3. Use Default Credentials if the above two both failed.

An example of a configuration can be seen below

```yaml
  testing:
    output_path: testing_data
    datasources:
      credentials:
        default:
          user: user
          password: password
        misc:
          user: admin
          password: secret
          url_regex: .*esproxy2*
    url: http://localhost:3000
    user_name: admin
    password: admin
    ignore_filters: False  # When set to true all Watched filtered folders will be ignored and ALL folders will be acted on
    watched:
      - General
      - Other 
 
 ```

#### DataSource Filters

This feature allows you to exclude datasource by name or include them by type.  Please note that the logic switches based on the data type.

**name filter:**

```yaml
...
datasources:
  filters:
    name_exclusions: "DEV-*|-Dev-*"
```

Will exclude any datasource that matches the name regex.

**Type Filter**

Will ONLY include datasource that are listed. 

```yaml
datasources:
  filters:
    valid_types:
      - elasticsearch
```

The snippet above will ONLY import datasources for elasticsearch




#### Notes

If you configure both, Auth Token and Username/Password, then the Token is given priority.
Watched folders under grafana is a white list of folders that are being managed by the tool.  By default only "General" is managed.  

env.output defines where the files will be saved and imported from.

### Global Flags

`globals.debug` when set will print a more verbose output (Development In Progress)
`globals.ignore_ssl_errors` when set will disregard any SSL errors and proceed as expected

### Environment Overrides

If you wish to override certain value via the environment, like credentials and such you can do so.  

The pattern for GDG's is as follows:  "GDG_SECTION__SECTION__keyname"

For example if I want to set the context name to a different value I can use:

```sh
GDG_CONTEXT_NAME="testing" gdg ctx show ## Which will override the value from the context file.
GDG_CONTEXTS__TESTING__URL="www.google.com" Will override the URL with the one provided.
 ```


**NOTE:** Complex data type are not supported, so if the value is an array it can't be currently set.




### Building/Running the app

Running it then should be as simple as:

```bash
$ make build
$ ./bin/gdg
```
