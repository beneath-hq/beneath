#!/usr/bin/env bash
protoc -I beneath/beneath_proto/ beneath/beneath_proto/*.proto --go_out=plugins=grpc:beneath/beneath_proto