# SignCTRL
![CI](https://github.com/BlockscapeNetwork/signctrl/actions/workflows/build_test.yml/badge.svg)
[![Codecov](https://codecov.io/gh/BlockscapeNetwork/signctrl/branch/master/graph/badge.svg)](https://codecov.io/gh/BlockscapeNetwork/signctrl)
[![Go Report Card](https://goreportcard.com/badge/github.com/blockscapenetwork/signctrl)](https://goreportcard.com/report/github.com/blockscapenetwork/signctrl)
[![License](https://img.shields.io/badge/License-Apache%202.0-olive.svg)](https://opensource.org/licenses/Apache-2.0)

SignCTRL is a high availability solution for Tendermint that enables the creation of a highly available, self-managing set of validators that uses the blockchain itself as a perfectly synchronous communication line for double-signing protection.

> :warning: SignCTRL is still beta-software and not considered production-ready software. Use at your own risk.

## Why SignCTRL?

1) Built-in double-signing protection
2) Very lightweight (no additional communication overhead for coordination work)
3) No more sentry nodes are needed, as the validators in the set back each other up
4) Minimal setup requires only two nodes

## Requirements

* Go `v1.16+`
* Tendermint `v0.34+` (with protobuf support)

## Download

Get the repository via

```shell
$ git clone https://github.com/BlockscapeNetwork/signctrl
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

To get started, please see the [Guides/Tutorials](docs/guides/README.md).</br>
If you get stuck, see the [FAQ](docs/core/faq.md).

## Documentation

The documentation can be found [here](docs/README.md).

## Security

Security and management of any key material is outside the scope of this service. Always consider your own security and risk profile when dealing with sensitive keys, services, or infrastructure.

## No Liability

As far as the law allows, this software comes as is, without any warranty or condition, and no contributor will be liable to anyone for any damages related to this software or this license, under any kind of legal claim.