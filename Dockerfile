FROM alpine:latest as build-env

RUN apk add --update --no-cache curl jq

WORKDIR /app

FROM alpine:latest

RUN apk add --update --no-cache ca-certificates tzdata

WORKDIR /app

COPY static /app/static
COPY edit.html /app/edit.html
COPY bin/server /app/bin/server

RUN addgroup -g 1001 codehex && adduser -D -G codehex -u 1001 codehex
USER 1001

EXPOSE 8080
CMD ["/app/bin/server"]