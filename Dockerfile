FROM golang:1.15-alpine3.12 AS dev

RUN apk add make

WORKDIR /status-reporter

COPY Makefile go.* /status-reporter/

RUN go mod download

COPY . .

RUN make build

FROM alpine:3.12

COPY --from=dev /status-reporter/dist/status-reporter /status-reporter

RUN mv /status-reporter /usr/bin/status-reporter

ENTRYPOINT [ "status-reporter" ]