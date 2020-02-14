# Frontend development

The frontend is found in `web/`.

## Stack

The primary libraries used in the frontend are:

- [React](https://reactjs.org/): Needs no introduction

- [Next.js](https://nextjs.org/): Facilitates server-side rendered React. The advantages are increased speed and SEO. It does make some things slightly more complicated, namely code that depends on global variables (including `document`) or the `res` object in Node. 

- [Apollo GraphQL](https://www.apollographql.com/docs/react/) (React version): Used to make GraphQL calls to the control server. Also manages all the application's state (implemented correctly, it should save us from needing Redux or similar).

- [Material UI](https://material-ui.com/) (React version): Component library used for all styling and all UI elements. It's not fantastic, but stable and surprisingly flexible. It was originally centered around the Material UI aesthetic, but seems to increasingly give us the power to redefine styles.

## Commands

### Apollo

Regenerate types for Apollo queries (currently only covers queries in `web/apollo/queries/`):

```yarn apollo:generate```
