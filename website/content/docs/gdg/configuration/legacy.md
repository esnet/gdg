---
title: "Legacy"
weight: 150
---

## Configuration

This is mostly left for reference to older documentation patterns for previous versions.

## Connection

### Connection Credentials

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

#### Legacy Behavior (Prior to v0.4.2)

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

