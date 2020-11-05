// This script generates the JSON file apollo/fragmentTypes.json, which contains info
// about fragments in our GraphQL schema for Apollo. See this page for details:
// https://www.apollographql.com/docs/react/data/fragments/#fragments-on-unions-and-interfaces

// Used in generateApolloTypes.sh
// Usage: ts-node apolloGeneratePossibleTypes.ts ENDPOINT_PATH OUT_FILE

import { API_URL} from "../lib/connection";

import fetch from "node-fetch";
import fs from "fs";

if (process.argv.length !== 4) {
  console.log(`Usage: ts-node apolloGeneratePossibleTypes.ts ENDPOINT_PATH OUT_FILE`);
  process.exit();
}

const ENDPOINT = process.argv[2];
const OUT_FILE = process.argv[3];

console.log(`Reading possible types from endpoint "${API_URL}${ENDPOINT}" to file at "${OUT_FILE}"`);

fetch(`${API_URL}${ENDPOINT}`, {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({
    variables: {},
    query: `
      {
        __schema {
          types {
            kind
            name
            possibleTypes {
              name
            }
          }
        }
      }
    `,
  }),
})
  .then((result) => result.json())
  .then((result) => {
    const possibleTypes: any = {};
    result.data.__schema.types.forEach((supertype: any) => {
      if (supertype.possibleTypes) {
        possibleTypes[supertype.name] = supertype.possibleTypes.map((subtype: any) => subtype.name);
      }
    });

    fs.writeFile(OUT_FILE, JSON.stringify(possibleTypes), (err) => {
      if (err) {
        console.error("Error writing possibleTypes.json", err);
      } else {
        console.log("Fragment types successfully extracted!");
      }
    });
  });
