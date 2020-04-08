# `test/`

This directory contains end-to-end/integration tests

### Running these tests

The tests must run at a package level to include the `TestMain` wrapper. The `env` environment variable should also be configured to `test`. From the root directory, run tests with

```
ENV=test go test ./test/integration/
```
