# SignCTRL

[![Go Report Card](https://goreportcard.com/badge/github.com/blockscapenetwork/signctrl)](https://goreportcard.com/report/github.com/blockscapenetwork/signctrl)
[![License](https://img.shields.io/badge/License-Apache%202.0-olive.svg)](https://opensource.org/licenses/Apache-2.0)

SignCTRL is a high availability solution for Tendermint that enables the creation of a highly available, self-managing set of validators that uses the blockchain itself as a perfectly synchronous communication line for double-signing protection.

> :warning: Be mindful of key security as SignCTRL currently only supports Tendermint's file-based signer. Make sure to properly secure your validator instance from unauthorized access.

## Why SignCTRL?

1) Built-in double-signing protection.
2) Very lightweight due to not introducing any additional communication overhead for coordination work.
3) No more sentry nodes are needed, as the validators in the set back each other up.
4) A minimal setup requires only two nodes to be run.

## Requirements

* Go `v1.15+`
* Tendermint `v0.34+` (with protobuf support)

## Download

Get the repository via

```shell
$ git clone https://github.com/BlockscapeNetwork/signctrl && cd signctrl
```

## Build & Install

To build the binary into the `./build/` directory, use

```shell
$ make build       # local os/arch
$ make build-linux # linux/amd64
```

Alternatively, install the binary directly to your `$GOPATH/bin` via

```shell
$ make install
```

## Getting Started

To get started, please see the [Guides/Tutorials](docs/guides/README.md).

## Documentation

The documentation can be found [here](docs/README.md)