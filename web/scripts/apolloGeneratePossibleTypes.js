// This script generates the JSON file apollo/fragmentTypes.json, which contains info
// about fragments in our GraphQL schema for Apollo. See this page for details:
// https://www.apollographql.com/docs/react/data/fragments/#fragments-on-unions-and-interfaces

const connection = require("../lib/connection");

const fetch = require("node-fetch");
const fs = require("fs");

const API_HOST = connection.API_URL;
const OUT_FILE = "./apollo/possibleTypes.json";

fetch(`${API_HOST}/graphql`, {
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
    const possibleTypes = {};
    result.data.__schema.types.forEach((supertype) => {
      if (supertype.possibleTypes) {
        possibleTypes[supertype.name] = supertype.possibleTypes.map((subtype) => subtype.name);
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
