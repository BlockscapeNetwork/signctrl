# Setup Guide

This is a step-by-step guide on how to set up your Tendermint validator with SignCTRL.

## Node Setup

In order to set up your node, you can follow Tendermint's official documentation.

* [Setting up the keyring](https://docs.cosmos.network/master/run-node/keyring.html)
* [Running a Node](https://docs.cosmos.network/master/run-node/run-node.html)

## SignCTRL Setup

Each validator node runs in tandem with its own SignCTRL daemon, so each and every validator in the set needs to have its own SignCTRL daemon with its own configuration.

### Initialization

SignCTRL needs to be configured in order to be able to talk to a validator. The configuration directory default to `$HOME/.signctrl` - it can be set to a custom directory, though, via the environment variable `$SIGNCTRL_CONFIG_DIR`.

First thing we're going to do is initialize SignCTRL via

```shell
$ signctrl init
```

This will create the following files in your configuration directory:

```text
$HOME/.mintctrl/
├── config.toml
└── conn.key
```

The `config.toml` is the configuration file for SignCTRL. We will be taking a look at it in great detail in the **Configuration** section.

The `conn.key` file is a secret key that is used to establish an encrypted connection between SignCTRL and the validator.

The last thing we need to do is import the validator node's `priv_validator_key.json` and `priv_validator_state.json` into the configuration directory. Your directory should now look like this:

```text
$HOME/.mintctrl/
├── config.toml
├── conn.key
├── priv_validator_key.json
└── priv_validator_state.json
```

> :information_source: If you don't already have a key and state file, or want to use new ones, you can use `signctrl init --keypair`.

### Configuration

In the previous section, we've created a `config.toml` file in our configuration directory.

```toml
[init]

# Minimum log level for SignCTRL logs.
# Must be either DEBUG, INFO, WARN or ERR.
log_level = ""

# Number of SignCTRL validator nodes running in parallel.
# This number cannot be changed during operation anymore.
# If you want to change this value you need to stop all
# nodes, modify the configuration on each node and start
# them up again.
# Must be 2 or higher.
set_size = 0

# Number of missed blocks in a row that triggers a rank
# update in the set.
# Must be 1 or higher.
threshold = 0

# SignCTRL node's rank on startup. It is used to determine
# which node in the set is currently signing (rank 1) and
# which nodes line up as backups (ranks 2-n).
# Must be 1 or higher.
rank = 0

# TCP socket address the validator listens on for an external
# PrivValidator process. SignCTRL dials this address to
# establish a connection to the validator.
# Must be in host:port format.
validator_laddr = ""

# TCP socket address the validator's RPC server listens on.
# Must be in host:port format.
validator_laddr_rpc = ""

[file_pv]

# The chain ID the signer should sign for.
chain_id = ""

# The path to the priv_validator_key.json file.
# Defaults to $SIGNCTRL_CONFIG_DIR/priv_validator_key.json if left empty.
key_file_path = ""

# The path to the priv_validator_state.json file.
# Defaults to $SIGNCTRL_CONFIG_DIR/priv_validator_state.json if left empty.
state_file_path = ""
```

The initial `config.toml` is merely a template and needs customization. There are a couple of things to consider:

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
validator_laddr = "127.0.0.1:3000"
validator_laddr_rpc = "127.0.0.1:26657"

[privval]

chain_id = "testchain" # Shared value
key_file_path = "/Users/myuser/.signctrl/priv_validator_key.json"
state_file_path = "/Users/myuser/.signctrl/priv_validator_state.json"
```

</td>
<td>

```toml
[init]

log_level = "INFO"
set_size = 2 # Shared value
threshold = 5 # Shared value
rank = 2 # Unique
validator_laddr = "127.0.0.1:3000"
validator_laddr_rpc = "127.0.0.1:26657"

[privval]

chain_id = "testchain" # Shared value
key_file_path = "/Users/myuser/.signctrl/priv_validator_key.json"
state_file_path = "/Users/myuser/.signctrl/priv_validator_state.json"
```

</td>
</tr>
</table>

### Running

Once all the nodes are configured, start up SignCTRL on each validator node (in no particular order) via

```shell
$ signctrl start
```

Then, start your validator via

```shell
$ simd start
```

> :information_source: Replace `simd` with your binary.