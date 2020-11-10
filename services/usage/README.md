# `services/usage/`

This service handles reading and tracking of quota usage (i.e. reads, writes and scans).

Note that if you use the service to *track* usage (and not just read usage), it has a worker component that must `Run` concurrently to properly flush tracked usage in the background.

It's used to *read and track* usage by the `data-server`, and to *read* usage metrics by the `control-server`.
