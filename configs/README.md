# `configs/`

This directory contains files for configuration and settings.

## `.env` files

In Beneath, all configuration happens via environment variables. In production, these are set via Kubernetes manifests and secrets. However, in development, it's useful to store them in flat `.env` files, which are loaded on launch.

The `example.env` file contains all environment variables used across the entire project. It's the authoritative source on what environment variables need to be set for which components. In development, copy it to `.test.env` and `.development.env` and set values for all of them. These files are ignored by git so they can contain secrets.

Until we have a better local development setup, a core team member can help you with configuration. See `docs/contributing/03-development-environment.md` for more.
