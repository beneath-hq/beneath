#!/usr/bin/env bash
protoc -I proto/ proto/*.proto --go_out=plugins=grpc:proto
protoc -I engine/driver/bigtable/proto/ engine/driver/bigtable/proto/*.proto --go_out=plugins=grpc:engine/driver/bigtable/proto
