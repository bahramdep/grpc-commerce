# Milestone 1: OrderService v1 contract

## Outcome

Design, lint, and generate the first real service boundary:

```text
OrderService.CreateOrder(CreateOrderRequest) -> CreateOrderResponse
```

This milestone is contract-only. The next slice implements the server with an
in-memory repository. Separating the steps forces us to review compatibility,
ownership, and failure semantics before generated types influence the design.

## Book mapping

- Babal chapter 3: Protobuf, stub generation, project structure, compatibility.
- Babal chapter 4: the Order service and gRPC adapter boundary.
- Jean chapters 2, 4, and 6: Protobuf encoding, project setup, and effective API
  design.

## Contract location

Create exactly one handwritten contract file:

```text
api/proto/commerce/order/v1/order_service.proto
```

Generated Go files belong under:

```text
gen/go/commerce/order/v1/
```

Never edit generated files by hand.

## Domain boundary

The initial command accepts:

- an idempotency key chosen by the caller;
- a customer identifier;
- one or more order items;
- each item contains a product identifier and positive quantity.

The caller does not choose:

- order ID;
- order status;
- creation timestamp;
- inventory reservation result;
- payment result.

Those values are owned by services, not trusted client input.

The response returns the created order snapshot with:

- server-generated order ID;
- customer ID and items;
- status;
- creation timestamp.

For this milestone the only valid created status is `PENDING`. Inventory and
payment transitions are added after their service boundaries exist.

## Protobuf design requirements

- Use `proto3` syntax.
- Package: `commerce.order.v1`.
- Go package:
  `github.com/bahramdep/grpc-commerce/gen/go/commerce/order/v1;orderv1`.
- Use dedicated request and response messages.
- Use `google.protobuf.Timestamp` for the server-owned creation time.
- The status enum zero value ends in `_UNSPECIFIED`.
- Give every field a stable, unique number; do not plan to reuse numbers.
- Add concise comments to the service, RPC, messages, and fields.
- Do not add database fields, transport metadata, inventory fields, payment
  fields, or speculative pagination.

## Questions to answer before writing the file

1. Why is the idempotency key a request field rather than part of `Order`?
2. Why must `status` and `created_at` be absent from the request?
3. Why does `OrderItem` use `product_id` rather than embedding a Product?
4. Should validation rules live in Protobuf today, or initially at the server
   boundary? What dependency would schema validation introduce?
5. What can we add later without breaking old clients? What changes would be
   breaking?

## Learner task

1. Write answers to the five questions in `docs/learning-log.md`.
2. Draft `order_service.proto` from the requirements; do not copy a finished
   schema from a tutorial.
3. Run `make proto-format` and `make proto-check`.
4. Generate code with `make proto-generate`.
5. Add the generated runtime dependencies with pinned versions:

   ```sh
   go get google.golang.org/protobuf@v1.36.11
   go get google.golang.org/grpc@v1.82.0
   go mod tidy
   ```

6. Run `make check`.

## Acceptance gate

- Buf format and lint pass.
- Generated files are reproducible and committed separately from handwritten
  contract changes.
- `go test ./...` compiles generated packages.
- Request fields express caller intent only; response fields express
  server-owned results.
- Package and `go_package` paths are versioned and consistent.
- The learner can identify at least three compatible and three breaking schema
  changes.

