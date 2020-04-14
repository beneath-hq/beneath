---
title: Share your data 
description:
menu:
  docs:
    parent: quick-starts
    weight: 400
weight: 400
---

Time required: 1 minute.

This quick-start assumes you've already installed the Beneath SDK and assumes that you have a Beneath project with a data stream.

##### 1. Add a user to the project
From your command line:
```bash
project add-user yoda -p starwars --view true --create true --admin false 
```

##### Notes
+ Only project Admins are allowed to edit user permissions for any given project.
+ You can allow anyone to view your data stream, and you don't have to worry about their usage affecting your bill. As long as the other user is in a different organization (which is the case by default), you won't be billed for any of their activity!
+ Remember that access permissions are granted at the project level. So make sure to keep your private streams and public streams in different projects.
