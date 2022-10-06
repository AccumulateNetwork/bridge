FROM golang:1.18 AS builder

ARG GOBIN=/go/bin/
ARG GOOS=linux
ARG GOARCH=amd64
ARG CGO_ENABLED=0
ARG GO111MODULE=on
ARG PKG_NAME=github.com/AccumulateNetwork/bridge
ARG PKG_PATH=${GOPATH}/src/${PKG_NAME}

WORKDIR ${PKG_PATH}
COPY . ${PKG_PATH}/

RUN go mod download && \
  go build -o /go/bin/bridge main.go

FROM alpine:3

RUN set -xe && \
  apk --no-cache add bash ca-certificates inotify-tools && \
  addgroup -g 1000 app && \
  adduser -D -G app -u 1000 app

WORKDIR /home/app

COPY --from=builder /go/bin/bridge ./
COPY ./entrypoint.sh ./entrypoint.sh

RUN \
  mkdir ./values && \
  chown -R app:app /home/app

RUN chmod +x ./entrypoint.sh

USER app

EXPOSE 8081

ENTRYPOINT [ "./entrypoint.sh" ]

CMD [ "./bridge", "-c", "/home/app/values/config.yaml" ]