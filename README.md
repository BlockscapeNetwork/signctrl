# Pairmint

Pairmint is a high availability solution for Tendermint-based blockchain validators. It mimics an external PrivValidator process and as such acts as a middleware layer between [Tendermint](https://github.com/tendermint/tendermint) and an external PrivValidator process, making it also compatible with the [TMKMS](https://github.com/iqlusioninc/tmkms). Furthermore, it is designed to get rid of all communication overhead between a set of redundant validator nodes running in parallel by using the blockchain itself as a perfectly synchronous communication line for double-signing protection.

## Prerequisites

* Go `v1.15+`

## Build/Install

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
$ make deps
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

The `init` command creates a `pairmint.toml` configuration file at the directory specified in the `PAIRMINT_CONFIG_DIR` environment variable (defaults to `$HOME/.pairmint`) and a `pm-identity.key` file which holds the seed used to establish a secret connection to Tendermint and an external PrivValidator process.

Please look through the `pairmint.toml` file after it's generated as it is only a template and initially not valid.

The `pairmint.toml` consists of two sections with the following parameters:

### Init

| Parameter   | Type   | Description                                                 |
|:------------|:-------|:------------------------------------------------------------|
| `log_level` | string | Minimum log level for pairmint's log messages.              |
| `set_size`  | int    | Fixed size of the pairminted validator set.                 |
| `threshold` | int    | Threshold value of missed blocks in a row for rank updates. |
| `rank`      | int    | Rank on node startup.                                       |

### Connection

| Parameter              | Type   | Description                                                                                              |
|:-----------------------|:-------|:---------------------------------------------------------------------------------------------------------|
| `validator_addr`       | string | The TCP socket address of the validator node for Pairmint to connect to.                                 |
| `priv_validator_laddr` | string | The TCP socket address for Pairmint to listen on for connections from an external PrivValidator process. |

## Running

After creating the configuration, start Pairmint with the following command:

```shell
$ pairmint start
```

By default, Pairmint provides a file-based signer. If you want to sign with the tmkms, you can pass in the `--tmkms` flag.

> :warning: Make sure to start the validator with `rank = 1` first. To be absolutely certain, wait for it to be fully synced and only then start the other nodes.
