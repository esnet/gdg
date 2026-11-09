---
title: "Globals"
description: "Reference for GDG global configuration flags that apply across all contexts, including logging, SSL, retry, and output settings."
weight: 101
---


Global flags are configuration that applies to all contexts.  The following flags are all nested under `global:`

### Clear Output

`clear_output` when set to true will remove all files in the destination folder prior to importing the data.

ie.  if fetching all dashboards for the default org, then all files in the folder: `{output_path}/org_main-org/dashboards` will be removed.

{{% notice style="warning" title="Caution" %}}
Be careful when using this pattern.  This removes all related files, if the operation fails all previous backups for that entity type will be lost.
{{% /notice %}}

### Logging

All logging options are grouped under `global.logging:`.

#### Verbose

`logging.verbose` when set to `true` enables DEBUG-level logging.  Usually only needed when diagnosing issues.

#### HTTP Traffic

`logging.http_traffic` controls whether outbound HTTP requests and their responses are dumped to the log.  The two
sub-fields are independent so you can include just the headers, just the body, or both:

| Field | Description |
|---|---|
| `logging.http_traffic.headers` | Log request and response headers |
| `logging.http_traffic.body` | Log request and response bodies |

```yaml
global:
  logging:
    verbose: true
    http_traffic:
      headers: true
      body: true
```

Both default to `false`.  Enabling `body` can produce very large log output for endpoints that return dashboard JSON.

### Ignore SSL

`ignore_ssl_errors` when set to true will accept invalid SSL certificates.

### Retry Count

`retry_count` when set will try N number of times before giving up on any request.  Please be careful if the number is too
high it can lead to very slow performance if performing several operations.

### Retry Delay

`retry_delay` when set will wait for the specified duration before trying again.  The time is parsed in the format supported
by go time.ParseDuration [package](https://pkg.go.dev/time#ParseDuration).
