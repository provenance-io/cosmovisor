#!/usr/bin/make -f


all: cosmovisor test

sums:
	scripts/update-sums.sh

cosmovisor:
	go build -mod=readonly ./cmd/cosmovisor

test:
	go test -mod=readonly -race ./...

docker-build:
	docker build -t provenanceio/cosmovisor .

docker-push:
	docker push provenanceio/cosmovisor

.PHONY: all cosmovisor test
