# `ee/cmd/beneath/`

This folder contains the main entrypoint for the EE version of Beneath.

It's `main.go` is largely a duplicate of `cmd/beneath/main.go`, but registers extra EE-related commands and dependencies (namely `ee/server/control/` and `ee/services/`). When making changes here, make sure to consider if they should be mirrored in `cmd/beneath` (or vice versa!).
