FROM alpine:latest
RUN mkdir /app
COPY gdg /app/gdg
VOLUME /app/conf
VOLUME /app/exports

WORKDIR /app 
ENTRYPOINT ["/app/gdg"]
