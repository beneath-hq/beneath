---
title: Share your data 
description:
menu:
  docs:
    parent: quick-starts
    weight: 90
weight: 90
---

Time required: 1 minute.

##### 1. Download the CLI if you haven't already
From your command line:
```bash
pip install beneath
```

##### 2. Add a user to the project
From your command line:
```bash
project add-user yoda -p starwars --view true --create true --admin false 
```

##### Notes
+ Only project Admins are allowed to edit user permissions for any given project.
+ You can allow anyone to view your data stream, and you don't have to worry about their usage affecting your bill. As long as the other user is in a different organization (which is the case by default), you won't be billed for any of their activity!
+ Remember that access permissions are granted at the project level. So make sure to keep your private streams and public streams in different projects.
