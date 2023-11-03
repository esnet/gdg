---
title: "Configuration"
weight: 14
---
## Getting started

This project requires Go to be installed. On OS X with Homebrew you can just run `brew install go`.





make a copy of [config/importer-example.yml](https://github.com/esnet/gdg/blob/master/config/importer-example.yml) and name it `config/importer.yml` You'll need GRAFANA ADMINISTRATIVE privileges to proceed.


### Authentication

#### Authentication Token

You can use an Authentication Token / API Key to authenticate with the Grafana API, which can be generated in your Grafana Web UI => Configuration => API Keys. You can then use it in your configuration file (eg. `importer.yml`).

{{< alert icon="👉" text="gdg is currently using viper to read in the config.  Since viper makes all keys lowercase, we also have the same limitation.  Camel case will be read in but be aware that a section named fooBar == Foobar == foobar etc." />}}
<!-- WARNING: gdg is currently using viper to read in the config.  Since viper makes all keys lowercase, we also have the same limitation.  Camel case will be read in but be aware that a section named fooBar == Foobar == foobar etc. -->

#### Service Accounts

Service Accounts are supported and interchangeable for tokens.  If you wish to use a service account, simply put the token value from the service account for `token:`.  Please make sure you've granted the account the needed permissions for the operations you are trying to perform.

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
    connections:
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

### Connection

#### Connection Credentials

#### Current Behavior (Version +v0.4.2)

If the connection has BasicAuth enabled, then we'll attempt to set the auth with the following rules.

We will try to find a match given the rules specified:

 - field: matches any valid gjson path and retrieves it's value.  ie.  `jsonData.maxConcurrentShardRequests` and validates it against a golang supported [Regex](https://github.com/google/re2/wiki/Syntax).
 - It goes down the list of rules and returns the auth for the first matching one.  The rules should be listed from more specific to more broad.  The default rule ideally should be at the end.

```json
```yaml
  testing:
    output_path: testing_data
    connections:
      credential_rules:
        - rules:
            - field: "name"
              regex: "misc"
            - field: "url"
              regex: ".*esproxy2*"
          auth:
            user: admin
            password: secret
    url: http://localhost:3000
    user_name: admin
    password: admin
    ignore_filters: False  # When set to true all Watched filtered folders will be ignored and ALL folders will be acted on
    watched:
      - General
      - Other 
 
 ```


##### Legacy Behavior:

If the connection has BasicAuth enabled, then we'll attempt to set the auth with the following precedence on matches:

1. Match of DS credentials based on DS name.
2. Match URL regex for the DS if regex specified.
3. Use Default Credentials if the above two both failed.

An example of a configuration can be seen below

```yaml
  testing:
    output_path: testing_data
    connections:
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

#### Connection Filters

#### Current Behavior (+v0.4.2)

You can filter based on any field and have it be an exclusive (default) or inclusive (ie Only allow values that match) to be listed/imported etc.

Pattern matching is the same as the Credentials mapping.

  - field represents any valid JSON Path 
  - regex: any valid [regex](https://github.com/google/re2/wiki/Syntax) supported by golang
  
The example below will exclude any connections named "Google Sheets".  It will also only include connections with the type elasticsearch or mysql

```yaml
contexts:
  testing:
    output_path: test/data
    connections:
      exclude_filters:
        - field: "name"
          regex: "Google Sheets"
        - field: "type"
          regex: "elasticsearch|mysql"
          inclusive: true
```

#### Legacy Behavior

This feature allows you to exclude connection by name or include them by type.  Please note that the logic switches based on the data type.

**name filter:**

```yaml
...
datasources:
  filters:
    name_exclusions: "DEV-*|-Dev-*"
```

Will exclude any connection that matches the name regex.

**Type Filter**

Will ONLY include connection that are listed. 

```yaml
datasources:
  filters:
    valid_types:
      - elasticsearch
```

The snippet above will ONLY import connections for elasticsearch




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
