# Ranking System

Pairmint's ranking system is the key factor enabling double-signing protection amongst its redundant validator nodes. With one signer and `1..n` backup nodes, Pairmint requires at least `n = 2` nodes to be run in parallel with their respective rank `n`.

The rank of a node determines when exactly it is allowed to sign blocks and when not to. The backup nodes (ranks `n >= 2`) constantly monitor the blockchain for consequtive missed blocks. Once a threshold `m` of missed blocks in a row is reached, the validators update their ranks in order to have the second highest ranked validator become the new signer while the old signer who failed to fulfil its duties queue up on the last rank again.

## Example

Suppose we have a set of `n = 3` validator nodes ranked from 1 to 3, rank 1 being the signer and ranks 2 and above being backups.

| Rank | Node |
|:----:|:----:|
| 1    | A    |
| 2    | B    |
| 3    | C    |

Now, let's set the threshold of missed blocks in a row to `m = 5`. This means that once the nodes notice five blocks in a row missing their own signature a rank update is initiated.

| Rank | Node | Ruleset                     |
|:----:|:----:|:----------------------------|
| 1    | A    | if `m == 5` set rank to `n` |
| 2    | B    | if `m == 5` set rank + 1    |
| 3    | C    | if `m == 5` set rank + 1    |

If A now misses `m = 5` blocks in a row, B and C each move up one rank and A falls back to the last rank `n-1`. From this point forward, B is signing all blocks until it misses `m = 5` blocks in a row and the cycle repeats.

| Rank | Node     | Ruleset                     |
|:----:|:--------:|:----------------------------|
| 1    | B (↑ +1) | if `m == 5` set rank to `n` |
| 2    | C (↑ +1) | if `m == 5` set rank + 1    |
| 3    | A (↓ -2) | if `m == 5` set rank + 1    |

## Edge Cases

### Node getting out of sync during operation

- Nodes don't count consecutive missed blocks on startup unless they are fully synced. It is only after they are fully synced that they start counting.
- If a node falls behind during operation, it will have to count missed blocks in a row while it's catching up.

### Two or more nodes failing back to back

- As long as there is at least one backup node online, it's no problem. Only a few more blocks will be missed.
