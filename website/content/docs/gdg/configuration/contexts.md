---
title: "Contexts"
weight: 103
---

Contexts define a connection to grafana and how to interact with it, authenticate, how to map the connection credentials, how users are created etc. All contexts are defined under the `contexts` key.

### API Token

{{< callout context="danger" title="Danger" icon="alert-octagon" >}}
Token has been deprecated as of v0.9.0, see https://software.es.net/gdg/docs/gdg/getting-started/ on setting up your auth using
a secure location or ENV vars.
{{< /callout >}}

The `token` key is used to authenticate against grafana.  Please be aware that there are certain limits to the token.

1.  They are scoped to a single organization.  They cannot be used across multiple orgs.
2. Some endpoints require basic auth and do not support token authentication even if it has admin rights.

### Connection Settings

The Connection Settings define the behavior related to the connections being imported.  Authorization mechanism for connections and so on.  All of these settings are under `connection:` label.

####  Filters

Filters define a list of rules that are applied to the connection imports/exports.  They allow a user to filter on any part of the JSON, apply a regex match on the given field and include/exclude it accordingly.

Each filter has 3 components:

1. Field, example: `name` which inspects the connection payload extracting the connection name.
2. regex: Any valid regex of string match.  Example: "DEV-*|-Dev-*" will exclude any connections that start with `DEV-` or contains the substring `-Dev-`.
3. inclusive: if set to false, the default will filter out anything that does not match.  Inclusive will only return connections that match the criteria.  aka.  you could exclude all connections that contain Dev but only include connection for elasticsearch.

#### Credential Rules

Connection credentials are not exposed via the API so GDG cannot save those settings.  As such this mechanism allows a user to map a certain connection to a set of credentials.

The rules are defined under `credential_rules` key.  The credentials used will be for the first matching one.  So ideally your default credentials will be the last one in the list. Otherwise, every connection will ignore any other rules and simply use the first one in the list.

Each Credential rule has 2 components.

1. A set of Rules that need to match.  These are always additive, ie every single rule needs to match in order to 'pass'
2. `secure_data` is the location of the auth data used to map the credentials.  If you use the context wizard it will create a simple file for you.

```yaml
basicAuthPassword: password
user: user
```
The auth file either be `json` or `yaml` but only a flat structure is supported. No nested values.
(yaml is recommended, it's generally more readable and json will likely get dropped in future versions)

Some credentials use a different set of keys that is required, simply adjust the auth file to match your needs.

The Rules format for connections settings are the same as the filters. An example configuration can be seen below.

```yaml
        - rules:
            - field: "name"
              regex: "misc"
            - field: "url"
              value: ".*esproxy2*"
          secure_data: "custom.yaml"
```

In the example above if the connection name is misc and the url matches esproxy2 then the connection credentials from custom.json will be used.

The default rule is defined below.  This should match every possible condition and will always use default.json:

```yaml
  - rules:
      # Default
      - field: "name"
        regex: ".*"
    secure_data: "default.yaml"
```
### Dashboard Settings

The entries under `dashboard_settings` define custom behavior for how dashboards are imported.  They can be typically ignored unless you wish to enable a specialized behavior.

Valid values are:

- `ignore_filters`: if you wish to download EVERY folder in grafana and disregard watched folders then set this to true. (Excluding CLI params)

### Monitored Folders

`monitored_folders` is a list of folders to watch.  This is an array of Folder Names and/or Inclusive Regex patterns.  There is currently no pattern to exclude matching regex.

### Monitored Folders Override

Monitored folders applies to all organizations.  If a different behavior is desired, you can use `monitored_folders_override`. If configured instead of using the monitored folders, it will use the folders defined that match the given org name.

Example Config:

```yaml
monitored_folders_override:
  - organization_name: "Staging Org"
    folders: ["Folder1", "Folder2", "General", "Testing"]
```

{{< callout context="caution" title="Caution" icon="alert-triangle" >}}
This setting replaces the watched folders, it is NOT a union of watched folders with the overrides.
{{< /callout >}}

### Organization Name

The `organization_name` key is the name of your organization.  The default Org is `Main Org.`, If you are operating on a specific Org, or renamed the org this value is required.

### Output Path

The `output_path` key is the path relative to the current working directory where all the backups are stored.  If using a cloud provider the output path will be relative to the base prefix.  aka.  if the prefix is 'production' and output_path is 'backups' the backup location will be: `s3://bucketName/backups/production/dashboards/...`

### Password

{{< callout context="danger" title="Danger" icon="alert-octagon" >}}
Password has been deprecated as of v0.9.0, see https://software.es.net/gdg/docs/gdg/getting-started/ on setting up your auth using
a secure location or ENV vars.
{{< /callout >}}


If using Basic auth the `password` field defines the password used to authenticate against grafana.  Both Username and password need to be defined.

{{< callout context="note" title="Note" icon="info-circle" >}}
If you configure both, an auth `Token` and `BasicAuth`, then the Token is given priority.
Watched folders under grafana is a white list of folders that are being managed by gdg. By default, only "General" is managed.

{{< /callout >}}

### Secure_Location

If you set a value for `secure_location`, gdg will use the `secure_location` value for all sensitive data instead of its default location (output_path/secure)

If the path start with a / it'll be treated as an absolute path, otherwise it'll be assumed to exist relative to the output_path configured.

### Storage

To wrap up the previous section, if you setup a cloud provider, you need to set the `storage` to point to the provider you wish to use.

### Username

The `user_name` defines the user to be used for basic auth

### User Settings

Just like connections, we cannot retrieve the credentials for a given user.  The User settings are saved, but the actual passwords need to be generated. There are two patterns a user can use.

1. Default sha256 hash of username
2. Random password

If the default pattern is used the login.json will be used to generate a sha256 which is used as the password. If the user is admin, then the password will be the sha256 of "admin.json".

```sh
echo -n admin.json  | openssl sha256
> SHA2-256(stdin)= f172318957c89be30c2c54abcebb778a86246bbad2325d7133c4dc605319f72b
```

if a random password is selected, then it will only be printed once during the import, but a random value is associated with each user.

to enable this feature ensure `random_password` is set to true.

After which password settings are very simple:

```yaml
min_length: 8 ## defines the minimum length of the password
max_length: 20 ## defines the maximum length of the password
```

### URL

The `url` key is the URL of your Grafana instance.
