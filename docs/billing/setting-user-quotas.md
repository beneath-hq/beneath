---
title: Setting user quotas
description:
menu:
  docs:
    parent: billing
    weight: 200
weight: 200
---

As an organization admin, it is possible to set custom quotas for each of your organization's members.

From your command line:
```bash
beneath organization update-quotas [--read-quota-mb READ_QUOTA_MB] [--write-quota-mb WRITE_QUOTA_MB] ORGANIZATION USERNAME
```