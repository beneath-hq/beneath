---
title: Models
description:
menu:
  docs:
    parent: core-resources
    weight: 400
weight: 400
---
One of Beneath’s core features is the ability to create _models_ to manipulate data streams. Models enable both stateless and stateful stream processing.

With models, you can apply any function to any data stream. Functions like:

- Group by and aggregate
- Filter
- Apply labels (using custom rules or using a machine learning model)
- Join multiple streams
- Apply custom functions

##### CLI commands
With the CLI, you can do the following:

- initialize a model
- simulate a model
- stage a model
- deploy a model

See the manual:
```bash
beneath model -h
```
