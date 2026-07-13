# Milestone 0: trustworthy development loop

## Why this exists

A service is not production-oriented if a clean checkout cannot be built and
checked consistently. This milestone creates the feedback loop before gRPC,
Protobuf, databases, or containers add more failure modes.

This corresponds to the toolchain and project-setup concerns in Babal chapter 3
and Jean chapter 4. Contract generation begins in Milestone 1.

## Current decisions

- Use one repository while learning, with real service boundaries added later.
- Pin Go 1.26.5, the current stable release when this milestone was created.
- Use the standard Go tools before adding third-party linters.
- Use the permanent module path `github.com/bahramdep/grpc-commerce`.
- Do not add gRPC, Protobuf, Buf, Docker, or Kubernetes dependencies without a
  concrete use case.

## Mentor contribution

- Repository conventions and ignore rules.
- Pinned Go version and toolchain verifier.
- Make targets for formatting, vetting, tests, and race detection.
- Minimal CI workflow.
- ADR and learning-log templates.

## Learner contribution

1. Install Go 1.26.5.
2. Run `make check-tools` and confirm the pinned compiler is active.
3. Review the repository commands and engineering rules.

## Acceptance gate

- `make check-tools` passes with Go 1.26.5.
- The learner can explain what `fmt`, `vet`, `test`, and `test-race` contribute.
- A clean checkout documents the same pinned toolchain and commands.

## Result

Completed on 2026-07-13. The learner verified Go 1.26.5 with
`make check-tools`. Production build metadata is deferred until the deployment
and observability milestone, where a running service will consume it.
