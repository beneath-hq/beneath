#!/usr/bin/env bash
protoc -I beneath/proto/ beneath/proto/gateway.proto --go_out=plugins=grpc:beneath/proto