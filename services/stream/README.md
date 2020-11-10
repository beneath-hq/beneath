# `services/stream/`

This service manages the `Stream` and `StreamInstance` models (also see `models/stream.go`).

It's functionality includes:

- Compiling stream schemas
- Managing instance versions
- Managing storage of instances in the underlying data systems (`infra/engine/`)
- Various stream metadata caches
