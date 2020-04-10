---
title: Services
description:
menu:
  docs:
    parent: core-resources
    weight: 500
weight: 500
---
Services, and their corresponding secrets, are entities used when specific, limited credentials are required. 
The use of *service* secrets allows you to protect your *user* secret from abuse. Additionally, service secrets can be allocated a fixed usage quota.

##### Examples
Service secrets should be used for:

**Models**. A model's service secret should be granted the minimum required permissions in order for the model to function.

**Web app front-ends**. A service secret in your web app's front-end will allow you to read/write data to Beneath. 
But, because the secret will be publically viewable, in order to to prevent abuse, the service secret should be granted the minimum required permissions in order for the web app to operate.

##### CLI commands
With the CLI, you can do the following:

- list services 
- create a service 
- update a service's description 
- update a service's permissions 
- issue a service secret 
- list service secrets 
- revoke a service secret 
- migrate a service to a new organization 
- delete a service

See the manual:
```bash
beneath service -h
```