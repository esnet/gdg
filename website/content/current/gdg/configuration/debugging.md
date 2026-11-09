---
title: "Debugging"
description: "GDG debugging and troubleshooting guide covering logging flags, HTTP traffic logging, and common configuration issues."
weight: 106
---
## Debugging / Trouble shooting

There are several configuration flags nested under `global.logging` that can be very useful to determine the issue.

```yaml
global:
  logging:
    verbose: true
    http_traffic:
      headers: true
      body: true
```

`logging.verbose` enables DEBUG-level logging, which may provide insight on the core issue you're running into.
`logging.http_traffic.headers` and `logging.http_traffic.body` control whether the headers and body of every HTTP
request and response are printed.  You can enable either or both independently.

For example, attempting to upload all the given folders I get the follow response.

{{% expand title="API Logs" %}}
```sh
POST /api/folders HTTP/1.1
Host: localhost:3000
User-Agent: Go-http-client/1.1
Content-Length: 36
Accept: application/json
Authorization: Basic YWRtaW46YWRtaW4=
Content-Type: application/json
X-Grafana-Org-Id: 1
Accept-Encoding: gzip

{"title":"Other","uid":"CWSuYt_nk"}

HTTP/1.1 409 Conflict
Content-Length: 80
Cache-Control: no-store
Content-Type: application/json
Date: Wed, 11 Sep 2024 17:07:57 GMT
X-Content-Type-Options: nosniff
X-Frame-Options: deny
X-Xss-Protection: 1; mode=block

{"message":"a folder with the same name already exists in the current location"}
```

{{% /expand %}}

As you can see from the logs, a POST request was made to `http://localhost:3000/api/folders`, the payload was
```json
{"title":"Other","uid":"CWSuYt_nk"}
```

It returned a 409 status since the folder already exists and the response was:
```json
{"message":"a folder with the same name already exists in the current location"}
```

