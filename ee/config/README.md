# `ee/config/`

This folder contains templates for extra EE-related config keys.

Beneath doesn't support loading multiple config files, and by default only looks in the root `config/` directory for `.ENV.yaml` files (i.e. it will not pickup on a config file in `ee/config/`). So you'll want to copy the relevant keys from these templates into your main config file.
