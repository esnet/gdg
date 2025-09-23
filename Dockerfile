FROM alpine:latest
ARG TARGETPLATFORM
RUN mkdir /app && apk add bash
COPY $TARGETPLATFORM/gdg /app/gdg
COPY $TARGETPLATFORM/gdg-generate /app/gdg-generate
VOLUME /app/config
VOLUME /app/exports

WORKDIR /app 
ENTRYPOINT ["/app/gdg"]
