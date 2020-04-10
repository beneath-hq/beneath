---
title: Unified infrastructure
description:
menu:
  docs:
    parent: introduction
    weight: 20
weight: 20
---

In order to provide a seamless data experience, Beneath is built with a comprehensive stack of database technologies. The Beneath system is an all-in-one package: no need to configure, manage, or integrate any of these services.

##### Gateway server
A gateway server functions as a data validator, API server, and load balancer.

##### Message queue
A message queue ties the whole system together. The queue eliminates data loss by ensuring every data packet is delivered to each destination at least once.

##### Stream processor
A stream processor enables continuous data transformations on event streams. The Beneath processor elastically scales its resources and is massively parallel.

##### Data warehouse
An OLAP database allows for large-scale analytical queries.

##### Operational datastore
A key-value store provides low-latency lookups for apps and dashboards.

##### Control plane
An all-encompassing control plane enables access management, data and model management, and payments.

