#!/usr/bin/env bash
python -m grpc_tools.protoc -I. --python_out=. engine.proto
python -m grpc_tools.protoc -I. --python_out=. --grpc_python_out=. gateway.proto

# regarding VS Code throwing an error in the X_pb2_grpc.py file:
# https://stackoverflow.com/questions/53634047/class-imagefile-has-no-fromstring-member-when-compiling-a-proto-file
