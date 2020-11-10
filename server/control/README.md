# `server/control/`

This directory contains the implementation of the control server. It mainly stitches together the GraphQL API, permission checks, output formats, etc., as it calls services for most of the actual functionality (see `services/` for details).

## Stack

These are the main libraries used in the control server

- [`gqlgen`](https://gqlgen.com/): For defining the GraphQL server. It generates Go files based on GraphQL schema files (see below). Not the most common choice, but it seems many people think it's now the best GraphQL library for Go.
- [`goth`](https://github.com/markbates/goth): For setting up authentication with Google and Github
- [`chi`](https://github.com/go-chi/chi): For HTTP routing

## Adding GraphQL resolvers in the *control server*

1. Add/update the GraphQL schema files in `server/control/schema`
2. Run `scripts/gqlgen-control.sh`
3. This is where it gets a little tricky -- it will generate stubs for all the resolvers in the file `server/control/resolver/generated.go`. You will want to pick out all the types/resolvers that have been updated/changed and add them to one of the other resolver files in the package (e.g. if it's a new user-related query, add it to `server/control/resolver/user`)
4. When you're done manually merging the new generated resolvers into the existing resolvers, delete the `generated.go` file. Good news: if you mess anything up, the type checker will complain!
