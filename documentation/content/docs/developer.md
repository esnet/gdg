---
title: "Developer Guide"
weight: 16
---
## Running Tests
1. Bring up a grafana instance locally with default credentials of admin/admin.  `docker-compose up -d grafana`
2. Once the instance is up simply run `go test ./...` or `make test`

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




