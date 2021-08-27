#!/usr/bin/make -f

CWD=$(shell pwd)
BUILDDIR=$(CWD)/build
TARGET=cosmovisor

COMMIT := $(shell git log -1 --format='%h')
BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
BRANCH_PRETTY := $(subst /,-,$(BRANCH))
BUILT  := $(shell date -u +%F-%T-%Z)
MODULE := $(shell go list ./)

# don't override user values
TAG_VERSION = $(shell git describe --exact-match 2>/dev/null)
BRANCH_VERSION = $(BRANCH_PRETTY)-$(COMMIT)

ifneq ($(TAG_VERSION),)
  VERSION = $(TAG_VERSION)
else
  VERSION = $(BRANCH_VERSION)
endif

ldflags = -w -s \
    -X github.com/provenance-io/cosmovisor/version.Name=$(TARGET) \
    -X github.com/provenance-io/cosmovisor/version.Module=$(MODULE) \
    -X github.com/provenance-io/cosmovisor/version.Version=$(VERSION) \
    -X github.com/provenance-io/cosmovisor/version.Commit=$(COMMIT) \
    -X github.com/provenance-io/cosmovisor/version.Built=$(BUILT)

all: build test

.PHONY: sums
sums:
	scripts/update-sums.sh

.PHONY: build
build:
	go build -ldflags '$(ldflags)' -mod=readonly -o $(BUILDDIR)/$(TARGET) ./cmd/$(TARGET)

.PHONY: test
test:
	go test -mod=readonly -race ./...

.PHONY: docker-build
docker-build:
	docker build -t provenanceio/$(TARGET) .

.PHONY: docker-push
docker-push:
	docker push provenanceio/$(TARGET)

##############################
# Release artifacts and plan #
##############################
RELEASE_CHECKSUM_NAME=sha256sum.txt
RELEASE_CHECKSUM=$(BUILDDIR)/$(RELEASE_CHECKSUM_NAME)
RELEASE_ZIP_NAME=$(TARGET)-linux-amd64.zip
RELEASE_ZIP=$(BUILDDIR)/$(RELEASE_ZIP_NAME)

.PHONY: build-release-clean
build-release-clean:
	rm -rf $(RELEASE_CHECKSUM) $(RELEASE_ZIP)

.PHONY: build-release-checksum
build-release-checksum: build-release-zip
	cd $(BUILDDIR) && \
	  shasum -a 256 $(RELEASE_ZIP_NAME) > $(RELEASE_CHECKSUM) && \
	cd ..

.PHONY: build-release-bin
build-release-bin: build
	chmod +x $(BUILDDIR)/$(TARGET)

.PHONY: build-release-zip
build-release-zip: build-release-bin
	cd $(BUILDDIR) && \
	  zip -r $(RELEASE_ZIP_NAME) . && \
	cd ..

.PHONY: build-release
build-release: build-release-zip build-release-checksum

###########
# Linting #
###########
LINTER := $(shell command -v golangci-lint 2> /dev/null)
MISSPELL := $(shell command -v misspell 2> /dev/null)
GOIMPORTS := $(shell command -v goimports 2> /dev/null)

.PHONY: gofmt
gofmt:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*.pb.go" | xargs gofmt -s -w

.PHONY: check-goimports
check-goimports:
ifndef GOIMPORTS
	echo "Fetching goimports"
	go get golang.org/x/tools/cmd/goimports
endif

.PHONY: goimports
goimports: check-goimports
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*.pb.go" | xargs goimports -w -local github.com/provenance-io/cosmovisor

.PHONY: check-gomisspell
check-gomisspell:
ifndef MISSPELL
	echo "Fetching misspell"
	go get -u github.com/client9/misspell/cmd/misspell
endif

.PHONY: gomisspell
gomisspell: check-gomisspell
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "*.pb.go" | xargs misspell -w

.PHONY: check-lint
check-lint:
ifndef LINTER
	echo "Fetching golangci-lint"
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.39.0
endif

.PHONY: lint
lint: check-lint goimports gofmt gomisspell
	golangci-lint run

