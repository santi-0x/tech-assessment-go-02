# Go Relay

An in-process bridge relay. A caller submits a source lock; the relay checks finality, checks a guardian batch, prices a fee, checks the nonce, and releases the amount from escrow.

- `relay.Relay` is the entry point.
- `nonce.Book` keeps a separate nonce sequence for each sender.
- `quorum` checks HMAC-SHA256 tags over `payload.EncodeRoot`.
- `access.Roles` records the operator.
- `broadcast.Log` keeps an in-process log. Nothing listens on a socket.

See `docs/architecture.md`.

## Check

Requires Go 1.22 or newer. The module has no dependencies.

```sh
sh script/run-check.sh
```
