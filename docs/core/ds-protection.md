# Double-Signing Protection

This page covers how SignCTRL's underlying ranking system provides protection against double-signing.

## SignCTRL Set

A SignCTRL set consists of two or more validators that share the same keypair, and thus represent the same validator entity. Letting all of those validators sign blocks simultaneously is a sure-fire way of getting slashed for double-signing, so there needs to be some form of coordination in terms of which validator in the set has permission to sign at any given point in time. Thus, the idea is very simple - only one validator in the set should sign blocks while the others back it up if it becomes unavailable.

Contrary to approaches that add a consensus layer, and therefore more communication overhead, to the set of redundant validators, like [Raft](https://raft.github.io/), SignCTRL uses the blockchain itself as a perfectly synchronous communication line that the validators use to coordinate signing permissions.

## Validator Ranking

SignCTRL employs a ranking system for the validators in its set that ranks the validators in descending order and it is this ranking system that is the key factor enabling double-signing protection.

### Ranks

A node's rank determines which blocks exactly it has permission to sign and which not. Only the highest-ranked validator signs blocks while the others queue up as backups. The validators can move up one rank at a time if one key criterion is met - and that is if too many blocks have been missed in a row. So, rank updates are triggered by too many blocks on the blockchain being missed in a row.

![Rank Updates](../imgs/rank-update.gif)

In order to detect missed blocks, the validators closely monitor every single block in the blockchain. This includes looking into every last block's commit signatures and checking for their own validator's signature. If the signature is missing, every validator in the set will see it and increment an internal counter. If a certain threshold is exceeded, ranks 2..n will notice first and accordingly move up one rank each. Once rank 1 becomes available again, it will have to sync up its blockchain state. Eventually, while syncing, it will also notice that is has been replaced and needs to shut itself down. It can then later be readded to the set with the lowest rank, though.

### State

Before the node shuts itself down, it persists its last rank and last height in a separate `signctrl_state.json` file. This file acts as a protection mechanism against launching a validator with an rank that has been rendered obsolete by a rank update in the set, which is the case if the requested height differs more than `threshold+1` from the last height persisted in the state file.

For now, the only way to recover from a deprecated state is to delete the `signctrl_state.json` and start the validator back up again with the correct `start_rank` in its `config.toml`.
