# `control/`

This directory contains the code for the control server, including the Postgres entities, GraphQL resolvers, and background task queue. As always, refer to `docs/contributing/02-project-structure.md` for an introduction to the repository.

## Stack

These are the main libraries used in the control server

- [`go-pg`](https://github.com/go-pg/pg): For interacting with the Postgres database. It's vaguely like an ORM. See the [wiki](https://github.com/go-pg/pg/v9/wiki) for good examples on how to use it, especially the ["Model Definition"](https://github.com/go-pg/pg/v9/wiki/Model-Definition) and ["Writing Queries"](https://github.com/go-pg/pg/v9/wiki/Writing-Queries) pages. We're using [this helper library](https://github.com/go-pg/migrations/v7) to run migrations (see below). 
- [`gqlgen`](https://gqlgen.com/): For defining the GraphQL server. It generates Go files based on GraphQL schema files (see below). Not the most common choice, but it seems many people think it's now the best GraphQL library for Go.
- [`goth`](https://github.com/markbates/goth): For setting up authentication with Google and Github
- [`chi`](https://github.com/go-chi/chi): For HTTP routing

## Adding and running migrations

1. Add a migration file to `beneath-core/control/migrations`
2. Update the relevant model(s) in `beneath-core/control/entity` to match the migration
3. The migration will automatically be applied when you start the control server
4. (There is also a migration tool, which you can run with `go run cmd/control_migrate/main.go XXX`, where XXX can be `up`, `down` and `reset` -- `reset` is especially useful during development)

## Adding GraphQL resolvers in the *control server*

1. Add/update the GraphQL schema files in `beneath-core/control/gql/schema`
2. Run `scripts/gqlgen-control.sh`
3. This is where it gets a little tricky -- it will generate stubs for all the resolvers in the file `beneath-core/control/resolver/generated.go`. You will want to pick out all the types/resolvers that have been updated/changed and add them to one of the other resolver files in the package (e.g. if it's a new user-related query, add it to `beneath-core/control/resolver/user`)
4. When you're done manually merging the new generated resolvers into the existing resolvers, delete the `generated.go` file. Good news: if you mess anything up, the type checker will complain!

## Adding new tasks

1. Create a file in `control/entity` with named by the pattern `[VERB][ACTION]Task`, e.g. `CleanupInstanceTask`
2. Add the task payload as fields in the struct
2. Register the task in the `init()` function with `taskqueue.RegisterTask`
3. Implement `Run(ctx context.Context) error` on the task struct with the logic you want to execute in the background
4. When you want to trigger the background task, call `taskqueue.Submit(t)` where t is the instantiated task struct

## Payments

### Changing payment plans

TODO: These instructions are do not cover all cases.

- To upgrade a customer to Enterprise, submit an http request to http://localhost:4000/billing/stripewire/initialize_customer with these parameters: organizationID, billingPlanID, emailAddress

- To downgrade a customer to Free, submit an http request to http://localhost:4000/billing/anarchism/initialize_customer with these parameters: organizationID, billingPlanID
 
### Stripe

To get a grip on how the Stripe Go library works, check out the [Usage](https://github.com/stripe/stripe-go#usage) section.

To test Stripe's webhooks in your local development environment, execute this command in a terminal: `stripe listen --forward-to localhost:4000/billing/stripecard/webhook`.
