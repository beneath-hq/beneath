#!/bin/bash -e

# Generate ee/apollo/possibleTypes.ts
yarn run ts-node ./scripts/apolloGeneratePossibleTypes.ts "/ee/graphql" "ee/apollo/possibleTypes.json"

# Generate Typescript types for our Apollo results in apollo/types/
yarn run apollo codegen:generate \
  --config ./ee/apollo.config.js \
  --outputFlat ee/apollo/types \
  --target typescript \
  --passthroughCustomScalars \
  --customScalarsPrefix Control
