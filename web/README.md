# `web/`

This directory contains the Beneath frontend (console) code. As always, refer to `docs/contributing/02-project-structure.md` for an introduction to the repository.

## Stack

The primary libraries used in the frontend are:

- [React](https://reactjs.org/): Needs no introduction

- [Next.js](https://nextjs.org/): Facilitates server-side rendered React. The advantages are increased speed and SEO. It does make some things slightly more complicated, namely code that depends on global variables (like `document` or `window`) or the `res` object in Node. 

- [Apollo GraphQL](https://www.apollographql.com/docs/react/) (React version): Used to make GraphQL calls to the control server. Also manages all the application's state (implemented correctly, it should save us from needing Redux or similar).

- [Material UI](https://material-ui.com/) (React version): Component library used for all styling and all UI elements. It's not fantastic, but stable and surprisingly flexible. It was originally centered around the Material UI aesthetic, but seems to increasingly give us the power to redefine styles.

## Structure

- `web/apollo` contains queries, types, and configs related to Apollo
- `web/components` contains standalone React components, and its subdirectories (whose names mirror pages in `web/pages` contain React components used by particular pages)
- `web/ee` contains code for the enterprise edition
- `web/hooks` contains useful, shared React hooks
- `web/lib` contains assorted configs, utilities and non-React JS code
- `web/pages` contains the distinct pages of the site (it's a Nextjs convention); note that we *don't* use nested pages to auto-generate dynamic routes (which is a recent Nextjs convention), see `web/server.js` for the routes

## Commands

### Running a development server

- First run `yarn install`
- Then run `yarn dev`

### Apollo

Regenerate types for Apollo queries (only covers queries in `web/apollo/queries/`):

```yarn apollo:generate```

## Using the `beneath` and `beneath-react` JS client libraries

The stream data exploration page uses the JS and React clients (see `clients/js` and `clients/js-react` under the repository root) to fetch data. It does *not* locally link to these, but instead uses the versions stored on NPM. It's significantly simpler and less brittle to do it this way, but it does mean that changes made in the clients aren't immediately reflected when developing the console.

If you need to edit the two simultaneously, you have to short-circuit the imports (and remember to run `tsc` first in the client libraries). Before committing, you should revert the short-circuiting code, publish the client libraries as new versions to NPM, and update the dependent versions in `web/package.json`.

## EE development

Code related to the enterprise edition is placed in a matching hierarchy in `web/ee/`.

Unlike the Beneath backend implementation, the `web` frontend implementation is not currently completely decoupled from the EE code. Concretely, we use the `IS_EE` global in `lib/connection.ts` to conditionally include EE components.
