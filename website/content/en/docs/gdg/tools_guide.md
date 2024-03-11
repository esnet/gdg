---
title: "Tools Guide"
weight: 17
---

This guide focuses on the 'tools' subcommand.  Every command that isn't specific to a CRUD operation falls under the tools command.

There are a few utility functions that have been introduced that might be useful to the user, or is geared at managing the configuration,
switching contexts or Orgs for a given user and so on.

### Authentication Management

This is mainly added as a convenience mechanism.  It was needed to support some testing and exposing the feature is useful as a really simple CLI to create tokens / service Keys.  You probably should be using other tooling for managing all your service files and tokens.   Unlike most other entities, this is not a backup feature as much as utility.

There are two sub commands for auth, service-accounts and tokens (will be deprecated at some point).

#### Token Management


```sh
./bin/gdg tools auth tokens list -- list current tokens (No access to the actual token secret)
./bin/gdg tools auth tokens new --  Create a new token.  new <name> <role> [ttl in seconds, forever otherwise]
./bin/gdg tools auth tokens clear -- Deletes all tokens
```

{{< details "Token Listing" >}}
```
┌────┬─────────┬───────┬───────────────┐
│ ID │ NAME    │ ROLE  │ EXPIRATION    │
├────┼─────────┼───────┼───────────────┤
│  1 │ testing │ Admin │ No Expiration │
└────┴─────────┴───────┴───────────────┘
```
{{< /details >}}

Example of creating a new token.

```sh
./bin/gdg auth tokens new foobar Admin 3600
```

{{< details "New Token" >}}

┌────┬────────┬─────────────────────────────────────────────────────────────┐
│ ID │ NAME   │ TOKEN                                                       │
├────┼────────┼─────────────────────────────────────────────────────────────┤
│  2 │ foobar │ eyJrIjoiNzU2WVhiMEZpVWNlV3hWSUVZQTuIjoiZm9vYmFyIiwiaWQiOjF9 │
└────┴────────┴─────────────────────────────────────────────────────────────┘

{{< /details >}}


#### Service Accounts


```sh
./bin/gdg tools auth svc  clear       delete all Service Accounts
./bin/gdg tools auth svc  clearTokens delete all tokens for Service Account
./bin/gdg tools auth svc  list        list API Keys
./bin/gdg tools auth svc  newService  newService <serviceName> <role> [ttl in seconds]
./bin/gdg tools auth svc  newToken    newToken <serviceAccountID> <name> [ttl in seconds]
```

```sh
./bin/gdg tools auth svc newService AwesomeSauceSvc admin
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


### Devel
Some developer helper utilities


```sh
./bin/gdg tools devel completion  [bash|fish|powershell|zsh] --  Will generate autocompletion for GDG for your favorite shell
./bin/gdg tools devel srvinfo -- print grafana server info
```


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

### Users

CRUD is under the 'backup' command.  The tools subcommand allows you to promote a given user to a grafana admin if you have the permission to do so.

NOTE: admin user is always ignored.

```sh
./bin/gdg tools users promote -u user@foobar.com -- promotes the user to a grafana admin
```
