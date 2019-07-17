#!/usr/bin/env bash
protoc -I proto/ proto/*.proto --go_out=plugins=grpc:proto