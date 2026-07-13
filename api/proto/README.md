# Protobuf API source

Versioned gRPC contracts live below this directory. The first contract will be:

```text
commerce/order/v1/order_service.proto
```

Generated Go files do not belong here. `make proto-generate` writes them under
`gen/go`.
