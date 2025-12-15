---
title: "Structure"
weight: 100
---

GDG Configuration is based on a YAML configuration.  It also has the option to override configuration values via environmental variables.

The default config file is gdg.yml which should exist in $CWD/config, $CWD or /etc/gdg.  If no valid file can be found, the application will fail.  An example config file can be found in [github](https://github.com/esnet/gdg/blob/main/config/gdg-example.yml) under the config folder.

The configuration has a few sections of note:
1. context_name: the selected context
2. Storage -- defines the storage engine if a cloud provider is used
3. contexts: defines every context you operate on, think of it as the definition of a grafana instance and how to connect to it.  It also defines behavior for what folders to inspect, authentication mechanism, etc.
4. global: defines global behavior for all contexts.  These settings are not context specific.  They tend to be things like enabling debugging, retry logic, api debugging etc.

## Environment Overrides

{{< callout context="caution" title="Caution" icon="alert-triangle" >}}
Complex data type is not supported. If the value is an array it can't be currently set, via ENV overrides.

{{< /callout >}}

If you wish to override certain value via the environment, like credentials and such you can do so.

The pattern for GDG's is as follows:  `GDG_SECTION__SECTION__keyname`

For example if I want to set the context name to a different value I can use:

```sh
GDG_CONTEXT_NAME="testing" gdg tools ctx show ## Which will override the value from the context file.
GDG_CONTEXTS__TESTING__URL="www.google.com" Will override the URL with the one provided.
 ```
