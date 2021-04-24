FROM golang:1.16

ENV CONSUL_VERSION=1.9.5
ENV VAULT_VERSION=1.7.1

RUN  apt-get update \
     && apt-get install -y unzip \
     && go get golang.org/x/lint \
     && curl --fail -Lso consul.zip "https://releases.hashicorp.com/consul/${CONSUL_VERSION}/consul_${CONSUL_VERSION}_linux_amd64.zip" \
     && unzip consul.zip -d /usr/bin \
     && curl --fail -Lso vault.zip "https://releases.hashicorp.com/vault/${VAULT_VERSION}/vault_${VAULT_VERSION}_linux_amd64.zip" \
     && unzip vault.zip -d /usr/bin

ENV CGO_ENABLED 0
ENV GOPATH /go:/cp
