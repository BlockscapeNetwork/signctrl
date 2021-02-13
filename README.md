# SignCTRL

[![Go Report Card](https://goreportcard.com/badge/github.com/blockscapenetwork/signctrl)](https://goreportcard.com/report/github.com/blockscapenetwork/signctrl)
[![License](https://img.shields.io/github/license/cosmos/cosmos-sdk.svg)](https://github.com/cosmos/cosmos-sdk/blob/master/LICENSE)

> :warning: Be mindful of key security as SignCTRL currently only supports Tendermint's file-based signer. Make sure to properly secure your validator instance from unauthorized access.

SignCTRL is a high availability solution for Tendermint-based blockchain validators. It uses the blockchain itself as a perfectly synchronous communication line between redundant validators running in parallel for double-signing protection.

## Requirements

* Go `v1.15+`
* Validator software compatible with Tendermint `v0.34+` in terms of protobuf support

## Build & Install

Get the repository either via

```shell
$ git clone https://github.com/BlockscapeNetwork/signctrl && cd signctrl
```

or

```shell
$ go get github.com/BlockscapeLab/signctrl && cd $GOPATH/src/github.com/BlockscapeLab/signctrl
```

SignCTRL can be built into the `./build` directory using

```shell
$ make build       # local os/arch
$ make build-linux # linux/amd64
```

or installed directly into `$GOPATH/bin` using

```shell
$ make install
```

## Configuration

Each validator node runs in tandem with its own SignCTRL daemon, and thus each one also has its own configuration. You can initialize SignCTRL using:

```shell
$ sc init
```

The `init` command creates a `config.toml` configuration file at the directory specified in the `$SIGNCTRL_CONFIG_DIR` environment variable (defaults to `$HOME/.signctrl`) and a `pm-identity.key` file which holds the seed used to establish a secret connection to the validator.

> :information_source: Please look through the `config.toml` file after it's generated as it is only a template and initially not valid.

If you don't already have a keypair, you can use the `--keypair` flag to generate a new `priv_validator_key.json` and `priv_validator_state.json` in your `$SIGNCTRL_CONFIG_DIR` directory. SignCTRL will know where to find them without you having to specify their location (see `key_file_path` and `state_file_path` parameters in the FilePV section).

If you do already have a keypair, you can either copy them into the `$SIGNCTRL_CONFIG_DIR` directory or leave the key and state files where they are and specify the paths to them in the FilePV section of the `config.toml`.

### Init

Init contains configuration parameters needed on initialization.

> :information_source: The configuration parameters `set_size` and `threshold` must be the same across all validator nodes in the SignCTRL set.

| Parameter             | Type   | Description                                                                                                                                                                             |
| :-------------------- | :----- | :-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `log_level`           | string | Minimum log level for SignCTRL's log messages. Must either DEBUG, INFO, WARN or ERR..                                                                                                   |
| `set_size`            | int    | Fixed size of the SignCTRL validator set. Must be 2 or higher.                                                                                                                          |
| `threshold`           | int    | Threshold value of missed blocks in a row for rank updates. Must be 1 or higher.                                                                                                        |
| `rank`                | int    | Rank on node startup. Must be 1 or higher.                                                                                                                                              |
| `validator_laddr`     | string | TCP socket address the validator listens on for an external PrivValidator process. SignCTRL dials this address to establish a connection to the validator. Must be in host:port format. |
| `validator_laddr_rpc` | string | TCP socket address the validator's RPC server listens on. Must be in host:port format.                                                                                                  |

### FilePV

FilePV contains configuration parameters for the file-based signer.

| Parameter         | Type   | Description                                                                                                               |
| :---------------- | :----- | :------------------------------------------------------------------------------------------------------------------------ |
| `chain_id`        | string | Chain ID the validator is part of.                                                                                        |
| `key_file_path`   | string | Path to the `priv_validator_key.json` file. Defaults to `$SIGNCTRL_CONFIG_DIR/priv_validator_key.json` if left empty.     |
| `state_file_path` | string | Path to the `priv_validator_state.json` file. Defaults to `$SIGNCTRL_CONFIG_DIR/priv_validator_state.json` if left empty. |

#### Example

Let's assume, you have two validators in your SignCTRL set. The following configurations are examples of their respective `config.toml` files.
An example of a valid `config.toml` could look like this:

##### Validator #1
```toml
[init]

log_level = "INFO"
set_size = 2   # Must be 2 for both validators!
threshold = 10 # Must be 10 for both validators!
rank = 1       # Must be unique! (No two validators can have the same rank)
validator_laddr = "127.0.0.1:3000"
validator_laddr_rpc = "127.0.0.1:26657"

[file_pv]

chain_id = "mychain"
key_file_path = "/Users/myuser/.signctrl/priv_validator_key.json"
state_file_path = "/Users/myuser/.signctrl/priv_validator_state.json"
```

##### Validator #2
```toml
[init]

log_level = "INFO"
set_size = 2   # Must be 2 for both validators!
threshold = 10 # Must be 10 for both validators!
rank = 2       # Must be unique! (No two validators can have the same rank)
validator_laddr = "127.0.0.1:3000"
validator_laddr_rpc = "127.0.0.1:26657"

[file_pv]

chain_id = "mychain"
key_file_path = ""   # Left empty, so it defaults to $SIGNCTRL_CONFIG_DIR/priv_validator_key.json
state_file_path = "" # Left empty, so it defaults to $SIGNCTRL_CONFIG_DIR/priv_validator_state.json
```

## Running

After creating the configuration, start SignCTRL using:

```shell
$ sc start
```
