# Migration Guide

This is a step-by-step guide on how to migrate from an existing node setup to a SignCTRL setup. For this guide, we're going to assume a single-node setup.

> :information_source: Please follow the [Setup Guide](../guides/setup.md) first, until you're ready to run your validator with SignCTRL.

## Rolling Migration

Just like with a regular version upgrade, a migration to SignCTRL can also be done during operation, such that there is no to minimal downtime (depending on your ability to quickly copy and paste commands into the terminal, so keep the commands below ready in a text editor).

First, please follow the [Setup Guide](../guides/setup.md) to prepare the binaries and configure the nodes properly, but **don't run them yet**.

## Migration Order

> :information_source: Replace all occurrences of `simd` with the name of your binary.

The best/safest way to migrate to SignCTRL is to migrate your already running single-node validator first. Make sure you assign `rank = 1` to it in its `config.toml`.

In order for SignCTRL to take effect, first stop your node via

```shell
$ sudo systemctl stop simd
```

then, start SignCTRL via

```shell
$ sudo systemctl start signctrl
```

and finally, start the validator again via

```shell
$ sudo systemctl start simd
```

> For convenience, you can also chain the commands above with small delays in between:
> ```shell
> $ sudo systemctl stop simd; sleep 0.5s; sudo systemctl start signctrl; sleep 0.5s; sudo systemctl start simd
> ```

If you've successfully migrated your single-node validator to SignCTRL, you can proceed starting the rest of the validators in the set (in no particular order) via the same commands.