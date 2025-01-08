---
title: "Globals"
weight: 101
---


Global flags are configuration that applies to all contexts.  The following flags are all nested under `globals:`

### Debug

When `debug` is set to true, verbose debugging is enabled.  Usually only needed for debugging when issues arise.

### Debug API

`debug_api` when set to true will echo out all the raw API calls, parameter and responses being received.  This can be
very helpful if you wish to debug behavior being seen or reverse engineering what GDG is actually doing.

### Ignore SSL

`ignore_ssl_errors` when set to true will accept invalid SSL certificates.

### Retry Count

`retry_count` when set will try N number of times before giving up on any request.  Please be careful if the number is too
high it can lead to very slow performance if performing several operations.

### Retry Delay

`retry_delay` when set will wait for the specified duration before trying again.  The time is parsed in the format supported
by go time.ParseDuration [package](https://pkg.go.dev/time#ParseDuration).
