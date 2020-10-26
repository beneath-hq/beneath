# `models/`

This folder contains structs for database models and bus events. It also contains simple implementation code, e.g. for validation or for adhering to interfaces, but it should not contain complex functionality (certainly not code that calls services or infrastructure).

When you add or change a model, remember to add related migrations to `migrations/`

`bus.Bus` marshals async events with `msgpack`, so every struct used in events should have sensible `msgpack` struct tags (eg. use `msgpack:"-"` for sub-objects like database relationships).
