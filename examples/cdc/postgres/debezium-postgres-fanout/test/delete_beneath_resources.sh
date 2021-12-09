#!/bin/bash
export BENEATH_ENV=dev
export ORGANIZATION=ericpgreen2
export PROJECT=debezium-postgres-testdb
# delete all tables
beneath project show $ORGANIZATION/$PROJECT | jq '.tables | .[] | .name' | tr -d '"' | while read line; do beneath table delete $ORGANIZATION/$PROJECT/$line ; done
# then delete the project
beneath project delete $ORGANIZATION/$PROJECT