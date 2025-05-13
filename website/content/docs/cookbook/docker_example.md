---
title: "Docker Compose Example"
weight: 51
---

You can configure gdg in a variety of different ways but if you want so use secret the is currently the simplest pattern:


```yaml
version: "3.8"
services:
  app:
    image: ghcr.io/esnet/gdg:0.7.2
    command: ["-c", "/run/secrets/config.yml", "tools", "contexts", "show"]
    secrets:
      - config.yml

secrets:
  config.yml:
    file: ./importer.yml
```

alternatively you can create a bash script that replaces the entrypoint as can be seen below:

```
#!/bin/sh
GDG_CONTEXTS__DEV__PASSWORD=`cat "$GF_SECURITY_ADMIN_PASSWORD__FILE"` exec /app/gdg "$@"
```

Note, bash is not in previous containers, it will be added to later versions so you'll have to use sh instead of bash.


This assumes that your grafana passwor is in a file called `gf_passwd`

```yaml
services:
  app:
    image: ghcr.io/esnet/gdg:0.7.2
    entrypoint: ["/bin/sh", "/app/wrapper.sh"]
    command: ["tools", "contexts", "show"]
    volumes:
      - ./importer.yml:/app/config/importer.yml:ro
      - ./wrapper.sh:/app/wrapper.sh
    environment:
      - GF_SECURITY_ADMIN_PASSWORD__FILE=/run/secrets/gf_passwd
    secrets:
      - gf_passwd

secrets:
  gf_passwd:
    file: ./gf_passwd
```
