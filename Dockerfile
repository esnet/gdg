# Build Stage
FROM golang:1.16.0  AS build-stage

LABEL app="build-grafana-dashboard-manager"
LABEL REPO="https://github.com/netsage-project/grafana-dashboard-manager"

ENV PROJPATH=/go/src/github.com/netsage-project/grafana-dashboard-manager

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /go/src/github.com/netsage-project/grafana-dashboard-manager
WORKDIR /go/src/github.com/netsage-project/grafana-dashboard-manager

RUN make build-alpine

# Final Stage
FROM golang:1.16.0 

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/netsage-project/grafana-dashboard-manager"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/grafana-dashboard-manager/bin

WORKDIR /opt/grafana-dashboard-manager/bin

COPY --from=build-stage /go/src/github.com/netsage-project/grafana-dashboard-manager/bin/gdg /opt/grafana-dashboard-manager/bin/
RUN \
    apt-get update && \
    apt install -y dumb-init  && \
    apt-get clean autoclean && \
    apt-get autoremove --yes && \
    rm -rf /var/lib/{apt,dpkg,cache,log}/ && \
    chmod +x /opt/grafana-dashboard-manager/bin/gdg

# Create appuser
RUN useradd -m  grafana-dashboard-manager
USER grafana-dashboard-manager

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/opt/grafana-dashboard-manager/bin/gdg"]
