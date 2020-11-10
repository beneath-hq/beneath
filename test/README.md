# `test/`

This directory contains end-to-end/integration tests

### Running these tests

The tests must run at a package level to include the `TestMain` wrapper. The `BENEATH_ENV` environment variable should also be set to `test`. From the root directory, run tests with

```
BENEATH_ENV=test go test ./test/integration/
```
