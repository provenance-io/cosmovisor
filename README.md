# Status

[![Apache 2.0 License][license-badge]][license-url]

[license-badge]: https://img.shields.io/github/license/provenance-io/cosmovisor.svg
[license-url]: https://github.com/provenance-io/cosmovisor/blob/main/LICENSE

# Provenance Blockchain fork of Cosmos-SDK's Cosmosvisor

This repo has been archived and should no longer be used.

If you are currently using a version of Cosmovisor built by this repo, you should switch to the one built and maintained by Cosmos-SDK: https://github.com/cosmos/cosmos-sdk/tree/main/tools/cosmovisor

<!-- TOC -->
  - [Migrating to the SDK's version](#migrating-to-the-sdk-s-version)
  - [Differences](#differences)
    - [Invocation](#invocation)
    - [New Options](#new-options)
    - [Version](#version)

## Migrating to the SDK's version

The file structure and most environment variables are the same.
However, the command to invoke `provenanced` using `cosmovisor` has changed slightly.

To switch to the SDK's version:

1. Stop your node.
2. Uninstall your current `cosmovisor` executable.
3. Install `cosmovisor` fom the SDK ([Instructions](https://github.com/cosmos/cosmos-sdk/tree/main/tools/cosmovisor#installation)).
4. If you use the `DAEMON_BACKUP_DATA_DIR` environment variable, change its name to `DAEMON_DATA_BACKUP_DIR`.
5. Update your execution commands to start with `cosmovisor run` instead of just `cosmovisor`. For example, if you currently execute `cosmovisor start`, it must be changed to `cosmovisor run start`.
6. Restart your node.

## Differences

### Invocation

The SDK's version now has a `run` sub-command that must be used when invoking the configured executable (e.g. `provenanced`).
All arguments provided to `cosmovisor run <args>` are provided to the configured executable the same way that this version behaves with just `cosmovisor <args>`.

For example, the following commands are equivalent:
* Without `cosmovisor`: `provenanced start`
* This version: `cosmovisor start`
* SDK's version: `cosmovisor run start`

### New Options

The SDK's version has some options that were not available in this version.
See [Command Line Arguments And Environment Variables](https://github.com/cosmos/cosmos-sdk/tree/main/tools/cosmovisor#command-line-arguments-and-environment-variables) for details.

The following environment variables were not available in this version but are options in the SDK's version:

* DAEMON_RESTART_DELAY
* UNSAFE_SKIP_BACKUP
* DAEMON_POLL_INTERVAL
* DAEMON_PREUPGRADE_MAX_RETRIES
* COSMOVISOR_DISABLE_LOGS

You can view your configuration by running the `cosmovisor config` command in the environment where you usually run `cosmovisor`.

### Version

Using this version, running the command `DAEMON_INFO=1 cosmovisor` would ouput version information.
The SDK's version does not do this, but has a `cosmovisor version` command instead.
