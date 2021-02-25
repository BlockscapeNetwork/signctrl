# Upgrade Guide

This is a step-by-step guide on how to perform a software upgrade for your SignCTRL validator set.

> :information_source: This guide assumes you have completed the [Setup Guide](./setup.md) before.

## Download The New Version

Check out the latest version via

```shell
$ git checkout <version_tag>
```

and install it via

```shell
$ make install
```

## Configuration

There is a possibility that the structure of the `config.toml` might contain breaking changes in which case you'll have to safely port over your old configuration to the new format.

If you're using the default configuration directory, back it up via

```shell
$ cp -R /path/to/.signctrl /path/to/.signctrl_old
```

and delete it via

```shell
$ rm -rf /path/to/.signctrl
```

#### Initialize the new config directory

Next, create a new configuration directory by initializing SignCTRL via

```shell
$ signctrl init
```

This will give you

```text
/path/to/.mintctrl/
├── config.toml
└── conn.key
```

> :information_source: The `conn.key` is only used on a per-connection-basis to authenticate SignCTRL to the validator, so it doesn't need to be ported over from the old configuration.

#### Copy the key and state files

Now, that you have your new `.signctrl/` configuration directory, you can copy over the `priv_validator_key.json` and `priv_validator_state.json` files via

```shell
$ cp /path/to/./signctrl_old/*.json /path/to/.signctrl
```

#### Compare the config.toml files

There is currently no automatic way of porting over old configuration settings into new formats, so for now, this process will remain manual. One way to go about it is to open both config.toml files side by side, copy over known fields from the old config.toml and finally fill in the new fields.

> :warning: **IMPORTANT WARNING** :warning:
> 
> **Before you restart** a node, please double-check the node's `rank`.
> It's possible that the rank specified in the `last_rank.json` becomes **obsolete** due to a rank update that happened while the node was stopped. 
> Please, always check the logs to see which rank the validator is currently at. You'll need a minimum log level of INFO for this.

## Rolling Update

> :information_source: Replace all occurrences of `simd` with the name of your binary.

Alright, we're all set up for the upgrade! The beauty of SignCTRL is that it can be upgraded with minimal or no downtime at all (depending on your ability to quickly copy and paste commands into the terminal, so keep the commands below ready in a text editor).

First thing you need to do is to stop the validator and the SignCTRL daemon (in this order) via

```shell
$ sudo systemctl stop simd
$ sudo systemctl stop signctrl
```

and start them back up again (in reverse order) via

```shell
$ sudo systemctl start signctrl
$ sudo systemctl start simd
```

> For convenience, you can also chain all the commands above with small delays in between:
> ```shell
> $ sudo systemctl stop simd; sleep 0.5s; sudo systemctl stop signctrl; sleep 0.5s; sudo systemctl start signctrl; sleep 0.5s; sudo systemctl start simd
> ```