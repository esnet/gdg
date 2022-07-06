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
  - SFTP (Not exactly cloud, but useful)

NOTE:  the [stow](https://github.com/graymeta/stow) was used to support all of these providers.  They should all work, but only S3 and Google have been properly tested.

## General Configuration

```yaml
storage_engine:
  any_label:
    kind: cloud
    cloud_type: [s3, google, azure, sftp]
    bucket_name: ""
    prefix: "dummy"
```

Additional configuration for the respective engines are dependent on the cloud providers.

S3 [config](https://github.com/graymeta/stow/blob/master/s3/config.go):
```yaml 
    access_key_id: ""
    secret_key:  ""
```

Google [config](https://github.com/graymeta/stow/blob/master/google/config.go):

```yaml 
        project_id: esnet-sd-dev
        json: keys/service.json
```

Azure [config](https://github.com/graymeta/stow/blob/master/azure/config.go):

```yaml 
    account: ""
    key: ""
```

SFTP [config](https://github.com/graymeta/stow/blob/master/sftp/config.go):

```yaml 
    host:
    port:
    password:
    private_key:
    base_path:
    host_public_key:

```

## Context Configuration

The only additional change to the context is to provide a storage label to use:

```yaml 
  testing:
    output_path: testing_data
    ...
    storage: any_label
    ...
```

So given the bucket name of `foo` with a prefix of `bar` with the output_path configured as `testing_data` the datasources will be imported to:

`s3://foo/bar/testing_data/datasources/` and exported from the same location.  If you need it to be in a different location you can update the prefix accordingly but at destination will follow the typical app patterns.