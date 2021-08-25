---
title: "Example Usage"
weight: 14
---

### Import / Download Dashboards

Minimal configuration (eg. the `importer.yml` file) that you need to download your dashboards from your Grafana endpoint:
```
context_name: all

contexts:
  all:
    url: https://grafana.example.org
    token: "<<Grafana API Token>>"
    dashboards_output: "dashboards"
    datasources_output: "dashboards"
    watched:
      - Example
      - Infrastructure

global:
  debug: true
  ignore_ssl_errors: false
```
You need to adjust three parts in the configuration in order to function:
- Grafana URL: This is just a URL where your Grafana is available.
- API Key: In your Grafana Web UI, go to Configuration, and API Keys. There you can create a new API token, which will be then used to authenticate with the Grafana API. Generate it, and paste it into the your configuration file (eg. `importer.yml`).
- Downloaded Folders: The `watched` field defines folders which will be considered for manipulation. You can see these folders in your Grafana Web UI, under Dashboards > Management. From there, you can simply define the folders you want to be downloaded in the `watched` list.

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
You can execute `./bin/gdg dash import`, and you will be able to find your dashboards, the ones which were listed previously in the table, in `dashboards` folder.

### Export / Upload Dashboards

Minimal configuration (eg. the `importer.yml` file) that you need to upload your dashboards from your Grafana endpoint:
```
context_name: all

contexts:
  all:
    url: https://grafana.example.org
    token: "<<Grafana API Token>>"
    dashboards_output: "dashboards"
    datasources_output: "dashboards"
    watched:
      - Example
      - Infrastructure

global:
  debug: true
  ignore_ssl_errors: false
```
You need to adjust three parts in the configuration in order to function:
- Grafana URL: This is just a URL where your Grafana is available.
- API Key: In your Grafana Web UI, go to Configuration, and API Keys. There you can create a new API token, which will be then used to authenticate with the Grafana API. Generate it, and paste it into the your configuration file (eg. `importer.yml`).
- Uploaded Folders: The `watched` field defines folders which will be considered for manipulation. Those are the folders in `dashboards` folder, which contain the dashboard files (eg. `examples.json`). Example folder structure:
```
├── bin
|   └── gdg
└── dashboards
    ├── Example
    |   └── cluster-scaling.json
    └── Infrastructure
        └── aws-ecs.json
```
You can execute `./bin/gdg dash export` to upload the local dashboards to your Grafana. Afterwards, you can try running `./bin/gdg dash list` in order to confirm that your dashboards were uploaded successfully.
