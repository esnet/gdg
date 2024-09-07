---
title: "Configuration"
weight: 14
---

## Configuration

make a copy of [config/importer-example.yml](https://github.com/esnet/gdg/blob/master/config/importer-example.yml) and
name it `config/importer.yml` or simply run `gdg tools ctx new <name>` which will walk you through setting up a new
context to use with your grafana installation.

## Authentication

### Authentication Token

You can use an Authentication Token / API Key to authenticate with the Grafana API, which can be generated in your
Grafana Web UI => Configuration => API Keys. You can then use it in your configuration file (eg. `importer.yml`).

{{< callout context="caution" title="Caution" icon="alert-triangle" >}}
gdg is currently using viper to read in the config. Since viper makes all keys lowercase, we also have the same
limitation. Camel case will be read in but be aware that a section named fooBar == Foobar == foobar etc.
{{< /callout >}}

### Service Accounts

Service Accounts are supported and interchangeable for tokens. If you wish to use a service account, simply put the
token value from the service account for `token:`. Please make sure you've granted the account the needed permissions
for the operations you are trying to perform.

```yaml
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
      credential_rules:
        - rules:
            - field: "name"
              regex: "misc"
            - field: "url"
              value: ".*esproxy2*"
          secure_data: "misc_auth.json"
        - rules:
            - field: "url"
              regex: ".*esproxy2*"
          secure_data: "proxy.json"
        - rules:
            - field: "name"
              regex: ".*"
          secure_data: "default.json"
global:
  debug: true
  ignore_ssl_errors: false
```

### Username / Password

You can also use username/password credentials of an admin user to authenticate with the Grafana API. You can specify
them in your configuration file (eg. `importer.yml`).

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
## Cloud Configuration

Several S3 compatible cloud providers are supported.  Please see this [section](https://software.es.net/gdg/docs/gdg/cloud-configuration/) for more detailed instructions.

## Connection

### Connection Credentials

#### Current Behavior (Version +0.5.2)

If the connection has BasicAuth enabled, then we'll attempt to set the auth with the following rules.

We will try to find a match given the rules specified:

- `field`: matches any valid json path and retrieves its value. ie.  `jsonData.maxConcurrentShardRequests` and validates
  it against a golang supported [Regex](https://github.com/google/re2/wiki/Syntax).
- It goes down the list of rules and returns the auth for the first matching one. The rules should be listed from more
  specific to more broad. The default rule ideally should be at the end.

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
          secure_data: "default.json"
    url: http://localhost:3000
    user_name: admin
    password: admin
    ignore_filters: False  # When set to true all Watched filtered folders will be ignored and ALL folders will be acted on
    watched:
      - General
      - Other

 ```

the secure_data will read the file from {output_path}/secure/. It will then use that
information to construct the datasource.

Default setting if you use basic auth is shown below.

```json
{
  "basicAuthPassword": "password",
  "user": "user"
}
```

#### Version v0.4.2-v0.5.1

{{< details "Legacy Behavior " >}}
Preview behavior did not support the use of a secure/secureData.json pattern, instead an auth: codeblock was used.

Please note that only basicAuth worked prior to version v0.5.2

Example can be seen below:

```yaml
  testing:
    output_path: testing_data
    connections:
      credential_rules:
        - rules:
            - field: name
              regex: .*
          auth:
            user: user
            password: secret
    url: http://localhost:3000
    user_name: admin
    password: admin
    ignore_filters: False  # When set to true all Watched filtered folders will be ignored and ALL folders will be acted on
    watched:
      - General
      - Other

 ```

{{< /details >}}

#### Version Prior to v0.4.2

{{< details "Legacy Behavior " >}}

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

{{< /details >}}

### Connection Filters

#### Current Behavior (+v0.4.2)

You can filter based on any field and have it be an exclusive (default) or inclusive (ie Only allow values that match)
to be listed/imported etc.

Pattern matching is the same as the Credentials mapping.

- field represents any valid JSON Path
- regex: any valid [regex](https://github.com/google/re2/wiki/Syntax) supported by golang

The example below will exclude any connections named "Google Sheets". It will also only include connections with the
type elasticsearch or mysql

```yaml
contexts:
  testing:
    output_path: test/data
    connections:
      filters:
        - field: "name"
          regex: "Google Sheets"
        - field: "type"
          regex: "elasticsearch|mysql"
          inclusive: true
```

#### Legacy Behavior

{{< details "Legacy Behavior " >}}
This feature allows you to exclude connection by name or include them by type. Please note that the logic switches based
on the data type.

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

{{< /details >}}

{{< callout context="note" title="Note" icon="info-circle" >}}
If you configure both, an auth `Token` and `BasicAuth`, then the Token is given priority.
Watched folders under grafana is a white list of folders that are being managed by gdg. By default, only "General" is managed.

{{< /callout >}}

## Organization

The organization is set for a given context via the `orgnization_name`.  If the org is not set, gdg will fallback on the default value that grafana starts out with `Main Org.`

Additionally, if there is an organization specific behavior, it can be added to a context by adding the following config to your config:

```yaml
    watched_folders_override:
      - organization_name: "Special Org"
        folders:
          - General
          - SpecialFolder
```

In this case, watched_folder is ignored in favor of the newly provided list.


## Users

Users can be imported/exported but the behavior is a bit limited.  We cannot retrieve the credentials of the given user.  If the users are uploaded, then any user uploaded will have a new password set.  The default behavior is set password to the sha256 of the login.json

Example:

```sh
echo -n admin.json  | openssl sha256
> SHA2-256(stdin)= f172318957c89be30c2c54abcebb778a86246bbad2325d7133c4dc605319f72b
```

As this can be a security risk if an intruder knows the algorithm, an option to generate random passwords is also available.  This can be configured for any `context`

```yaml
    user:
      random_password: true
      min_length: 8
      max_length: 20
```

The downside, is naturally this is a one time operation.  Once the password is set it once again can no longer be retrieved.  The only time the password is printed is after the successful upload of all users.

## Global Flags

These are flags that apply across all contexts. They are top level configuration and are used to drive gdg's application
behavior.

Here are the currently supported flags you may configure.

```yaml
global:
  debug: true
  ignore_ssl_errors: false ##when set to true will ignore invalid SSL errors
  retry_count: 3 ## Will retry any failed API request up to 3 times.
  retry_delay: 5s  ## will wait for specified duration before trying again.

```

- `debug`: when set will print a more verbose output (Development In Progress). Setting the env flag of DEBUG=1 will
  also generate verbose output for all HTTP calls.
- `ignore_ssl_errors`: when set will disregard any SSL errors and proceed as expected
- `retry_count`: will try N number of times before giving up
- `retry_delay`: a duration to wait before trying again. Be careful with this setting, if it's too short, the retry
  won't matter, if too long the app will be very slow.

## Environment Overrides

{{< callout context="caution" title="Caution" icon="alert-triangle" >}}
Complex data type is not supported. If the value is an array it can't be currently set, via ENV overrides.

{{< /callout >}}

If you wish to override certain value via the environment, like credentials and such you can do so.

The pattern for GDG's is as follows:  `GDG_SECTION__SECTION__keyname`

For example if I want to set the context name to a different value I can use:

```sh
GDG_CONTEXT_NAME="testing" gdg tools ctx show ## Which will override the value from the context file.
GDG_CONTEXTS__TESTING__URL="www.google.com" Will override the URL with the one provided.
 ```




