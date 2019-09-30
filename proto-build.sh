#!/usr/bin/env bash
python -m grpc_tools.protoc -I . --python_out=. ./beneath/proto/engine.proto
python -m grpc_tools.protoc -I . --python_out=. --grpc_python_out=. ./beneath/proto/gateway.proto
