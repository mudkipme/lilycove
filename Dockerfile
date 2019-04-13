FROM ubuntu:18.04 as builder
WORKDIR /go/src/github.com/mudkipme/lilycove/

RUN apt-get update && apt-get install -y software-properties-common wget git
RUN add-apt-repository ppa:longsleep/golang-backports && apt-get update && apt-get install -y golang-go go-dep

RUN wget -qO - https://packages.confluent.io/deb/5.2/archive.key | apt-key add -
RUN add-apt-repository "deb [arch=amd64] https://packages.confluent.io/deb/5.2 stable main" && apt-get update && apt-get install -y librdkafka-dev
COPY Gopkg.toml Gopkg.lock ./
ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH
RUN dep ensure --vendor-only
COPY . .
RUN GOOS=linux go build -a -o lilycove .

FROM ubuntu:18.04
WORKDIR /root
COPY --from=builder /go/src/github.com/mudkipme/lilycove/lilycove .
RUN apt-get update && apt-get install -y software-properties-common wget
RUN wget -qO - https://packages.confluent.io/deb/5.2/archive.key | apt-key add -
RUN add-apt-repository "deb [arch=amd64] https://packages.confluent.io/deb/5.2 stable main" && apt-get update && apt-get install -y librdkafka-dev
RUN rm -rf /var/lib/apt/lists/*
VOLUME /etc/config.toml
ENTRYPOINT ["/root/lilycove"]
