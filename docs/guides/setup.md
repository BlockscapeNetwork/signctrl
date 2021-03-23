# Setup Guide

This is a step-by-step guide on how to set up your Tendermint validator with SignCTRL.

> :information_source: Please complete the **Build & Install** section in the [README](../../README.md) first.

## Node Setup

In order to set up your node, you can follow Tendermint's official documentation.

* [Setting up the keyring](https://docs.cosmos.network/master/run-node/keyring.html)
* [Running a Node](https://docs.cosmos.network/master/run-node/run-node.html)

## SignCTRL Setup

Each validator node runs in tandem with its own SignCTRL daemon, so each and every validator in the set needs to have its own configuration.

### Initialization

SignCTRL needs to be configured in order to be able to talk to a validator. The configuration directory defaults to `$HOME/.signctrl` - it can be set to a custom directory, though, via the environment variable `$SIGNCTRL_CONFIG_DIR`.

First thing we're going to do is initialize SignCTRL via

```shell
$ signctrl init
```

This will create the following files in your configuration directory:

```text
$HOME/.signctrl/
├── config.toml
└── conn.key
```

The `config.toml` is the configuration file for SignCTRL. We will be taking a look at it in great detail in the **Configuration** section.

The `conn.key` file is a secret key that is used to establish an encrypted connection between SignCTRL and the validator.

The last thing we need to do is import the validator node's `priv_validator_key.json` and `priv_validator_state.json` into the configuration directory. Your directory should now look like this:

```text
$HOME/.signctrl/
├── config.toml
├── conn.key
├── priv_validator_key.json
└── priv_validator_state.json
```

> :information_source: If you don't already have a `priv_validator_key.json` and `priv_validator_state.json`, or want to use new ones, you can use `signctrl init --new-pv`.

### Configuration

In the previous section, we've created a `config.toml` file in our configuration directory.

```toml
#################################################
### Init defines the configuration parameters ###
### that SignCTRL needs on initialization.    ###
#################################################

[init]

# Minimum log level for SignCTRL logs.
# Must be either DEBUG, INFO, WARN or ERR.
log_level = "INFO"

# Number of validators in the SignCTRL set.
# This value must be the same across all validators
# in the set.
# Must be 2 or higher.
set_size = 2

# Number of missed blocks in a row that triggers
# a rank update in the set.
# This value must be the same across all validators
# in the set.
# Must be 1 or higher.
threshold = 10

# Rank of the validator on startup.
# Rank 1 signs, while ranks 2..n serve as backups
# until the threshold is exceeded and ranks are
# updated.
# Must be 1 or higher.
rank = 0

# TCP socket address the validator listens on for
# an external PrivValidator process.
# Must be a TCP address in the host:port format.
validator_laddr = "tcp://127.0.0.1:3000"

# TCP socket address the validator's RPC server
# listens on.
# Must be a TCP address in the host:port format.
validator_laddr_rpc = "tcp://127.0.0.1:26657"

# Time after which SignCTRL assumes it lost the
# connection to the validator and retries dialing
# it.
# Must be 1 or higher. Use 's' for seconds, 'm' for
# minutes and 'h' for hours.
retry_dial_after = "15s"

####################################################
### Privval defines the types of private         ###
### validators that sign incoming sign requests. ###
####################################################

[privval]

# The chain the validator validates for.
chain_id = ""
```

The initial `config.toml` provides a set of default values for most fields. Please make sure to customize the fields `rank` and `chain_id` to your individual needs after generation.

Furthermore, there are a couple of things to consider:

* `set_size`, `threshold` and `chain_id` must be shared values across all validators in the set
* `rank` must be unique, so no two validators in the set can have the same rank

#### Example Configuration

The following two configurations are an example for a SignCTRL set of two validators.

<table>
<tr>
<th>Validator #1</th>
<th>Validator #2</th>
</tr>
<tr>
<td>

```toml
[init]

log_level = "INFO"
set_size = 2 # Shared value
threshold = 5 # Shared value
rank = 1 # Unique
validator_laddr = "tcp://127.0.0.1:3000"
validator_laddr_rpc = "tcp://127.0.0.1:26657"
retry_dial_after = "15s"

[privval]

chain_id = "testchain" # Shared value
```

</td>
<td>

```toml
[init]

log_level = "INFO"
set_size = 2 # Shared value
threshold = 5 # Shared value
rank = 2 # Unique
validator_laddr = "tcp://127.0.0.1:3000"
validator_laddr_rpc = "tcp://127.0.0.1:26657"
retry_dial_after = "15s"

[privval]

chain_id = "testchain" # Shared value
```

</td>
</tr>
</table>

### Unit File

It is recommended to use `systemctl` to run SignCTRL. Here's an example of a `signctrl.service` unit file:

```text
[Unit]
Description=signctrl
Requires=network-online.target
After=network-online.target

[Service]
Restart=on-failure
User=signer
Group=signer
PermissionsStartOnly=true
ExecStart=/home/signer/go/bin/signctrl start
KillSignal=SIGTERM
LimitNOFILE=4096
Environment=SIGNCTRL_CONFIG_DIR=/Users/signer/.signctrl

[Install]
WantedBy=multi-user.target
```

### Running

> :information_source: Replace all occurrences of `simd` with the name of your binary.

Once all the nodes are configured, start the SignCTRL service first via

```shell
$ sudo systemctl start signctrl
```

and then, start your validator via

```shell
$ sudo systemctl simd start
```

> For convenience, you can also chain the commands above with small delays in between:
> ```shell
> $ sudo systemctl start signctrl; sleep 0.5s; sudo systemctl start simd
> ```

> :information_source: It doesn't matter which order you start your validators in. Ranks 2..n will always wait for a signature from rank 1 before they start counting missed blocks in a row and consequently updating their ranks.