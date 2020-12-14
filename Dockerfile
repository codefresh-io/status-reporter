FROM golang:1.15-alpine3.12 AS dev

RUN apk add make

WORKDIR /go/src/github.com/codefresh-io/status-reporter

ENV GO111MODULE=on
ENV GOLANGS=-mod=vendor

ARG GO111MODULE=on
ARG GOLANGS=-mod=vendor

COPY Makefile go.* ./
COPY . .

RUN make build

FROM alpine:3.12

COPY --from=dev /go/src/github.com/codefresh-io/status-reporter/dist/status-reporter /status-reporter

RUN mv /status-reporter /usr/bin/status-reporter

ENTRYPOINT [ "status-reporter" ]