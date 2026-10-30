---
title: "Storage"
description: "GDG storage backend configuration for S3, GCS, Azure Blob, and custom S3-compatible providers like MinIO and Ceph."
weight: 102
---

This is only needed if you intend to use a cloud provider to store your backups.  If you are downloading your backups to your local file system you can skip this section.

GDG should work with most S3 compatible providers.  It leverages the [go-cloud](https://github.com/google/go-cloud) framework to provide this functionality. The standard providers should 'just work' out of the box relying on the authentication mechanism that each provider supports.  ie.  AWS looks for ~/.aws/credentials, google will look for its respective config and so on. GDG also supports custom providers that allows for S3 compatible self-hosted solutions such as Minio, Ceph, etc.

All configuration below fall under the `storage_engine` section, where a new label is introduced for each provider you would like to define.  The value doesn't matter, but you'll need to reference it in the `context` section.

## Simple Cloud Storage

These would be S3, AWS, Azure.

### Cloud Type

`cloud_type` should be set to the provider you wish to use.  Please note, that custom requires additional values to be set.

Supported values are:

  - 's3' -  AWS S3
  - 'gs' - Google Storage (GS)
  - 'azblob' - Azure Storage
  - custom (S3 Compatible clouds)

{{% notice style="warning" title="Caution" %}}
https://github.com/google/go-cloud was used to support all of these providers.  They should all work, but only S3 and Google have been fully tested.
{{% /notice %}}

Most of these rely on the system configuration.  Here are some references for each respective environment:

  * Google Storage:
    * [https://cloud.google.com/docs/authentication#service-accounts](https://cloud.google.com/docs/authentication#service-accounts)
    * [https://cloud.google.com/docs/authentication/provide-credentials-adc#local-user-cred](https://cloud.google.com/docs/authentication/provide-credentials-adc#local-user-cred)
  * S3: [https://docs.aws.amazon.com/sdk-for-go/api/aws/session/](https://docs.aws.amazon.com/sdk-for-go/api/aws/session/)
  * Azure: [https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/storage/azblob](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/storage/azblob)

### Bucket Name
`bucket_name` is the name of the bucket to use to store your data.

### Prefix

If you would like to configure a prefix, you can set a value that will be appended to the final output path.

### Standard Config Example:

```yaml
storage_engine:
  any_label:
    cloud_type: [s3, gs, azblob, custom]
    bucket_name: ""
    prefix: "dummy"
```


## Other Properties

The rest of these properties listed below are ignore by any of the standard providers and are intended only to be used by the custom type.  Note you can configure any provider as 'custom' but you'll need to set far more properties as well as store your credentials in the config.  It's likely both a better pattern and more secure to rely on the cloud provider auth mechanism.

### Authentication

The auth as of v0.9.0+, will only look for auth in the secure location defined in your context. gdg will look for a file name that follows this naming pattern: s3_`storage_name`. So if you are configuring a storage backend, and you named it super_cloud, then your config name will be called `s3_super_cloud.yml`.  Same as the authentication mechanism for grafana s3 support JSON and YAML data format. I'd recommend using yaml but if you prefer json feel free to use it. If the same file name is found with both yaml and json, then the yaml file will be used.

To override credentials via environment variables, use the GDG-namespaced pattern:

```sh
GDG_STORAGE_ENGINE__<STORAGE_NAME>__ACCESS_ID=myaccesskey
GDG_STORAGE_ENGINE__<STORAGE_NAME>__SECRET_KEY=mysecretkey
```

Where `<STORAGE_NAME>` is the label defined under `storage_engine` in upper case. See the [Environment Variables reference](../structure/#cloud--custom-s3-storage-credentials-v0100) for details.

### Init Bucket

`init_bucket` will attempt to create a bucket.  It will warn if it already exists and continue as is.

### Endpoint

`endpoint` is the endpoint for the cloud provider.  For S3 it's `s3.amazonaws.com`, for minio it's `localhost:9000` and so on.

### Region:

`region` is the region for the cloud provider.  If you're defining AWS S3 as a custom data type this value is important.  Otherwise the default value is us-east-1


Example of a custom config:

```yaml
storage_engine:
  some_label:
    cloud_type: custom
    prefix: dummy
    bucket_name: "mybucket"
    init_bucket: "true" ## Will attempt to create the bucket if it does not exist.
    endpoint: "http://localhost:9000"
    region: us-east-1
```


## Context Configuration

In the context,  you will need to set the `storage` value to the name of the label you defined in the storage section.
