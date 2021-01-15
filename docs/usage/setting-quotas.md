---
title: Custom quotas
description: A guide to setting read and write quotas on individual users and services
menu:
  docs:
    parent: billing
    weight: 200
weight: 200
---

TODO

## Services

It is good practice to set custom, reasonable quotas for each of your services in order to prevent any unexpected behavior.

From your command line:
```bash
beneath service update [-o ORGANIZATION] [--read-quota-mb READ_QUOTA_MB] [--write-quota-mb WRITE_QUOTA_MB] SERVICENAME
```

## Users

As an organization admin, it is possible to set custom quotas for each of your organization's members.

From your command line:
```bash
beneath organization update-quotas [--read-quota-mb READ_QUOTA_MB] [--write-quota-mb WRITE_QUOTA_MB] ORGANIZATION USERNAME
```
