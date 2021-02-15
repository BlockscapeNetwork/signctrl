# SignCTRL

[![Go Report Card](https://goreportcard.com/badge/github.com/blockscapenetwork/signctrl)](https://goreportcard.com/report/github.com/blockscapenetwork/signctrl)
[![License](https://img.shields.io/github/license/cosmos/cosmos-sdk.svg)](https://github.com/cosmos/cosmos-sdk/blob/master/LICENSE)

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

Get the repository either via

```shell
$ git clone https://github.com/BlockscapeNetwork/signctrl && cd signctrl
```

or

```shell
$ go get github.com/BlockscapeLab/signctrl && cd $GOPATH/src/github.com/BlockscapeLab/signctrl
```

## Build & Install

The binary can be built into the `./build` directory via

```shell
$ make build       # local os/arch
$ make build-linux # linux/amd64
```

or installed directly into `$GOPATH/bin` via

```shell
$ make install
```

## Running

Run the SignCTRL daemon via

```shell
$ start signctrl
```

## Getting Started

* [Setting up a validator with SignCTRL](../signctrl/docs/guides/setup.md)
* [Performing a software upgrade](../signctrl/docs/guides/upgrade.md)
* [Migrating from another setup](../signctrl/docs/guides/migration.md)