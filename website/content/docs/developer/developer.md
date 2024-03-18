---
title: "Developer Guide"
weight: 1
---

## Dependencies

Development requirements.
  - As gdg is written in [go](https://go.dev/), a current compiler is required.
  - task management tool [go-task](https://github.com/go-task/task). [Installation](https://taskfile.dev/installation/)
  - [Docker](https://www.docker.com/products/docker-desktop/) for running tests

Install the remaining dependencies via: `task install_tools`

## Building/Running gdg

Running it then should be as simple as:

```bash
$ task build_all
$ ./bin/gdg  ## main binary
$ ./bin/gdg-generate  ## Dashboard Templating engine
```

Requires [task](https://github.com/go-task/task.git) to be installed locally

## Running Tests

BasicAuth:
   `task test`
Token:
  `task test_tokens`

## Making a release

You can validate that goreleaser work appropriately via

`task release-snapshot` or `task release`

The actual release will be done once a tag has been created via the CICD pipeline, artifacts will be generated and a website update will be published.
