# `cmd/`

Beneath is not just a single executable -- it's several different "micro"services. This directory contains a subdirectory for each executable service.

## Creating an executable

- Create a directory named for the executable
- Add a `main.go` file to it
- All configuration should be loaded within the `main.go` and passed down from there (using environment variables, see existing executables for examples)
- The `main.go` file should call logic stored outside of `cmd/` (this is just the entry point, not the place to store the entire codebase)
