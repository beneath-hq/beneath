#!/usr/bin/env bash

shopt -s expand_aliases
alias compile="protoc --go_out=plugins=grpc,paths=source_relative:."

compile infrastructure/engine/proto/*.proto
compile infrastructure/engine/driver/bigtable/proto/*.proto 
compile infrastructure/engine/driver/bigquery/proto/*.proto 
compile server/data/grpc/proto/*.proto
