# `services/table/`

This service manages the `Table` and `TableInstance` models (also see `models/table.go`).

It's functionality includes:

- Compiling table schemas
- Managing instance versions
- Managing storage of instances in the underlying data systems (`infra/engine/`)
- Various table metadata caches
