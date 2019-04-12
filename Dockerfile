FROM golang:alpine as builder
WORKDIR /go/src/github.com/mudkipme/lilycove/
RUN apk add librdkafka-dev build-base
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure --vendor-only
COPY . .
RUN GOOS=linux go build -a -o lilycove .

FROM alpine:edge
WORKDIR /root
COPY --from=builder /go/src/github.com/mudkipme/lilycove/lilycove .
RUN apk add --no-cache librdkafka
VOLUME /etc/config.toml
ENTRYPOINT ["/root/lilycove"]