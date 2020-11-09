#!/usr/bin/env bash

shopt -s expand_aliases
alias compile="protoc --go_out=plugins=grpc,paths=source_relative:."

compile infra/engine/proto/*.proto
compile infra/engine/driver/bigtable/proto/*.proto 
compile infra/engine/driver/bigquery/proto/*.proto 
compile server/data/grpc/proto/*.proto
