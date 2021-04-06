#!/usr/bin/make -f

CWD=$(shell pwd)
BUILDDIR=$(CWD)/build
TARGET=cosmovisor

all: build test

.PHONY: sums
sums:
	scripts/update-sums.sh

.PHONY: build
build:
	go build -mod=readonly -o $(BUILDDIR)/$(TARGET) ./cmd/$(TARGET)

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
