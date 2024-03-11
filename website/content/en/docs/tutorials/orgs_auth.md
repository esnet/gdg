---
title: "Organization and Authentication"
weight: 3
date: 2023-09-01T00:00:00
---

## Concepts

At it's core an Organization in grafana is an entity that allows you (the user) to organize and structure entities to seperate access for both usability
and security.  So a Connection under org1 would never be able to be configured to use a dashboard under Org2.


Authentication with GDG and grafana can take a few different patterns.

1. Grafana Admin  - this is your typical admin/admin default user that comes with most installs.  You have full access to do everything.
2. Org Admin - this is a user that is an admin for one or multiple Orgs and can manage most entities under the given org but not high level entities.

Each user can be authenticated with 'BasicAuth' or APIKeys/Service Tokens.

 - Basic Auth allows a user to change Orgs context if they have access to more than one.
 - Service Token/API Keys are bound to a given org, so if the user tries to change the Org, it won't work.  It grants access, viewer, editor, admin for a given Org.


If you are working with multiple Orgs, you will have a much easier time if you use basic auth.  You can certainly simply rotate the tokens as you like though GDG is a bit
better at dealing with basic auth and switching orgs accordingly.


## Organization Workflow

### List Orgs (Grafana Admin)

will retrieve all the components from Grafana and save to local file system.



```sh
gdg backup orgs list

┌────┬───────────┐
│ ID │ ORG       │
├────┼───────────┤
│  1 │ Main Org. │
│  2 │ DumbDumb  │
│  3 │ Moo       │
└────┴───────────┘
```

Let's take a look at our context

```yaml
---local:
storage: ""
enterprise_support: false
url: http://localhost:3000
token: "SomeTokenHere"
user_name: admin
password: admin
organization_name: Main Org.
watched:
    - General
    - Other
connections:
    credential_rules:
        - rules:
            - field: name
              regex: .*
          auth:
            user: user
            password: password
datasources: {}
filter_override:
    ignore_dashboard_filters: false
output_path: test/data
```

The organization_name is set to `Main Org.` and is the default if unspecified.


### Inspect Current Auth Org

Let's have a look at our Token.

```sh
gdg  tools  org tokenOrg
```

```
┌────┬───────────┐
│ ID │ NAME      │
├────┼───────────┤
│  1 │ Main Org. │
└────┴───────────┘
```



This is an immutable value and may cause issues if we switch.  Depending on the call the behavior is to give token preference or basic auth.  So if the basic auth is succesfully namespace into a given org, the token will still point to the wrong one and cause issues.  IF you wish to use Tokens, then avoid using basic auth.


We can also look at what our User Org is set to  using:

```sh
gdg tools org userOrg
```


```
┌────┬───────────┐
│ ID │ NAME      │
├────┼───────────┤
│  1 │ Main Org. │
└────┴───────────┘
```
This value though IS changeable.



### List Dashboards
Now that we take a look at the dashboards under Org 1.

```sh
gdg b dash list
INFO[0002] Listing dashboards for context: 'local'
┌─────┬──────────────────────────────┬──────────────────────────────┬─────────┬───────────┬──────────────┬────────────────────────────────────────────────────────────────┐
│  ID │ TITLE                        │ SLUG                         │ FOLDER  │ UID       │ TAGS         │ URL                                                            │
├─────┼──────────────────────────────┼──────────────────────────────┼─────────┼───────────┼──────────────┼────────────────────────────────────────────────────────────────┤
│ 166 │ Bandwidth Dashboard          │ bandwidth-dashboard          │ General │ 000000003 │ netsage      │ http://localhost:3000/d/000000003/bandwidth-dashboard          │
│ 167 │ Bandwidth Patterns           │ bandwidth-patterns           │ General │ 000000004 │ netsage      │ http://localhost:3000/d/000000004/bandwidth-patterns           │
│ 174 │ Dashboard Makeover Challenge │ dashboard-makeover-challenge │ Other   │ F3eInwQ7z │              │ http://localhost:3000/d/F3eInwQ7z/dashboard-makeover-challenge │
│ 175 │ Flow Analysis                │ flow-analysis                │ Other   │ VuuXrnPWz │ flow,netsage │ http://localhost:3000/d/VuuXrnPWz/flow-analysis                │
│ 176 │ Flow Data for Circuits       │ flow-data-for-circuits       │ Other   │ xk26IFhmk │ flow,netsage │ http://localhost:3000/d/xk26IFhmk/flow-data-for-circuits       │
│ 177 │ Flow Data for Projects       │ flow-data-for-projects       │ Other   │ ie7TeomGz │              │ http://localhost:3000/d/ie7TeomGz/flow-data-for-projects       │
│ 178 │ Flow Data per Country        │ flow-data-per-country        │ Other   │ fgrOzz_mk │ flow,netsage │ http://localhost:3000/d/fgrOzz_mk/flow-data-per-country        │
│ 179 │ Flow Data per Organization   │ flow-data-per-organization   │ Other   │ QfzDJKhik │ flow,netsage │ http://localhost:3000/d/QfzDJKhik/flow-data-per-organization   │
│ 180 │ Flow Information             │ flow-information             │ Other   │ nzuMyBcGk │              │ http://localhost:3000/d/nzuMyBcGk/flow-information             │
│ 181 │ Flows by Science Discipline  │ flows-by-science-discipline  │ Other   │ WNn1qyaiz │ flow,netsage │ http://localhost:3000/d/WNn1qyaiz/flows-by-science-discipline  │
│ 169 │ Individual Flows             │ individual-flows             │ General │ -l3_u8nWk │ netsage      │ http://localhost:3000/d/-l3_u8nWk/individual-flows             │
│ 168 │ Individual Flows per Country │ individual-flows-per-country │ General │ 80IVUboZk │ netsage      │ http://localhost:3000/d/80IVUboZk/individual-flows-per-country │
│ 170 │ Loss Patterns                │ loss-patterns                │ General │ 000000006 │ netsage      │ http://localhost:3000/d/000000006/loss-patterns                │
│ 171 │ Other Flow Stats             │ other-flow-stats             │ General │ CJC1FFhmz │ flow,netsage │ http://localhost:3000/d/CJC1FFhmz/other-flow-stats             │
│ 172 │ Science Discipline Patterns  │ science-discipline-patterns  │ General │ ufIS9W7Zk │ flow,netsage │ http://localhost:3000/d/ufIS9W7Zk/science-discipline-patterns  │
│ 173 │ Top Talkers Over Time        │ top-talkers-over-time        │ General │ b35BWxAZz │              │ http://localhost:3000/d/b35BWxAZz/top-talkers-over-time        │
└─────┴──────────────────────────────┴──────────────────────────────┴─────────┴───────────┴──────────────┴────────────────────────────────────────────────────────────────┘
```

### Switching Organizations


Switching context to Org 2.

```sh
gdg tools orgs set 2
INFO[0000] Succesfully set Org ID for context: local
```

Let's confirm that we trully changed contexts.

```sh
gdg tools org userOrg
```


```
┌────┬───────────┐
│ ID │ NAME      │
├────┼───────────┤
│  2 │ DumbDumb  │
└────┴───────────┘
```

### Listing Orgs Dashboards

Listing dashboards under Org 2 will result in an empty set.

```sh
gdg b dash list
INFO[0000] Listing dashboards for context: 'local'
INFO[0000] No dashboards found
```

Let's switch back to org 1 and donwload our dashboards.

```sh
gdg tools orgs set 1
INFO[0000] Succesfully set Org ID for context: local
```


### Download Orgs Dashboards

```sh
gdg backup dash download
```

```
INFO[0000] Importing dashboards for context: 'local'
┌───────────┬──────────────────────────────────────────────────────────────────────┐
│ TYPE      │ FILENAME                                                             │
├───────────┼──────────────────────────────────────────────────────────────────────┤
│ dashboard │ test/data/org_1/dashboards/General/bandwidth-dashboard.json          │
│ dashboard │ test/data/org_1/dashboards/General/bandwidth-patterns.json           │
│ dashboard │ test/data/org_1/dashboards/Other/dashboard-makeover-challenge.json   │
│ dashboard │ test/data/org_1/dashboards/Other/flow-analysis.json                  │
│ dashboard │ test/data/org_1/dashboards/Other/flow-data-for-circuits.json         │
│ dashboard │ test/data/org_1/dashboards/Other/flow-data-for-projects.json         │
│ dashboard │ test/data/org_1/dashboards/Other/flow-data-per-country.json          │
│ dashboard │ test/data/org_1/dashboards/Other/flow-data-per-organization.json     │
│ dashboard │ test/data/org_1/dashboards/Other/flow-information.json               │
│ dashboard │ test/data/org_1/dashboards/Other/flows-by-science-discipline.json    │
│ dashboard │ test/data/org_1/dashboards/General/individual-flows.json             │
│ dashboard │ test/data/org_1/dashboards/General/individual-flows-per-country.json │
│ dashboard │ test/data/org_1/dashboards/General/loss-patterns.json                │
│ dashboard │ test/data/org_1/dashboards/General/other-flow-stats.json             │
│ dashboard │ test/data/org_1/dashboards/General/science-discipline-patterns.json  │
│ dashboard │ test/data/org_1/dashboards/General/top-talkers-over-time.json        │
└───────────┴──────────────────────────────────────────────────────────────────────┘
```

Please note the path has org_1 in the path.  Starting with version 0.5 of GDG we always namespace the entities we back by the org they belong to.
