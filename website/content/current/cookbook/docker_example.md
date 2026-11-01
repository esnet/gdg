---
title: "Docker Compose Example"
weight: 51
description: "Illustrates an example on how to use docker and taking advantage of the built in secret feature in conjunction with GDG."
---

You can configure gdg in a variety of different ways but if you want so use secret the is currently the simplest pattern:


```yaml
version: "3.8"
services:
  app:
    image: ghcr.io/esnet/gdg:0.8.0
    volumes:
      - ./gdg.yml:/app/config/gdg.yml:ro
    secrets:
      - auth_staging.yaml
      - default.yaml

secrets:
  auth_staging.yaml:
    file: ./auth_staging.yaml
  default.yaml:
    file: ./default_connection_auth.yaml

```

Then update your config file with the following:

`secure_location: /run/secrets/`

For any connection settings, you will have to additionally define all the connection settings accordingly.


