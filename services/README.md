# `services/`

As always, refer to `contributing/02-project-structure.md` for an introduction to the repository.

In Beneath, "services" are domain-specific packages for operations involving persistant data. They operate on models and infrastructure (`models/` and `infra/`), and communicate via the bus (`bus/`) or direct dependency injection.

Services *do not* necessarily run in separate processes, although that is a useful way to think about them. 

Ideally, a service should manage its domain in isolation and not need to call other services. Cross-service functionality should be facilitated by services listening for and reacting to events (which are published on the bus). In practice, it's fine to violate this ideal and inject a service as a dependency to another service if that makes sense.

## Background task processing

It's often useful to be able to run some processing in a background/worker process. We effectively accomplish background task processing with async listeners to bus events (`bus/`).
