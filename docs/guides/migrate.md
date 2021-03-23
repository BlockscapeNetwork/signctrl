# Migration Guide

This is a step-by-step guide on how to migrate from an existing node setup to a SignCTRL setup. For this guide, we're going to assume a single-node setup with no external signing service.

> :information_source: Please follow the [Setup Guide](../guides/setup.md) first, until you're ready to run your validator with SignCTRL.

## Rolling Migration

Just like with a regular version upgrade, a migration to SignCTRL can also be done during operation, such that there is no to minimal downtime (depending on your ability to quickly copy and paste commands into the terminal, so keep the commands below ready in a text editor).

First, please follow the [Setup Guide](../guides/setup.md) to prepare the binaries and configure the nodes properly, but **don't run them yet**.

## Migration Order

> :information_source: Replace all occurrences of `simd` with the name of your binary.

Make sure you assigned `start_rank = 1` to your validator in its `config.toml`.

In order for SignCTRL to take effect, first stop your already running single-node validator via

```shell
$ sudo systemctl stop simd
```

then, start SignCTRL via

```shell
$ sudo systemctl start signctrl
```

SignCTRL now waits for the validator to start up so it can dial it to establish an encrypted connection.

Finally, start the validator again via

```shell
$ sudo systemctl start simd
```

> For convenience, you can also chain the commands above with small delays in between:
> ```shell
> $ sudo systemctl stop simd; sleep 0.5s; sudo systemctl start signctrl; sleep 0.5s; sudo systemctl start simd
> ```

If you've successfully migrated your single-node validator to SignCTRL, you can proceed starting the rest of the validators in the set (in no particular order) in the same fashion.