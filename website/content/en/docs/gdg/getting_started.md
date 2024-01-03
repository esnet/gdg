---
title: "Getting Started"
weight: 13
---

### Setup new configuration

You can create new context configuration using an interactive setup.
```
$ ./bin/gdg tools ctx new mycontext
```

When creating a new context, you will be asked for authorization type, your default datasource and username/password, along with which folders you wish to manage under the context. You have three options:

1. Default option ("General")
2. List of folders you wish to manage
3. Wildcard configuration (all folders)

### Import / Download Dashboards

Minimal configuration (eg. the `importer.yml` file) that you need to download your dashboards from your Grafana endpoint:
```
context_name: all

contexts:
  all:
    url: https://grafana.example.org
    token: "<<Grafana API Token>>"
    # user_name: admin
    # password: admin
    output_path: exports
    watched:
      - Example
      - Infrastructure

global:
  debug: true
  ignore_ssl_errors: false
```
You need to adjust three parts in the configuration in order to function:
- Grafana URL: This is just a URL where your Grafana is available.
- API Key OR Username / Passoword for Admin user. See [authentication](configuration.md) section if you need more information.
- Downloaded Folders: The `watched` field defines folders which will be considered for manipulation. You can see these folders in your Grafana Web UI, under Dashboards > Management. From there, you can simply define the folders you want to be downloaded in the `watched` list. The dashboards are downloaded as JSON files in the `$OUTPUT_PATH/dashboards/$GRAFANA_FOLDER_NAME` directory. Where `$OUTPUT_PATH` is the path defined in the `dashboard_output` configuration property and `$GRAFANA_FOLDER_NAME` the name of the folder from which the dashboards were downloaded

After you are done, and you can execute `./bin/gdg dash list` successfully, eg.:
```
$ ./bin/gdg dash list
time="2021-08-22T11:11:27+02:00" level=warning msg="Error getting organizations: HTTP error 403: returns {\"message\":\"Permission denied\"}"
time="2021-08-22T11:11:28+02:00" level=info msg="Listing dashboards for context: 'all'"
┌────┬───────────────────────────────────┬───────────────────────────────────┬────────────────┬────────────┬────────────────────────────────────────────────────────────────────────────┐
│ ID │ TITLE                             │ SLUG                              │ FOLDER         │ UID        │ URL                                                                        │
├────┼───────────────────────────────────┼───────────────────────────────────┼────────────────┼────────────┼────────────────────────────────────────────────────────────────────────────┤
│  8 │ AWS CloudWatch Logs               │ aws-cloudwatch-logs               │ Infrastructure │ AWSLogs00  │ https://grafana.example.org/d/AWSLogs00/aws-cloudwatch-logs                │
│  6 │ AWS ECS                           │ aws-ecs                           │ Infrastructure │ ly9Y95XWk  │ https://grafana.example.org/d/ly9Y95XWk/aws-ecs                            │
│  5 │ AWS ELB Application Load Balancer │ aws-elb-application-load-balancer │ Infrastructure │ bt8qGKJZz  │ https://grafana.example.org/d/bt8qGKJZz/aws-elb-application-load-balancer  │
│  4 │ AWS RDS                           │ aws-rds                           │ Infrastructure │ kCDpC5uWk  │ https://grafana.example.org/d/kCDpC5uWk/aws-rds                            │
│  3 │ AWS S3                            │ aws-s3                            │ Infrastructure │ AWSS31iWk  │ https://grafana.example.org/d/AWSS31iWk/aws-s3                             │
│ 17 │ Cluster Autoscaling               │ cluster-autoscaling               │ Example        │ iHUYtABMk  │ https://grafana.example.org/d/iHUYtABMk/cluster-autoscaling                │
└────┴───────────────────────────────────┴───────────────────────────────────┴────────────────┴────────────┴────────────────────────────────────────────────────────────────────────────┘
```
After executing `./bin/gdg dash import` you can find the dashboards of the `Infrastructure` folder in the local directory `dashboards/dashboards/Infrastructure` and the dashboards of the `Example` directory in the local directory `dashboards/dashboards/Example`.

### Export / Upload Dashboards

Minimal configuration (eg. the `importer.yml` file) that you need to upload your dashboards from your Grafana endpoint:
```
context_name: all

contexts:
  all:
    url: https://grafana.example.org
    token: "<<Grafana API Token>>"
    # user_name: admin
    # password: admin
    output_path: exports
    watched:
      - Example
      - Infrastructure

global:
  debug: true
  ignore_ssl_errors: false
```
You need to adjust three parts in the configuration in order to function:
- Grafana URL: This is just a URL where your Grafana is available.
- API Key OR Username / Passoword for Admin user. See [authentication](configuration.md) section if you need more information.
- Uploaded Folders: The `watched` field defines folders which will be considered for manipulation. The dashboards should be stored as JSON files in the `$OUTPUT_PATH/dashboards/$GRAFANA_FOLDER_NAME` directory. Where `$OUTPUT_PATH` is the path defined in the `dashboard_output` configuration property and `$GRAFANA_FOLDER_NAME` the name of the folder to which the dashboards will be uploaded. In case of the above configuration file, the dashboards should be stored locally in the `dashboards/dashboards/Example` and `dashboards/dashboards/Infrastructure/` directories.

```
├── bin
|   └── gdg
└── exports
    └── dashboards
        ├── Example
        |   └── cluster-scaling.json
        └── Infrastructure
            └── aws-ecs.json
```
You can execute `./bin/gdg dash export` to upload the local dashboards to your Grafana. Afterwards, you can try running `./bin/gdg dash list` in order to confirm that your dashboards were uploaded successfully.
