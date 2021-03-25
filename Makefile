#!/usr/bin/make -f


all: cosmovisor test

cosmovisor:
	go build -mod=readonly ./cmd/cosmovisor

test:
	go test -mod=readonly -race ./...

docker-build:
	docker build -t provenanceio/cosmovisor .

docker-push:
	docker push provenanceio/cosmovisor

.PHONY: all cosmovisor test
