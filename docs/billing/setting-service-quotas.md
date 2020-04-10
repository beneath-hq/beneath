---
title: Setting service quotas
description:
menu:
  docs:
    parent: billing
    weight: 300
weight: 300
---

It is good practice to set custom, reasonable quotas for each of your services in order to prevent any unexpected behavior.

From your command line:
```bash
beneath service update [-o ORGANIZATION] [--read-quota-mb READ_QUOTA_MB] [--write-quota-mb WRITE_QUOTA_MB] SERVICENAME
```
