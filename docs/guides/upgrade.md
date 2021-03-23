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
/path/to/.signctrl/
├── config.toml
└── conn.key
```

> :information_source: The `conn.key` is only used on a per-connection-basis to authenticate SignCTRL to the validator, so it doesn't need to be ported over from the old configuration.

#### Copy the key and state files

Now, that you have your new `.signctrl/` configuration directory, you can copy over the `priv_validator_key.json` and `priv_validator_state.json` files via

```shell
$ cp /path/to/.signctrl_old/priv_validator_*.json /path/to/.signctrl
```

#### Compare the config.toml files

There is currently no automatic way of porting over old configuration settings into new formats, so for now, this process will remain manual. Open both config.toml files side by side, copy over known fields from the old config.toml and finally fill in the new fields.

## Rolling Update

> :information_source: Replace all occurrences of `simd` with the name of your binary.

Alright, we're all set up for the upgrade! The beauty of SignCTRL is that it can be upgraded with minimal or no downtime at all.

Just restart the SignCTRL daemon via

```shell
$ sudo systemctl restart signctrl
```