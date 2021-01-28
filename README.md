# Pairmint

Pairmint is a high availability solution for Tendermint-based blockchain validators which uses the blockchain itself as a perfectly synchronous communication line between redundant validators running in parallel for double-signing protection.

## Requirements

* Go `v1.15+`
* Validator software compatible with Tendermint `v0.34+`

## Build & Install

Get the repository either via

```shell
$ git clone https://github.com/BlockscapeNetwork/pairmint && cd pairmint
```

or

```shell
$ go get github.com/BlockscapeLab/pairmint && cd $GOPATH/src/github.com/BlockscapeLab/pairmint
```

Once you have cloned the repository, use

```shell
$ make go-mod-cache
```

to download all the dependencies.

Pairmint can be built into the `./build` directory using

```shell
$ make build       # local os/arch
$ make build-linux # linux/amd64
```

or installed directly into `$GOPATH/bin` using

```shell
$ make install
```

## Configuration

Before putting Pairmint into operation, it needs to be initialized using:

```shell
$ pairmint init
```

The `init` command creates a `pairmint.toml` configuration file at the directory specified in the `$PAIRMINT_CONFIG_DIR` environment variable (defaults to `$HOME/.pairmint`) and a `pm-identity.key` file which holds the seed used to establish a secret connection to the validator.

> :information_source: Please look through the `pairmint.toml` file after it's generated as it is only a template and initially not valid.

If you don't already have a keypair, you can use the `--keypair` flag to generate a new `priv_validator_key.json` and `priv_validator_state.json` in your `$PAIRMINT_CONFIG_DIR` directory. Pairmint will know where to find them without you having to specify their location (see `key_file_path` and `state_file_path` parameters in the FilePV section).

If you do already have a keypair, you can either copy them into the `$PAIRMINT_CONFIG_DIR` directory or leave the key and state files where they are and specify the paths to them in the FilePV section of the `pairmint.toml`.

### Init

Init contains configuration parameters needed on initialization.

| Parameter             | Type   | Description                                                                                                                                                |
| :-------------------- | :----- | :--------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `log_level`           | string | Minimum log level for pairmint's log messages.                                                                                                             |
| `set_size`            | int    | Fixed size of the pairminted validator set.                                                                                                                |
| `threshold`           | int    | Threshold value of missed blocks in a row for rank updates.                                                                                                |
| `rank`                | int    | Rank on node startup.                                                                                                                                      |
| `validator_laddr`     | string | TCP socket address the validator listens on for an external PrivValidator process. Pairmint dials this address to establish a connection to the validator. |
| `validator_laddr_rpc` | string | TCP socket address the validator's RPC server listens on.                                                                                                  |

### FilePV

FilePV contains configuration parameters for the file-based signer.

| Parameter         | Type   | Description                                                                                                 |
| :---------------- | :----- | :---------------------------------------------------------------------------------------------------------- |
| `chain_id`        | string | Chain ID the validator is part of.                                                                          |
| `key_file_path`   | string | Path to the `priv_validator_key.json` file. Defaults to `$PAIRMINT_CONFIG_DIR/priv_validator_key.json`.     |
| `state_file_path` | string | Path to the `priv_validator_state.json` file. Defaults to `$PAIRMINT_CONFIG_DIR/priv_validator_state.json`. |

## Running

After creating the configuration, start Pairmint using:

```shell
$ pairmint start
```

> :warning: Make sure to start the validator with `rank = 1` first. To be absolutely certain, wait for it to be fully synced and only then start the other nodes.
