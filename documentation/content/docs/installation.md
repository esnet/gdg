---
title: "Installation"
weight: 1
---
## Installation

The easiest way to install GDG is to get one of the pre-compiled binaries from our release page which can be found [here](https://github.com/esnet/gdg/releases).  Packages are not yet supported but will be coming soon since goreleaser has that feature.

Planned package [support](https://github.com/esnet/gdg/issues/89) for:
  - HomeBrew
  - Debian
  - RPMs
  - APK

If you have go install you may run the following command to install gdg

```sh 
go install github.com/esnet/gdg@latest
```

You can verify the version by running `gdg version`.

## Configuration

You can then create a simple configuration using `gdg ctx new` which will do a best effort to guide to setup a basic config that will get you up and going or read the more detailed documentation that can be found [here](/gdg/docs/configuration/)


