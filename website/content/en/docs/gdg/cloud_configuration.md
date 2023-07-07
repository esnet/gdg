---
title: "Cloud Configuration"
weight: 15
---

# Cloud Support

Support for using a few cloud providers as a storage engine is now supported.  When enabled, the local file system is only used for reading configuration files.  Everything else relies on the data from the cloud provider matching up to your configuration.

Currently the following providers are supported:

  - AWS S3
  - Google Storage (GS)
  - Azure

{{< alert icon="👉" text="https://github.com/google/go-cloud was used to support all of these providers.  They should all work, but only S3 and Google have been properly tested. " />}}

<!-- NOTE:  the [go-cloud](https://github.com/google/go-cloud) was used to support all of these providers.  They should all work, but only S3 and Google have been properly tested. -->

Most of these rely on the system configuration.  Here are some references for each respective environment:

  * Google Storage:  
    * [https://cloud.google.com/docs/authentication#service-accounts](https://cloud.google.com/docs/authentication#service-accounts)
    * [https://cloud.google.com/docs/authentication/provide-credentials-adc#local-user-cred](https://cloud.google.com/docs/authentication/provide-credentials-adc#local-user-cred)
  * S3: [https://docs.aws.amazon.com/sdk-for-go/api/aws/session/](https://docs.aws.amazon.com/sdk-for-go/api/aws/session/)
  * Azure: [https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/storage/azblob](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/storage/azblob)


## General Configuration

```yaml
storage_engine:
  any_label:
    kind: cloud
    cloud_type: [s3, gs,  azblob]
    bucket_name: ""
    prefix: "dummy"
```

All authentication and authorization is done outside of GDG.  

## Context Configuration

The only additional change to the context is to provide a storage label to use:

```yaml 
  testing:
    output_path: testing_data
    ...
    storage: any_label
    ...
```

So given the bucket name of `foo` with a prefix of `bar` with the output_path configured as `testing_data` the connections will be imported to:

`s3://foo/bar/testing_data/connections/` and exported from the same location.  If you need it to be in a different location you can update the prefix accordingly but at destination will follow the typical app patterns.