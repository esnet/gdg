FROM alpine:latest
RUN mkdir /app
COPY gdg /app/gdg
COPY gdg-generate /app/gdg-generate
VOLUME /app/config
VOLUME /app/exports

WORKDIR /app 
ENTRYPOINT ["/app/gdg"]
