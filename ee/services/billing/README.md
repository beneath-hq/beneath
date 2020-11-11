# `ee/services/billing/`

The billing service handles billing info and consumed resources. It keeps tabs on *what to bill*, but the `payments` service handles the actual invoicing.

Some of the billing logic is not super trivial. For sanity, we've tried to gather most of the complex logic in `commit.go`.
