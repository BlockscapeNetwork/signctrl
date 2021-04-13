# Frequently Asked Questions

### How can I use my existing keypair with SignCTRL?

Just copy and paste your `priv_validator_key.json` and `priv_validator_state.json` into your SignCTRL configuration directory.

### What should I check before I start my validators?

Before starting any validator in the set, **always** make sure no two validators are assigned to the same `start_rank`.

### My SignCTRL node has shut down and my validator is listening for a new connection. What do I do?

If your SignCTRL node has shut down, it is paramount that the **validator daemon is restarted before SignCTRL is restarted**.

> :warning: Do NOT leave the validator daemon running and start SignCTRL! This will lead to double-signing!

This is merely a workaround for now. The issue is further investigated in [this issue](https://github.com/BlockscapeNetwork/signctrl/issues/24).

### Which order should I start my validators in?

It doesn't matter which order you start your validators in. Starting ranks `2..n` prior to rank `1` is just as safe to do as vice-versa because ranks `2..n` will always wait for rank `1` to sign at least one block before they start counting blocks missed in a row.

### How do I update the SignCTRL binary?

Follow the [Upgrade Guide](../guides/upgrade.md).

### How do I update my validator's binary?

1) Stop the validator daemon.
2) Wait for SignCTRL to try redialing the validator (`retry_dial_after` in the `config.toml`).
3) Start the validator daemon.

### How do I migrate from my existing setup to SignCTRL?

Follow the [Migration Guide](../guides/migrate.md).

### How can I add/remove validators to/from the SignCTRL set?

At this point in time, it's not possible to add or remove validator's to/from the set on the fly.

### SignCTRL immediately shuts itself down when I try to start it.

This is a protection mechanism rooted in the `signctrl_state.json` file. It protects against launching a validator with an rank that has been rendered obsolete by a rank update in the set, which is the case if the requested height differs more than `threshold+1` from the last height persisted in the state file. In order to fix this, please follow the steps below.

1) Check each validator's rank via `signctrl status`, i.e. validator 1 is ranked 1st and validator 2 us ranked 3rd, which means that rank 2 is free.
2) Update the validator's `start_rank` in the `config.toml` to the free rank.
3) Delete the `signctrl_state.json` file.
4) Start SignCTRL.
