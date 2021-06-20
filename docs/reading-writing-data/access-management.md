---
title: Access management
description: A guide to managing access to tables in Beneath
menu:
  docs:
    parent: reading-writing-data
    weight: 600
weight: 600
---

The gist of access management in Beneath is as follows: [users]({{< ref "/docs/misc/resources.md#users" >}}) and [services]({{< ref "/docs/misc/resources.md#services" >}}) can be granted permissions on resources like tables, projects and organizations. Different resources have different permissions, like `read` and `write` for tables. Additionally, [projects]({{< ref "/docs/misc/resources.md#projects" >}}) can be marked `public`, which lets anyone access their tables.

You issue and use [secrets]({{< ref "/docs/misc/resources.md#secrets" >}}) to authenticate as a _user_ or a _service_ from your code, notebook, command-line, or similar. _User_ secrets are useful during development to connect to Beneath from your code. _Service_ secrets should be used in production systems or publicly-exposed code to more strictly limit access permissions and monitor usage.

## Examples

### Authenticating in Beneath CLI

See [the CLI installation guide]({{< ref "/docs/quick-starts/install-sdk.md" >}}).

### Creating a secret for a user

You can issue and copy a new secret from the ["Secrets" tab of your profile page](https://beneath.dev/-/redirects/secrets) in the Beneath Console.

There are three types of secrets:

- **Full (CLI) secrets:** These have full access to your user, and are normally used for command-line authentication
- **Private read secrets:** These are limited to `view` permissions on all the resources you have access to
- **Public read secrets:** These are limited to `view` permissions on public projects and tables on Beneath

**Never share your user secrets!** You should not share or expose user secrets nor use them in production systems. Use a service secret if you need a secret in a production system or if you need to expose a secret (e.g. in a shared notebook or in your frontend code).

### Creating services, setting quotas and granting permissions

Services are useful when deploying or publishing code that reads or writes to Beneath. You can control their access permissions, and monitor and limit their usage. Read ["Services"]({{< ref "/docs/misc/resources.md#services" >}}) for more details.

By default, a service cannot read any tables, and you must set all service permissions manually, _including permissions for public tables_ (unlike users).

You can create services in the web console or using the CLI. For example, to create a service and grant read and write permissions to a table from the CLI, run (change or remove the flags depending on your use case):

```bash
beneath service create ORGANIZATION/PROJECT/NEW_SERVICE --read-quota-mb 100 --write-quota-mb 100
beneath service update-permissions ORGANIZATION/PROJECT/NEW_SERVICE ORGANIZATION/PROJECT/STREAM --read --write
```

You're now ready to issue a secret for the service.

### Creating a secret for a service

You can create secrets for services using the web console or the CLI. For example, to create a service secret from the CLI, run the following command:

```bash
beneath service issue-secret ORGANIZATION/PROJECT/SERVICE --description "YOUR SECRET DESCRIPTION"
```

You can now use the secret to connect to Beneath from your code. Most client libraries will automatically use your secret if you set it in the `BENEATH_SECRET` environment variable (see the documentation for your client library for other ways of passing the secret).

**Think carefully before sharing service secrets!** If you need to expose a secret publicly (e.g. in your front-end code or in a notebook), make sure it belongs to a service with sensible usage quotas and only `read` permissions. In all other cases, keep your secret very safe and do not commit it into Git.

### Inviting a user to an organization

> **Note:** When you _invite_ a user to your organization, you take over the full billing responsibility for the user's activity. If you just want to grant the user access to view stuff in your organization (but not pay their bills), see [Granting a user access to an organization]({{< relref "#granting-a-user-access-to-an-organization" >}}).

First, an organization admin should send an invitation to the user (the flags designate the invitees new permissions). From the CLI, run:

```bash
beneath organization invite-member ORGANIZATION USERNAME --view --create --admin
```

Second, the invited user should accept the invitation, also using the CLI:

```bash
beneath organization accept-invite ORGANIZATION
```

### Granting a user access to an organization

> **Note:** When you _grant_ a user access to your organization, they remain responsible for paying the bills for their own usage on Beneath. If you also want to pay for their usage on Beneath, see [Inviting a user to an organization]({{< relref "#inviting-a-user-to-an-organization" >}}).

Run the following command (change the flags to configure permissions):

```bash
beneath organization update-permissions ORGANIZATION USERNAME --view --create --admin
```

The user doesn't have to be a part of the organization in advance.

### Granting a user access to a project and its tables

In Beneath, _user_ access to tables is managed at the project-level. You cannot grant a user access to only one table (however, if you need a secret with permissions for just a single table, use a _service_).

To add a user to a project, use the Beneath CLI to run the following command (change the flags to configure permissions):

```bash
beneath project update-permissions ORGANIZATION/PROJECT USERNAME --view --create --admin
```

The user doesn't have to be a part of the same organization as you or the project to get access.
