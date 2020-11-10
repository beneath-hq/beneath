# `migrations/`

This folder contains control-level migrations for Postgres. 

All migrations should be SQL-based, and not use `go-pg`s features for automatically creating tables based on structs (since that makes future changes harder).

Migrations can be run and manipulated using the backend CLI, run `go run cmd/beneath/main.go migrate -h` for details. Migrations currently also auto-run when a control server is started. Migrations are tracked in the table `gopg_migrations`.

We use the [`go-pg` migrations library](https://github.com/go-pg/migrations/v7) to run migrations. 
