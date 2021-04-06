# Quick reference

- **Maintained by**: [The Provenance Team](https://github.com/provenance-io/cosmovisor)
- **Where to file issues**: [Provenance Cosmovisor Issue Tracker](https://github.com/provenance-io/cosmovisor/issues)

# Current release tags and `Dockerfile` links:

- [`v0.3.0`](https://github.com/provenance-io/cosmovisor/blob/v0.3.0/Dockerfile)

# Historical release tags and `Dockerfile` links:

- [`v0.2.0`](https://github.com/provenance-io/cosmovisor/blob/v0.2.0/Dockerfile)
- [`v0.1.0`](https://github.com/provenance-io/cosmovisor/blob/v0.3.0/Dockerfile)

# Quick reference (cont.)

- **Supported architectures**: [`amd64`]
- **Source of this description**: [docs](https://github.com/provenance-io/cosmovisor/blob/docker/README.md)

# What is this image?

The `provenanceio/cosmovisor` images are used to quickly utilize the `cosmovisor` binary to manage upgrades of networks running on [cosmos-sdk](https://github.com/cosmos/cosmos-sdk).

# How to use this image

Running `autod` with cosmovisor, assuming ./testnet-visor/cosmovisor follows the directory structure described [here](https://github.com/provenance-io/cosmovisor#data-folder-layout).

```console
$ docker run --rm --name my-node -it \
    -v $(pwd)/testnet-visor:/cosmovisor \
    -v $(pwd)/testnet-home:/home/autod \
    -e DAEMON_HOME=/cosmovisor \
    -e DAEMON_NAME=autod \
    -e DAEMON_ALLOW_DOWNLOAD_BINARIES=true \
    -e DAEMON_RESTART_AFTER_UPGRADE=true \
    -e DAEMON_BACKUP_DATA_DIR=/home/autod/data \
  provenanceio/cosmovisor start --home=/home/autod
```

Using as a source image in a Dockerfile (ie: [provenance testnet node](https://github.com/provenance-io/testnet/tree/main/docker/node/visor/Dockerfile))

```dockerfile
# Pull cosmovisor docker layer.
FROM provenanceio/cosmovisor as visor

# Build out our node on a standard debian image.
FROM debian:buster-slim as node
COPY . .
RUN go build ./cmd/...
COPY --from=visor /usr/bin/cosmovisor /usr/bin/cosmovisor
# ...
# ...
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/usr/bin/cosmovisor", "start"]
```

Installing via `go get ...` within a Dockerfile.

```dockerfile
FROM golang:1.15-buster as node
COPY . .
RUN go build ./cmd/...
RUN go get github.com/provenance-io/cosmovisor/cmd/cosmovisor
# ...
# ...
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["/usr/bin/cosmovisor", "start"]
```
