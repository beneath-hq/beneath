---
title: Share your data 
description: A guide to updating project-level access controls
menu:
  docs:
    parent: quick-starts
    weight: 600
weight: 600
---

Time required: 1 minute.

This quick-start assumes you've already installed the Beneath SDK, assumes that you have a Beneath project with a data stream, and assumes you have a Professional plan, which allows for private projects.

Permissions for Beneath data is managed at the project level. So, in order to share a data stream with someone, you need to give that person `view` permissions for the entirety of the data stream's project.

## Add a user to the project
If you have `admin` permissions for the project, you can run from your command line:
```bash
beneath project update-permissions ORGANIZATION_NAME/PROJECT_NAME USERNAME --view true --create true --admin false 
```

## Notes
 - You can allow anyone to view your data stream, and you don't have to worry about their usage affecting your bill. As long as the other user is in a different organization (which is the case by default), you won't be billed for any of their activity!
