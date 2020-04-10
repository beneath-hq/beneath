---
title: Share a data stream with others
description:
menu:
  docs:
    parent: access-management
    weight: 300
weight: 300
---

In order to share a data stream with another user, you must give the other user read access to the stream's project.

- View permissions allow the user to view the entirety of the project and to read from any stream. Each user is billed for their own usage, so there's no need to worry about your bill when sharing projects with others.
- Create permissions allow the user to add any streams or models to the project.
- Admin permissions allow the user to delete streams or models from the project, and to manage access controls to the project.

From your command line:
```bash
project add-user yoda -p starwars --view true --create true --admin false 
```