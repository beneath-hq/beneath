#!/usr/bin/env bash
protoc -I beneath/proto/ beneath/proto/*.proto --go_out=plugins=grpc:beneath/proto