---
title: Usage, monitoring and quotas
description: A description of how Beneath tracks and monitors usage
menu:
  docs:
    parent: usage
    weight: 100
weight: 100
---

Beneath tracks table reads, writes and scans to provide granular usage monitoring and billing.

Every table, user and service in Beneath has a "Monitoring" tab in the web [console](https://beneath.dev/?noredirect=1) where you can see usage breakdowns, including hour-by-hour usage and quota usage in the current period.

## How usage is calculated

For **reads and writes**, the usage in _bytes_ is calculated as the total [Avro](https://en.wikipedia.org/wiki/Apache_Avro)-encoded bytes transferred. Avro uses a table's schema to achieve a very compact encoding, which can easily be 80% smaller than JSON. Using Avro-encoded size means you can expect to get a higher mileage out of your Beneath quota than you might expect. For easy comparison, we count the Avro size even for requests that use the JSON-based REST API for tables.

For SQL warehouse query **scans**, the usage depends on the underlying data warehouse, but will typically be calculated based on the number of bytes loaded from a compressed columnar format (so if your query involves fewer columns, it's likely to scan fewer bytes).

## Quotas and quota periods

Organizations, users and services have usage quotas to prevent abuse or unexpected bills. Quotas operate on 31-day cycles starting on the day they're set. Your billing plan dictates your top-level quotas, but you can set lower quotas for specific services (or for specific users in multi-user organizations).

When you set a quota or change billing plans, the quota usage resets and the quota period cycle is updated to start at the change time. Note that since quota usage periods depend on the date the quota is set, your billing quota period and service-specific quota periods will likely differ and cannot be directly compared.

## Service quotas

It is especially good practice to set custom, reasonable quotas for your [services]({{< ref "/docs/misc/resources.md#services" >}}) in order to prevent any unexpected use.

You can set quotas from the web console or the command-line. To set a quota from the command-line, run the following command for details:

```bash
beneath service update -h
```
