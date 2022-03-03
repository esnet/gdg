# Build Stage
FROM golang:1.17.0  AS build-stage

LABEL app="build-gdg"
LABEL REPO="https://github.com/netsage-project/gdg"

ENV PROJPATH=/go/src/github.com/netsage-project/gdg

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /go/src/github.com/netsage-project/gdg
WORKDIR /go/src/github.com/netsage-project/gdg

RUN make build-alpine

# Final Stage
FROM golang:1.17.0 

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/netsage-project/gdg"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/gdg/bin

WORKDIR /opt/gdg/bin

COPY --from=build-stage /go/src/github.com/netsage-project/gdg/bin/gdg /opt/gdg/bin/
RUN \
    apt-get update && \
    apt install -y dumb-init  && \
    apt-get clean autoclean && \
    apt-get autoremove --yes && \
    rm -rf /var/lib/{apt,dpkg,cache,log}/ && \
    chmod +x /opt/gdg/bin/gdg

# Create appuser
RUN useradd -m  gdg
USER gdg

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/opt/gdg/bin/gdg"]
