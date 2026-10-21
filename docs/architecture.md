# Architecture

go-relay accepts a source-chain lock, checks that it is final, and releases the amount from an in-process escrow. Nothing listens on a socket. Guardians attest a batch root. The relay proves a single lock against that root before it pays.

```
            release
  caller ------------> Relay ---- ready? ----> Inbox
                         |                       |
                         |                       +--> Merkle proof
                         |
                         +--> Cursor
                         |
                         +--> Quorum / Payload
                         |
                         +--> Nonce book
                         |
                         +--> Ledger
                         |
                         +--> Log
```

## Relay

`relay.Relay` is the entry point. `Release` rejects a paused relay, checks the lock against `config.Config`, reads the source head, prices the fee, requires a recorded proof, takes a nonce, and pays the recipient from escrow.

The config carries the source chain, the destination chain, the confirmation depth, the operator, the guardian threshold, the starting fee, the fee recipient, and the maximum amount. A lock for a different chain is rejected.

## Cursor

`header.Cursor` stores the latest source header. `Advance` accepts the next header only when its parent hash is the previous header's hash. A lock is final when the source head is at least `confirmations` blocks ahead of the lock block.

## Access

`access.Roles` records the operator for this relay. Registering a guardian, advancing the head, funding escrow, committing a batch, and changing the fee or the pause flag are operator calls.

## Fee

`fee` takes a basis-point cut of the amount, capped at 1% (100 bps). The cut is attributed to the fee recipient held by the relay. The rest is credited to the lock's recipient. Both amounts come out of escrow.

## Guardians

Each guardian is registered with a 32-byte key. `CommitBatch` builds the bytes with `payload.EncodeRoot` and checks tags with `quorum`. A batch root is committed when at least `threshold` guardian tags verify. `mac` computes HMAC-SHA256. The root payload is prefixed with `go-relay/v1/batch` and includes both chain ids.

## Payload

`payload.Encode` is the lock leaf preimage. It is prefixed with `go-relay/v1/lock` and includes both chain ids, the sender, the recipient, the asset, the amount, the nonce, and the lock block. `payload.Leaf` hashes that preimage with a leaf domain byte. The fee is relay policy and is not part of the leaf.

## Proof

`merkle` hashes a leaf with its siblings. Internal nodes use a different domain byte than leaves. `Prove` checks that path against a committed root and records the leaf. `Release` pays a leaf only after that proof is recorded.

## Nonce

`nonce.Book` keeps the next nonce for a sender. `Commit` accepts only that nonce and then advances it. The nonce is part of the leaf. Two senders do not share a sequence. `Allocate` is the same cursor, used on its own.

## Ledger

`vault.Ledger` holds the escrow pool and per-account balances. `Fund` adds to the pool. A release debits the pool for the gross amount and credits the recipient and the fee recipient.

## Helpers

`format.Group` groups an integer amount for a log line. `lag.Window` keeps a rolling mean of delay samples. `hex.Decode` reads a hex key, including an optional `0x` prefix.

## Broadcast

`broadcast.Log` appends each release receipt to a log the caller can read.
