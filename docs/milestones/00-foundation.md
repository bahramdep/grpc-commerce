# Milestone 0: trustworthy development loop

## Why this exists

A service is not production-oriented if a clean checkout cannot be built and
checked consistently. This milestone creates the feedback loop before gRPC,
Protobuf, databases, or containers add more failure modes.

This corresponds to the toolchain and project-setup concerns in Babal chapter 3
and Jean chapter 4. We deliberately stop before contract generation; that is
Milestone 1.

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

1. Install Go 1.26.5 and run `make check-tools`.
2. Create `internal/platform/buildinfo/buildinfo.go` with:
   - a `Version` value type containing `Commit`, `BuildTime`, and `Dirty`;
   - a constructor that rejects an empty commit and malformed RFC3339 build
     time;
   - no logging, environment reads, global mutation, or panic.
3. Create table-driven tests covering valid input, empty commit, invalid time,
   and the zero value.
4. Add a fuzz test for the constructor. Its invariant: arbitrary input must
   never panic; success must always contain a non-empty commit and valid RFC3339
   timestamp.
5. Run `make fmt` and `make check`.
6. Add the first entry to `docs/learning-log.md` explaining:
   - why `Version` is a value rather than a global variable;
   - why parsing belongs in the constructor;
   - what the race detector can and cannot prove here.

## Constraints for the exercise

- Use only the Go standard library.
- Prefer a concrete type; do not introduce an interface.
- Return errors; do not panic.
- Tests must not depend on wall-clock time or environment variables.
- Do not add a CLI, HTTP server, gRPC server, dependency-injection framework,
  or configuration library.

## Acceptance gate

- `make check-tools` passes with Go 1.26.5.
- `make fmt-check`, `make vet`, `make test`, and `make test-race` pass.
- `go test -fuzz=Fuzz -fuzztime=10s ./internal/platform/buildinfo` finds no
  panic or invariant violation.
- The package API is small and documented.
- The learning-log entry explains decisions rather than restating code.
- A clean checkout can reproduce the same results using only documented
  commands.

## Review questions

Be prepared to answer these in the code review:

1. Should `BuildTime` be stored as `string` or `time.Time`? What boundary does
   each choice create?
2. Should a dirty working tree be an error? Why or why not?
3. Why is an interface unnecessary for this package today?
4. What bug classes can `go test -race` detect, and what does a passing run not
   prove?
5. Which constructor guarantees should callers be able to rely on forever?
