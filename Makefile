
.PHONY: build build-alpine clean test help default



BIN_NAME=gdg

VERSION := $(shell grep "const Version " version/version.go | sed -E 's/.*"(.+)"$$/\1/')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
BUILD_DATE=$(shell date '+%Y-%m-%d-%H:%M:%S')
IMAGE_NAME := "esnet/gdg"



default: build

authors:
	echo "Authors\n=======\n" > AUTHORS.md
	git log --raw | grep "^Author: " | sort | uniq | cut -d ' ' -f2- | sed 's/^/- /' >> AUTHORS.md


linux: clean 
	env GOOS='linux' GOARCH='amd64' go build -ldflags "-X github.com/esnet/gdg/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/esnet/gdg/version.BuildDate=${BUILD_DATE}" -o bin/${BIN_NAME}_linux

help:
	@echo 'Management commands for gdg:'
	@echo
	@echo 'Usage:'
	@echo '    make build             Compile the project.'
	@echo '    make get-deps          runs dep ensure, mostly used for ci.'
	@echo '    make build-alpine      Compile optimized for alpine linux.'
	@echo '    make package           Build final docker image with just the go binary inside'
	@echo '    make release-snapshot  Test Release locally using goreleaser'
	@echo '    make release           Push Release to github'
	@echo '    make tag               Tag image created by package with latest, git commit and version'
	@echo '    make test              Run tests on a compiled project.'
	@echo '    make push              Push tagged images to registry'
	@echo '    make clean             Clean the directory tree.'
	@echo

build:
	@echo "building ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	go build -ldflags "-X github.com/esnet/gdg/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/esnet/gdg/version.BuildDate=${BUILD_DATE}" -o bin/${BIN_NAME}

install:
	@echo "installing ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	go install -ldflags "-X github.com/esnet/gdg/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/esnet/gdg/version.BuildDate=${BUILD_DATE}"
	mv ${GOPATH}/bin/gdg ${GOPATH}/bin/${BIN_NAME}

get-deps:
	go mod tidy

build-alpine:
	@echo "building ${BIN_NAME} ${VERSION}"
	@echo "GOPATH=${GOPATH}"
	go build -ldflags '-w -linkmode external -extldflags "-static" -X github.com/esnet/gdg/version.GitCommit=${GIT_COMMIT}${GIT_DIRTY} -X github.com/esnet/gdg/version.BuildDate=${BUILD_DATE}' -o bin/${BIN_NAME}

package:
	@echo "building image ${BIN_NAME} ${VERSION} $(GIT_COMMIT)"
	docker build --build-arg VERSION=${VERSION} --build-arg GIT_COMMIT=$(GIT_COMMIT) -t $(IMAGE_NAME):local .

tag: 
	@echo "Tagging: latest ${VERSION} $(GIT_COMMIT)"
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):$(GIT_COMMIT)
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):${VERSION}
	docker tag $(IMAGE_NAME):local $(IMAGE_NAME):latest

push: tag
	@echo "Pushing docker image to registry: latest ${VERSION} $(GIT_COMMIT)"
	docker push $(IMAGE_NAME):$(GIT_COMMIT)
	docker push $(IMAGE_NAME):${VERSION}
	docker push $(IMAGE_NAME):latest

clean:
	@test ! -e bin/${BIN_NAME} || rm bin/${BIN_NAME}
	rm -fr dist/

release-snapshot: clean 
	goreleaser build --snapshot

release: clean 
	goreleaser release

test:
	go test -v ./... -cover

