#!/usr/bin/env bash

shopt -s expand_aliases
alias compile="protoc --go_out=plugins=grpc,paths=source_relative:."

compile engine/proto/*.proto
compile engine/driver/bigtable/proto/*.proto 
compile engine/driver/bigquery/proto/*.proto 
compile gateway/grpc/proto/*.proto
