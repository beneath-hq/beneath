# `engine/`

This directory contains the engine, which encapsulates logic for interfacing with the underlying data systems.

The files in directly in this repository are wrappers for functionality provided by the underlying drivers. Most of the interesting stuff happens in `engine/driver/`.

## Updating protocol buffer definitions

- Run `scripts/proto-build.sh` to (re)generate all protocol buffers classes in the repository
