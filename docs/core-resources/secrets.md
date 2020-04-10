---
title: Secrets
description:
menu:
  docs:
    parent: core-resources
    weight: 110
weight: 110
---
Secrets are Beneath's terminology for API private keys. Secrets are used to authenticate to Beneath. 

##### User secrets vs. service secrets
User secrets are used to connect via CLI or via any client library.

Service secrets are used to deploy a model or to use in the front-end of your application. Read more about services <a href="/docs/services">here</a>.

##### Permissions
Every secret has unique permissions that allow it to read and write to a fixed set of Streams. Service secrets, which are publically visible, should be constrained so that they can only read/write to the intended Streams.

##### Billing and quotas
All secrets track usage and are limited to a quota. Usages are used to calculate end-of-month bills. So don't share your secrets! 

##### CLI authentication
Authenticate your CLI session by providing your user secret. From your command line: 
```bash
beneath auth SECRET
```
Beneath stores your secret in a hidden folder on your Desktop called “.beneath”