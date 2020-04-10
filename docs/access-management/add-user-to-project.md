---
title: Add user to project
description:
menu:
  docs:
    parent: "access-management"
    weight: 100
weight: 100
---

- View permissions allow the member to see all of the project's contents.
- Create permissions allow the member to create items within the project.
- Admin permissions allow the member to add/delete members from the project.

From your command line:
```bash
beneath project add-member myProject newMemberUsername --view true --create true --admin false 
```
