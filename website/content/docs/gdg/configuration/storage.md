---
title: "Storage"
weight: 102
---

This is only needed if you intend to use a cloud provider to store your backups.  If you are downloading your backups to your local file system you can skip this section.

GDG should work with most S3 compatible providers.  It leverages the [go-cloud](https://github.com/google/go-cloud) framework to provide this functionality. The standard providers should 'just work' out of the box relying on the authentication mechanism that each provider supports.  ie.  AWS looks for ~/.aws/credentials, google will look for its respective config and so on. GDG also supports custom providers that allows for S3 compatible self hosted solutions such as Minio, Ceph, etc.

All configuration below fall under the `storage_engine` section, where a new label is introduced for each provider you would like to define.  The value doesn't matter but you'll need to reference it in the `context` section.

## Simple Cloud Storage

These would be S3, AWS, Azure.


### Kind

The flag `kind` is now deprecated, but if you need to set it, it should always be set to 'cloud'.

### Cloud Type

`cloud_type` should be set to the provider you wish to use.  Please note, that custom requires additional values to be set.

Supported values are:

  - 's3' -  AWS S3
  - 'gs' - Google Storage (GS)
  - 'azblob' - Azure Storage
  - custom (S3 Compatible clouds)

{{< callout context="caution" title="Caution" icon="alert-triangle" >}}
https://github.com/google/go-cloud was used to support all of these providers.  They should all work, but only S3 and Google have been fully tested.
{{< /callout >}}

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
    kind: cloud
    cloud_type: [s3, gs,  azblob]
    bucket_name: ""
    prefix: "dummy"
```


## Other Properties

The rest of these properties listed below are ignore by any of the standard providers and are intended only to be used by the custom type.  Note you can configure any provider as 'custom' but you'll need to set far more properties as well as store your credentials in the config.  It's likely both a better pattern and more secure to rely on the cloud provider auth mechanism.


### Access Id

`access_id` is the access id used to authenticate.  This can be a username, or a key depending on the provider.  It's not typically seen a secret.

### Secret Key

`secret_key` is the secret used to valid access.  This is a sensitive credential and should not be shared.

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


## Context Configuration

In the context,  you will need to set the `storage` value to the name of the label you defined in the storage section.
