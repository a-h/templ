# syntax=docker/dockerfile:1
FROM golang:1.22

ARG TEMPL_VERSION
RUN go install github.com/a-h/templ/cmd/templ@$TEMPL_VERSION

WORKDIR /app

COPY run.sh /run.sh

ENV TEMPL_VERSION=$TEMPL_VERSION
ENTRYPOINT ["/bin/bash", "-c", "/run.sh"]
