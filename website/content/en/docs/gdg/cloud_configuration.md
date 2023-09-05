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
  - Custom (S3 Compatible clouds)

{{< alert icon="👉" text="https://github.com/google/go-cloud was used to support all of these providers.  They should all work, but only S3 and Google have been properly tested. " />}}

<!-- NOTE:  the [go-cloud](https://github.com/google/go-cloud) was used to support all of these providers.  They should all work, but only S3 and Google have been properly tested. -->

Most of these rely on the system configuration.  Here are some references for each respective environment:

  * Google Storage:
    * [https://cloud.google.com/docs/authentication#service-accounts](https://cloud.google.com/docs/authentication#service-accounts)
    * [https://cloud.google.com/docs/authentication/provide-credentials-adc#local-user-cred](https://cloud.google.com/docs/authentication/provide-credentials-adc#local-user-cred)
  * S3: [https://docs.aws.amazon.com/sdk-for-go/api/aws/session/](https://docs.aws.amazon.com/sdk-for-go/api/aws/session/)
  * Azure: [https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/storage/azblob](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/storage/azblob)


## Cloud Configuration

### General

```yaml
storage_engine:
  any_label:
    kind: cloud
    cloud_type: [s3, gs,  azblob]
    bucket_name: ""
    prefix: "dummy"
```

All authentication and authorization is done outside of GDG.

### Custom

Examples of these S3 compatible clouds would be [minio](https://min.io/product/s3-compatibility) and [Ceph](https://docs.ceph.com/en/latest/radosgw/s3/).

```yaml
storage_engine:
  some_label:
    custom: true   ## Required, if set to true most of the 'custom' configuration will be disregarded.
    kind: cloud
    cloud_type: s3
    prefix: dummy
    bucket_name: "mybucket"
    access_id: ""  ## this value can also be read from: AWS_ACCESS_KEY. config file is given precedence
    secret_key: ""  ## same as above, can be read from: AWS_SECRET_KEY with config file is given precedence.
    init_bucket: "true" ## Only supported for custom workflows. Will attempt to create a bucket if one does not exist.
    endpoint: "http://localhost:9000"
    region: us-east-1
    ssl_enabled: "false"
```

for custom cloud, the cloud type will be s3, `access_id` and `secret_key` are needed and ONLY supported for the custom cloud.  Additionally, the `custom` flag needs to be set to true.

 - `init_bucket` is another custom only feature that will attempt to create a bucket if one does not exist.
 - `endpoint` is a required parameter though it does have a fallback to localhost:9000
 - `region` defaults to us-east-1 if not configured.


## Context Configuration

This is applicable both standard clouds and cusom.  The only additional change to the context is to provide a storage label to use:

```yaml
  testing:
    output_path: testing_data
    ...
    storage: any_label
    ...
```

So given the bucket name of `foo` with a prefix of `bar` with the output_path configured as `testing_data` the connections will be imported to:

`s3://foo/bar/testing_data/connections/` and exported from the same location.  If you need it to be in a different location you can update the prefix accordingly but at destination will follow the typical app patterns.
