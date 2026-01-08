---
title: "Secure Fields"
weight: 105
---
# Secure Fields

```yaml
secure_config:
  alerting:
    - '#.receivers.#.settings.url'
    - '#.receivers.#.settings.password'
    - '#.receivers.#.settings.token'
```

Secure fields are regex patterns that should be matched against the payload of a given entity. GDG itself needs to be
updates as this is not natively applied to every resource. This might be changed in a future version but there is something
to be said for speed vs flexibility.

Currently, the only supported entity is contact-points with alerting. The regex patterns above are ones that I've seen
containing sensitive date. It's likely not a comprehensive list, but if you are happy with the list you won't need to modify it.

The default [secure.yml](https://github.com/esnet/gdg/blob/main/config/secure.yml) can be found on github. That is loaded
by default then any additional data found in your config file would override all the given values.

The regex paths are based on [gjson](https://github.com/tidwall/gjson) and [sjson](https://github.com/tidwall/sjson) which
are used to find and replace the matching values with the encoded counterpart.
