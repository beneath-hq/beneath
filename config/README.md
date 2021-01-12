# `configs/`

This directory contains files for configuration and settings.

In Beneath, you can set config keys usding the `--config FILE` flag or by setting environment variables. 

`config.yaml` contains a base template of configurable keys and their default values. It is generated from the ConfigKeys registered in `cmd/beneath/`. It does not include driver-specific keys and EE-related keys, which you need to paste into your actual config file directly.

## Configuring dev and test environments

Copy `config.yaml` to `.development.yaml` in this folder, and Beneath will automatically load it when you run it with `BENEATH_ENV=dev`. The same works for `.test.yaml` and `.production.yaml`. If necessary, you can put configuration shared between environments in `.default.yaml`. These files are ignored by git so they can contain local secrets.

Until we have a better local development setup, a core team member can help you with configuration. See `contributing/03-development-environment.md` for more.

## Regenerating `config.yaml`

The following command regenerates `config.yaml`:

```
go run cmd/beneath/main.go config generate config/config.yaml
```
