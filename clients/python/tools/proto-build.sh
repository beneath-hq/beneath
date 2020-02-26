#!/usr/bin/env bash

# check we're in the client repo
if [[ ! -f tools/$(basename $0) ]]; then
  echo "you must execute this script from the python client root folder (clients/python)"
  exit 1
fi

# copy canonical grpc proto file into client
PROTO_SRC=../../gateway/grpc/proto/gateway.proto
PROTO_DST=beneath/proto/gateway.proto
cp $PROTO_SRC $PROTO_DST

# generate python bindings
python -m grpc_tools.protoc -I . --python_out=. --grpc_python_out=. ./beneath/proto/gateway.proto
