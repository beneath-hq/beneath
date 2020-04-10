---
title: Beneath concepts
description:
menu:
  docs:
    parent: introduction
    weight: 40
weight: 40
---

The following are concepts that are core to Beneath's simplicity and ease-of-use.

##### Control plane and data plane
**Control plane**<br> 
The control plane enables the user to manage the metadata around all resources. Access controls, billing, quotas, descriptions, resource organization.

**Data plane**<br>
The data plane is a fault-tolerant, elastic, massively parallel system that can effectively handle any volume of data.

##### The data terminal and the command line interface
**The Data Terminal**<br>
The Data Terminal is the epicenter of your organization's streaming data. The Data Terminal functions as a data stream browser: you can view all public data on the platform and you can view all your private data. You can view and query live data streams. You can monitor stream uptime and usage. You can get automatically-generated live data APIs for each of your streams.

**The Command Line Interface (CLI)**<br>
The CLI is the de-facto way to manage the control plane. Use the CLI to create new streams, modify schema, manage access controls, manage quotas, etc.

##### Schema required
In order to ensure data quality, Beneath requires a schema for each data stream. This ensures sound data quality, which will trickle down to all following streams.
It ensures all databases and integrated systems won't be surprised by bad data.

##### Streams and instances
PENDING
<!-- "Batch" streams have instances. Each instance is a new batch upload. Yet, in the Beneath terminology, we refer to all data sets as streams. -->


<!-- ##### Root streams and derived streams
Root streams:
Derived streams: -->

<!-- ##### Core entities
**Organizations**.<br>
Organizations are entities that handle billing. In practice, users join their company's "organization," and the organization's admin will handle billing on behalf of all of its users.<br>

**Projects**.<br>
Projects are collections of Streams and Models. Projects are analagous to GitHub repositories. Projects can be either public - shared with the world - or private - shared with a set list of users. Every project is owned by an organization, which will be responsible for charges incured. <br>

**Streams**.<br>
Streams are unbounded, continuously updated data sets. Streams are ordered and replayable. Every stream has a key.

**Models**.<br>
Models are 

**Stream instance**.<br>
Stream instance...

**Services**.<br>
Services are... -->


<!-- Resource structure: streams, instances; grouped by projects, by organization; users vs services -->
