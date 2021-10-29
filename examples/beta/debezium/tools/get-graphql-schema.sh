#! /usr/bin/env bash

# check we're in the client repo
if [[ ! -f tools/$(basename $0) ]]; then
  echo "you must execute this script from the java client root folder (clients/java)"
  exit 1
fi

./gradlew downloadApolloSchema \
  --endpoint="http:host.docker.internal:4000/graphql" \
  --schema="src/main/graphql/dev/beneath/schema.graphqls"