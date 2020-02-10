#!/usr/bin/env bash

shopt -s expand_aliases
alias compile="protoc --go_out=plugins=grpc,paths=source_relative:."

compile proto/*.proto
compile engine/proto/*.proto
compile engine/driver/bigtable/proto/*.proto 

