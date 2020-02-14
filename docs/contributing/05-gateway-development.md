# Gateway server development

The gateway server code is predominantly found in `gateway/`. 

## Adding and modifying endpoints

We strive to offer both gRPC and REST interfaces for all functionality in the gateway. These are implemented separately to create the most intuitive experience for each protocol. When adding or modifying endpoints, make sure that your changes are reflected in both the `gateway/http/` and `gateway/grpc/` subpackages. 

## Updating protocol buffer definitions

- Run `scripts/proto-build.sh` to (re)generate all protocol buffers classes in the project
