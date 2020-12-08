# `clients/`

This directory contains all the language-specific client libraries for interfacing with Beneath from external applications.

Rules for clients:

- Each client should be named after the language and possibly framework, e.g., `js-react`
- Each client should contain a `README.md` file for external users and a `CONTRIBUTING.md` file for contributors
- For documentation, we're inspired by [how Google does it](https://cloud.google.com/pubsub/docs/reference/libraries)), which means:
  - Each client should generate API reference docs at a standalone site
  - Tutorials and explainers should go in the main Beneath documentation (i.e., the root `docs/` folder)
  - Example projects should go in the root `examples/` folder
