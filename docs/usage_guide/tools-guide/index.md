# Tools Guide<no value>

This guide focuses on the 'tools' subcommand.  Every command that isn't specific to a CRUD operation falls under the tools command.

There are a few utility functions that have been introduced that might be useful to the user, or is geared at managing the configuration,
switching contexts or Orgs for a given user and so on.

### Authentication Management

This is mainly added as a convenience mechanism.  It was needed to support some testing and exposing the feature is useful as a really simple CLI to create tokens / service Keys.  You probably should be using other tooling for managing all your service files and tokens.   Unlike most other entities, this is not a backup feature as much as utility.

There are two sub commands for auth, service-accounts and tokens (will be deprecated at some point).

#### Token Management

No longer supported, Deprecated in v0.7.3 as it's been removed from the official grafana API.

Please see below on how to add a token to a service account.

#### Service Accounts


```sh
./bin/gdg tools auth svc  clear       delete all Service Accounts
./bin/gdg tools auth svc  delete      delete the given service account from grafana
./bin/gdg tools auth svc  list        list API Keys
./bin/gdg tools auth svc  new newService <serviceName> <role> [ttl in seconds]
```


```sh
./bin/gdg tools auth svc new AwesomeSauceSvc admin
```

{{< details "New Service" >}}

┌────┬─────────────────┬───────┐
│ ID │ NAME            │ ROLE  │
├────┼─────────────────┼───────┤
│  4 │ AwesomeSauceSvc │ Admin │
└────┴─────────────────┴───────┘
{{< /details >}}

```sh
./bin/gdg tools auth svc newToken 4 AwesomeToken
```

{{< details "New Service" >}}

┌───────────┬──────────┬──────────────┬────────────────────────────────────────────────┐
│ SERVICEID │ TOKEN_ID │ NAME         │ TOKEN                                          │
├───────────┼──────────┼──────────────┼────────────────────────────────────────────────┤
│         4 │        3 │ AwesomeToken │ glsa_a14JOaGExOkDuJHjDWScXbxjTBIXScsw_39df7bf5 │
└───────────┴──────────┴──────────────┴────────────────────────────────────────────────┘
{{< /details >}}

```sh
./bin/gdg tools auth svc list
```

{{< details "Service Listing" >}}

┌────┬─────────────────┬───────┬────────┬──────────┬──────────────┬───────────────┐
│ ID │ SERVICE NAME    │ ROLE  │ TOKENS │ TOKEN ID │ TOKEN NAME   │ EXPIRATION    │
├────┼─────────────────┼───────┼────────┼──────────┼──────────────┼───────────────┤
│ 4  │ AwesomeSauceSvc │ Admin │ 1      │          │              │               │
│    │                 │       │        │        3 │ AwesomeToken │ No Expiration │
└────┴─────────────────┴───────┴────────┴──────────┴──────────────┴───────────────┘
{{< /details >}}


```sh
./bin/gdg tools auth service-accounts delete 4
```

```sh
2025-01-26 16:54:00 INF Deleting Service Accounts for context context=testing serviceAccountId=4
2025-01-26 16:54:00 INF Service account has been removed serviceAccountId=4
```

#### Service Account Tokens

```sh
./bin/gdg tools auth svc tokens new <serviceAccountID> <name> [ttl in seconds]
./bin/gdg tools auth svc tokens clear <serviceAccountID>
```

Examples:
```sh
./bin/gdg tools auth svc tokens new 4 myToken 0
./bin/gdg tools auth svc tokens clear 4
```

### Dashboard Linter

Integrated the official grafana [linter](https://github.com/grafana/dashboard-linter/) into GDG. Allows you to run the linter as part of gdg.

```sh
gdg tools dashboard lint -d bandwidth-patterns -f testing
```

You can execute this on a single dashboard, or a folder.  --autofix is available but should be considered a beta feature.


### Devel
Some developer helper utilities


```sh
./bin/gdg completion  [bash|fish|powershell|zsh] --  Will generate autocompletion for GDG for your favorite shell
./bin/gdg tools devel srvinfo -- print grafana server info
```

### Helpers
A few helpers utilities added to make it easier to construct CLI regex patters

```sh
gdg tools helpers folder encode "ES net/LHC Data Challenge"
```

output: INF Encoded result output=ES\+net/LHC\+Data\+Challenge

The output is what you should be using for filtering your dashboards in the configuration file.

```sh
gdg tools helpers folder decode "ES\+net/LHC\+Data\+Challenge"
```

output: INF Decoded result output="ES net/LHC Data Challenge"

the output is what gdg will match against.


### Organizations
Command can use `organizations` or `org` to set the organizations in the configuration file.

NOTE: this only manages top level of the orgs structure. Mainly used for a lazy man pattern.

```sh
./bin/gdg tools org set --orgName <name> OR --orgSlugName <name> -- Sets a given Org filter.  All Dashboards and Datasources etc are uploaded to the given Org only.
```

Additionally `addUser`, `updateUserRole`, `deleteUser`, `listUsers` are all used to manage a user's membership within a given organization.


### Organizations Preferences

There are a few properties that can be set to change behavior.  Keep in mind that all of these entity need to be owned by the Org, you cannot reference to a dashboard outside of a given org.

```sh
## will set the weekstart as Tuesday and a default Org theme of dark
gdg t orgs prefs set --orgName "Main Org." --theme dark --weekstart tuesday
## Retrieve the Orgs Preferences
gdg t orgs prefs get --orgName "Main Org."
```


```
┌──────────────────┬─────────┐
│ FIELD            │ VALUE   │
├──────────────────┼─────────┤
│ HomeDashboardUID │         │
│ Theme            │ dark    │
│ WeekStart        │ tuesday │
└──────────────────┴─────────┘
```

### Organization Users CRUD

```sh
gdg tools organizations users add [OrgSlug] [userID] Role[admin,editor,viewer] ## Add user to org
    example:  gdg tools organizations users add testing 3 admin
gdg tools organizations users list OrgID ## List all users for a given org
    example:  gdg tools organizations users list 4
gdg tools organizations users updateRole [OrgSlug] [UserId] Role[admin,editor,viewer]
    example: gdg tools organizations users updateRole testing 2 admin
gdg tools organizations users currentOrg  ## displays the logged in User's current associated Org
gdg tools organizations users delete OrgID  ## Removes a user from the given org
```

### Users

CRUD is under the 'backup' command.  The tools subcommand allows you to promote a given user to a grafana admin if you have the permission to do so.

NOTE: admin user is always ignored.

```sh
./bin/gdg tools users promote -u user@foobar.com -- promotes the user to a grafana admin
```
