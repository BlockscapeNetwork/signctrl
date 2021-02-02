# Signing Process

Pairmint itself doesn't sign any messages. It rather functions as a lockgate to the external PrivValidator process, so it's in charge of passing votes/proposals through to the PrivValidator if, and only if, the node is actually allowed to sign according to its current rank.

## Primary Happy Path

TODO

![Signer Happy Path](../imgs/Pairmint-Signer-Happy-Path.png)

![Signer Misses Block](../imgs/Pairmint-Signer-Misses-Block.png)

![Signer Exceeds Threshold](../imgs/Pairmint-Signer-Exceeds-Threshold.png)

![Backup Happy Path](../imgs/Pairmint-Backup-Happy-Path.png)

![Backup Misses Block](../imgs/Pairmint-Backup-Misses-Block.png)

![Backup Exeeds Threshold](../imgs/Pairmint-Backup-Exceeds-Threshold.png)
