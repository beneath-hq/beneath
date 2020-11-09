# `bus/`

The bus facilitates sync (in-process) and async (background worker) event dispatch and handling. Services (in `services/`) send and subscribe to events messages, which are defined in `models/`.

The bus was inspired by the [communication model in Grafana](https://github.com/grafana/grafana/blob/master/contribute/architecture/communication.md).

## Sync and async handlers

- *Sync event handlers* are called sequentially in-process when an event is published, and causes the event publisher to fail on error. If an event is sent in a DB transaction, subscribed handlers will also run inside that transaction!
- *Async event handlers* are called in a background worker. They retry on error. Event messages are serializes with msgpack and passed through the MQ (`infra/mq/`).

Use sync event handlers for light operations or operations that require the publisher to fail on error. Use async event handlers for everything else.

## Background worker

The background worker calls `bus.Run` to start processing async events dispatched from other processes. It's started with the `control-worker` startable (see `cmd/beneath/`).

**It's critical that every service that handles async events is initialized in the worker process** before bus.Run is called! (See `cmd/beneath/dependencies/services.go`.)
