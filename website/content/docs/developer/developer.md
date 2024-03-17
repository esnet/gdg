---
title: "Developer Guide"
weight: 19
---

## Dependencies

Development requirements.
  - As gdg is written in [go](https://go.dev/), a current compiler is required.
  - task management tool [go-task](https://github.com/go-task/task). [Installation](https://taskfile.dev/installation/)
  - [Docker](https://www.docker.com/products/docker-desktop/) for running tests

Install the remaining dependencies via: `task install_tools`

## Running Tests

BasicAuth:
   `task test`
Token:
  `task test_tokens`

## Making a release

Install goreleaser.

```sh
brew install goreleaser/tap/goreleaser
brew reinstall goreleaser`
```

Alternatively if you have a more recent version of Go.

```sh
go install github.com/goreleaser/goreleaser@latest
```

export your GITHUB_TOKEN.

```sh
export GITHUB_TOKEN="secret"
```

git tag v0.1.0
goreleaser release


NOTE: CI/CD pipeline should do all this automatically.  `make release-snapshot` is used to test the release build process.  Once a build is tagged all artifacts should be built automatically and attached to the github release page.

NOTE: mac binary are not signed so will likely complain.




