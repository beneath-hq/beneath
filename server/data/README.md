# `gateway/`

Despite the name, this module only encapsulates connectivity to the data plane (loading data to/from streams).

## Structure

- `gateway/http/` and `gateway/grpc/` contains the REST and gRPC server definitions
- `gateway/pipeline/` contains the background data processing pipeline (runs as a separate executable)
- `gateway/subscriptions/` contains logic shared between `gateway/http/` and `gateway/grpc/` for handling subscriptions via Websockets and gRPC Unary Subscriptions respectively

## Adding and modifying endpoints

We strive to offer both gRPC and REST interfaces for all functionality in the gateway. These are implemented separately to create the most intuitive experience for each protocol. When adding or modifying endpoints, make sure that your changes are reflected in both the `gateway/http/` and `gateway/grpc/` subpackages. 

## Updating protocol buffer definitions

- Run `scripts/proto-build.sh` to (re)generate all protocol buffers classes in the repository
