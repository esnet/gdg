---
title: "Structure"
weight: 22
---
The structure of the templating config is all defined under the key `entities.dashboards`

The app expects to find a file named 'templates.yml'.  You can also find an example in [git](https://github.com/esnet/gdg/blob/main/config/templates-example.yml) which is likely to be more updated.

```yaml
entites:
  dashboards:
    - template_name: template_example
      output:
        - folder: "General"
          organization_name: "Main Org."
          dashboard_name: "Testing Foobar"
          template_data:
            Title: Bob Loves Candy
            enabledlight: true
            lightsources:
              - sun
              - moon
              - lightbulb
              - office lights
        - folder: "Testing"
          organization_name: Some Other Org
          dashboard_name: ""
          template_data:
            Title: Uncle McDonalds
            enabledlight: true
            lightsources:
              - sun
              - moon
```

### Template Name

`template_name` is the name of the template which will be used to generate the given dashboards.  The file is expected to live under `{output_path}/templates/{template_name}.go.tmpl`.  The output_path would be the same the one configured for the main gdg app.

# Output

Output contains an array of one or more configuration that defines the expected behavior.  The same template file can be used to write multiple dashboards to various locations under multiple orgs and so on.

### Folder

`folder` defines the dashboard location in grafana.

### Organization Name

`organization_name` defines the name of the organization that owns the dashboard.

### Template Data

`template_data` defines the data that will be used to generate the dashboard.  This is unstructured, so any valid yaml can be specified here.  The values in template data will be used in the template file.  For example above template data contains a field named `Title` which will be used in the template file to set the dashboard title.

