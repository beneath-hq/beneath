#!/bin/bash -e

# Generate apollo/possibleTypes.ts
yarn run ts-node ./scripts/apolloGeneratePossibleTypes.ts "/graphql" "apollo/possibleTypes.json"

# Generate Typescript types for our Apollo results in apollo/types/
yarn run apollo codegen:generate \
  --config ./apollo.config.js \
  --outputFlat apollo/types \
  --target typescript \
  --passthroughCustomScalars \
  --customScalarsPrefix Control
