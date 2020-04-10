---
title: Add user to organization
description:
menu:
  docs:
    parent: access-management
    weight: 200
weight: 200
---

- View permissions are for all users whose billing will be handled by the organization.
- Admin permissions permit changing the billing information and organization membership.

##### First, the organization admin must send the user an invitation
From the organization's admin's command line:
```bash
beneath organization invite-member newMemberUsername --view true --admin false 
```
##### Second, the invited user must accept the invitation
From the invited user's command line:
```bash
beneath organization accept-invite myNewOrganization
```