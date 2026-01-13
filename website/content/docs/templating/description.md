---
title: "Usage Guide"
weight: 25
---

GDG has now introduced a new supporting tool that works in conjunction with GDG. It is currently dependent on the GDG
configuration
since it will operate on the currently selected context. You can confirm what the current context is by
running `gdg tools ctx show`

For example, my current output is as follows:

```yaml
context_name:
  storage: ""
  enterprise_support: true
  url: http://localhost:3000
  token: ""
  user_name: admin
  password: admin
  organization_name: Main Org.
  watched_folders_override: [ ]
  watched:
    - General
    - Other
  connections:
    credential_rules:
      - rules:
          - field: name
            regex: misc
          - field: url
        secure_data: "misc.json"
      - rules:
          - field: url
            regex: .*esproxy2*
        secure_data: "proxy.json"
      - rules:
          - field: name
            regex: .*
        secure_data: "default.yaml"
  dashboard_settings:
    ignore_filters: false
  output_path: test/data
```

Most of the config isn't that interesting, except the output_path will be used to determine where the newly generated
dashboards will be. Make sure you have a valid configuration before continuing.

### What does gdg-generate do?

There are use cases where an almost identical dashboard is needed except we need to replace certain parts of it.

For example, parts of a query need to be different, a different title, brand it to specific customer with a different
logo, or footer. All of these are difficult to control from grafana itself and even in the best case scenario it's not
great user experience. This allows you to configure and generate a new dashboard with any set of variables and
dictionaries that you will seed to the tool.

### Configuration

The configuration that drives this application is `templates.yml`. You can see an example below.

```yaml
entities:
  dashboards:
    - template_name: "template_example"  ##Matches the name to a file under ouput_path/templates/*.go.tmpl
      output: ## The section below defines one or multiple destination and the associated configuration
        ## that goes with it.
        - folder: "General"  ## Name of the new folder where the template will be created
          organization_name: "Main Org."
          dashboard_name: ""  ## Optional, defaults to template_name.json
          template_data: ## Template Data the dictionary of datasets that can be used in the template,
            # it's basically your 'seed data'.  Everything is contains is absolutely arbitrary
            # and can have any structure as long as it's valid yaml
            Title: "Bob Loves Candy"  ## Dashboard Titlte
            enabledlight: false  ## Boolean check to enable/disable behavior
            lightsources: ## some arbitrary list we get to play with
              - sun
              - moon
              - lightbulb
              - office lights
```

One caveat. The "Keys" will all be lowercased due to how the data is being read in. Meaning, even though
`Title` is specified, the template will see the value under "title" instead.

### Available Functions

Additionally, there a few functions exposed and available to you that allows you to modify

```
| Function Name    | Example                                 | Input                | Output                   |
|------------------|-----------------------------------------|----------------------|--------------------------|
| ToSlug           | {{ .title \| ToSlug }}                  | Bob Candy            | bob-candy                |
| QuotedStringJoin | {{ .lightsources \| QuotedStringJoin }} | [sun,moon,lightbulb] | "sun","moon","lightbulb" |
```

There is also a large collection of functions that have been imported from [sprig](https://masterminds.github.io/sprig/)
and are available for use.

### Example Templating Snippets

Data Injection

```json
{
  "annotations": {
    "list": [
      {
        "$$hashKey": "{{ .title | lower | ToSlug}}",
        // Inserting data and piping it to two different functions.  In this case, ToLower is redundant, but it serves as a chained example.
        "builtIn": 1,
        "datasource": "Grafana",
        "enable": true,
        "hide": true,
        "iconColor": "rgba(0, 211, 255, 1)",
        "name": "Annotations Alerts",
        "type": "dashboard"
      }
    ]
  }
}
```

Iterating and conditionals.

```json
{
  "link_text": [
    {{ if .enabledlight }}
    // conditional to check if to insert or not
      {{ range $v: = .lightsources }}
        // Iterating through list
        {{ $v }}
      // Inserting value
      {{ end }}
    {{ end }}
  ]
}
```

Inserting a comma delimited list

```json
"link_url": [
  "{{ .lightsources | join ", " }}",
  "/grafana/d/000000003/bandwidth-dashboard",
  "/grafana/d/xk26IFhmk/flow-data",
]
```

### Usage

As part of the installation you will have access to gdg-generate.


{{< callout note >}}--config, --template-config, and -t are optional parameters.  gdg-generate will fallback on defaults if
none are specified.  If -t is not provided, all templates will be processed
 {{< /callout >}}

```sh
gdg-generate --config config/gdg.yml --template-config config/template.yaml template generate  -t template_example
```

Example output:

```sh
2023-11-16 09:49:03 INF gen/main.go:16 Reading GDG configuration
2023-11-16 09:49:03 INF gen/main.go:20 Configuration file is:  config=importer.yml
2023-11-16 09:49:03 INF gen/main.go:29 Context is set to:  context=testing
2023-11-16 09:49:03 INF templating/templating.go:83 Processing template template=template_example
2023-11-16 09:49:03 INF templating/templating.go:97 Creating a new template folder=General orgId=2 data="map[enabledlight:false lightsources:[sun moon lightbulb office lights] title:Bob Loves Candy]"
2023-11-16 09:49:03 INF templating/templating.go:100 Writing data to destination output=test/data/org_2/dashboards
2023-11-16 09:49:03 INF templating/templating.go:131 template Path: path=test/data/templates
```

A new file has been created under test/data/org_2/dashboards/General/template_example.json
