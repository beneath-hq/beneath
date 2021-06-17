# `server/data/`

This package implements the data-plane server, which handles loading data to/from tables.

The `http/` and `grpc/` contains the REST and gRPC server handlers respectively. For live subscriptions, they respectively offer websockets and gRPC unary streaming interfaces. They both largely delegate to `services/data.Service`, which contains the non-protocol specific handler implementations.

## Adding and modifying endpoints

We strive to offer both gRPC and REST interfaces for all data-plane functionality. These are implemented separately to create the most intuitive experience for each protocol. When adding or modifying endpoints, make sure that your changes are reflected in both the `http/` and `grpc/` subpackages.

## Updating protocol buffer definitions

- Run `scripts/proto-build.sh` to (re)generate all protocol buffers classes in the repository
