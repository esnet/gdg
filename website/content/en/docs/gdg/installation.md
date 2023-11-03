---
title: "Installation"
weight: 1
---
## Installation

The easiest way to install GDG is to get one of the pre-compiled binaries from our release page which can be found [here](https://github.com/esnet/gdg/releases).  Packages for a few distributions have been added.  The release cycle relies on goreleaser so anything that is well supported can be added it at some point.  There is no APT or such you can connect to but the packages are available for download.

The following packages are currently supported:
  - RPM
  - APK
  - Docker

### Package Installation

Install from package involves downloading the appropriate package from the [release](https://github.com/esnet/gdg/releases) and installing it as you usually do on your favorite Distro.

```sh
rpm -Uvh ./gdg_0.3.1_amd64.rpm
dpkg -i ./gdg_0.3.1_amd64.deb
```

### Docker usage

The docker tags are released started with 0.3.1.  Each release will generate a major version and minor version tag.

You can see the available images [here](https://github.com/esnet/gdg/pkgs/container/gdg)

```sh
docker pull ghcr.io/esnet/gdg:0.3.1
```

NOTE: ghcr.io/esnet/gdg:0.3 will also point to 0.3.1 until 0.3.2 is released after which it'll point to 0.3.2

Example compose.

```yaml
version: '3.7'
services:
  gdg:
    image:  ghcr.io/esnet/gdg:0.3.1
    command: "--help"            ## Add additional parameters here
#    command: "ds export"       ## Pass any cmd on here.
    volumes:
      - ./conf:/app/conf         ## where the configuration lives
      - ./exports:/app/exports  ## doesn't need to be /app/exports but you should export the destination of where exports are being written out to.
```

From the CLI:

```sh
docker run -it --rm -v $(pwd)/conf:/app/conf -v $(pwd)/exports:/app/exports ghcr.io/esnet/gdg:latest  ds --help
```

### Installing via Go

If you have go install you may run the following command to install gdg

```sh
go install github.com/esnet/gdg@latest  #for latest
go install github.com/esnet/gdg@v0.3.1  #for a specific version
```

You can verify the version by running `gdg version`.

## Configuration

You can then create a simple configuration using `gdg ctx new` which will do a best effort to guide to setup a basic config that will get you up and going or read the more detailed documentation that can be found [here](/gdg/docs/gdg/configuration/)


**NOTE**: wizard doesn't currently support ALL features but it should help you get a head start.
