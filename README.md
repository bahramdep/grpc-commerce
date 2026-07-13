# gRPC Commerce

A production-oriented learning project built incrementally in Go. The finished
system will contain Order, Inventory, Payment, and Activity services using gRPC,
Protobuf, PostgreSQL, OpenTelemetry, containers, and Kubernetes.

We are intentionally starting smaller than that description. Each capability is
introduced only when a use case and an acceptance test justify it.

## Current milestone

Milestone 0 established a trustworthy development loop. Milestone 1 designs the
first production-facing contract: `OrderService.CreateOrder`. Read
[`docs/milestones/01-order-contract.md`](docs/milestones/01-order-contract.md).

## Prerequisites

- Go 1.26.5
- Buf 1.71.0
- GNU Make or a compatible `make`
- A POSIX shell for the toolchain check

Code generation uses pinned remote plugins, so a separate local `protoc` or
`protoc-gen-go` installation is not required. Docker and Kubernetes are
introduced in later milestones.

## Commands

```sh
make help
make check-tools
make fmt
make proto-check
make proto-generate
make check
```

## Module path

The permanent module path is `github.com/bahramdep/grpc-commerce`.

## Engineering rules

- Contracts and failure behavior are designed before transport implementation.
- Generated code is never edited by hand.
- Service data ownership is exclusive; no cross-service table access.
- Every goroutine must have an owner, stop condition, and join path.
- Outbound calls have explicit deadlines; retries require safe semantics.
- Decisions with lasting consequences are recorded under `docs/adr/`.
- A milestone is complete only when its acceptance gate passes.
