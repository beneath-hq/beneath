# `ee/`

This folder contains the backend code for the Enterprise Edition (EE) of Beneath (the frontend EE code is found in `web/ee/`). It carries a different license than the rest of the repository.

The Community Edition (CE) can be compiled without including this folder. We achieve this separation by using different top-level entrypoints. To run the EE version, run `ee/cmd/beneath/main.go` instead of `cmd/beneath/main.go`. Easy!

The `ee/` folder has a project layout that matches the root layout. So it's always worth inspecting the corresponding root folder for details when editing the EE version. The corresponding non-EE version will typically have more details on the implementation.

Here's a rough overview of how the EE edition code works:

- It has a separate migrator and migrations in `ee/migrations/` that tracks migration versions on the EE-only models (in `ee/models/`) separately (in the table `gopg_migrations_ee`). The EE-models and their migrations must be decoupled from the CE models.
- The `ee/cmd/beneath/main.go` file wraps `cmd/beneath/` and takes care of wiring up the extra dependencies and making sure they're created/registered for the relevant startables. It leverages the extensible nature of `cmd/beneath/cli/` to accomplish this neatly.
- It does not extend the CE control server's GraphQL schema, but instead defines an extra GraphQL server at `/ee/graphql` on the control server's router, which hosts the EE-related server functionality found in `ee/server/control/`. The EE frontend multiplexes between the two CE and EE endpoints (Apollo supports connecting to multiple GraphQL backends from one client).
