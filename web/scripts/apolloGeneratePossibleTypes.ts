// This script generates the JSON file apollo/fragmentTypes.json, which contains info
// about fragments in our GraphQL schema for Apollo. See this page for details:
// https://www.apollographql.com/docs/react/data/fragments/#fragments-on-unions-and-interfaces

import { API_URL} from "../lib/connection";

import fetch from "node-fetch";
import fs from "fs";

const OUT_FILE = "./apollo/possibleTypes.json";

fetch(`${API_URL}/graphql`, {
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
