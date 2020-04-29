---
title: Resources
description: An overview of every resource that can be created and managed in Beneath
menu:
  docs:
    parent: managing-resources
    weight: 300
weight: 300
---

## `user`

**Definition:** A `user` represents a real person who operates on Beneath. 

**Relations:**
- A `user` belongs to one `organization`, which handles billing for the `user`. If the `user` doesn't belong to an `organization`, a "personal" `organization` is automatically and transparently created for the `user` with the same name as the `user`'s username.
- A `user` has many (zero or more) `secret`s.

**Access management:**
- A `user` can be granted access to an `organization`.
- A `user` can be granted access to a `project`.

**View:**
- *Terminal:* Go to `https://beneath.dev/USERNAME`
- *CLI*: (Not available)

## `organization`

**Definition:** An `organization` is the top-level owner of users, services and projects. Billing is managed at the `organization` level, which means that every resource that can accrue bills is directly or indirectly linked to exactly one `organization`. 

**Relations:**
- An `organization` has many (one or more) `user`s. Since all `user`s must belong to an `organization`, when a `user` is created, they're added to a "personal" `organization` that is automatically and transparently created with the same name as the username.
- An `organization` has many (zero or more) `project`s.
- An `organization` has many (zero or more) `service`s.

**Access management:**
- A `user` can access an `organization`.
  - The `view` permission grants the `user` permission to browse the members and projects in the organization.
  - The `admin` permission grants the `user` permission to add and delete members, create new projects, create services and change billing information.

**View:**
- *Terminal:* Go to `https://beneath.dev/ORGANIZATION_NAME`
- *CLI:* Run `beneath organization --help` for details.

## `project`

**Definition:** A `project` is a collection of `stream`s. You can think of them like repositories in Git.

**Relations:**
- A `project` belongs to one `organization`.
- A `project` has many (zero or more) `stream`s.

**Access management:**
- A `user` can access a `project`.
  - The `view` permission grants the `user` permission to browse the contents of the project, including viewing and querying records in its `stream`s.
  - The `create` permission grants the `user` permission to create and delete `stream`s in the project, including writing data directly to root `stream`s.
  - The `admin` permission grants the `user` permission to add, remove and change permissions for other `user`s.

**View:**
- *Terminal:* Go to `https://beneath.dev/ORGANIZATION_NAME/PROJECT_NAME`
- *CLI:* Run `beneath project --help` for details.

## `stream`

**Definition:** A `stream` is the prototype of a collection of records with a common schema. For more information about `stream`s and the related `stream instance`s, see [Streams]({{< ref "/docs/reading-writing-data/streams" >}}).

**Relations:**
- A `stream` belongs to one `project`.
- A `stream` has many (zero or more) `stream instance`s.

**Access management:**
- A `user` can access a `stream` through its permissions for the parent `project`.
  - The `view` permission on `project` grants permission to view and query records.
  - The `create` permission on `project` grants permission to write records.
- A `service` can access a `stream`. (Note the difference: `service`s have direct permissions for a `stream`, while `user`s get indirect permissions on a `project`-level)
  - The `read` permission grants the `service` permission to read and query records.
  - The `write` permission grants the `service` permission to write records.

**View:**
- *Terminal:* Go to `https://beneath.dev/ORGANIZATION_NAME/PROJECT_NAME/streams/STREAM_NAME`
- *CLI:* Run `beneath stream --help` for details.

## `stream instance`

**Definition:** A `stream instance` represents a single version of a `stream`. For more information, see [Streams]({{< ref "/docs/reading-writing-data/streams >}}).

**Relations:**
- A `stream instance` belongs to one `stream`.

**Access management:** A `stream instance` inherits the permissions of its parent `stream`.

**View:**
- *Terminal:* Go to `https://beneath.dev/ORGANIZATION_NAME/PROJECT_NAME/streams/STREAM_NAME` (only shows the primary `stream instance`)
- *CLI:* Run `beneath stream instance --help` for details.

## `service`

**Definition:** A `service` represents a system with access to read or write data to Beneath. You can think of a `service`s as a `user` for your code. They're especially useful for creating secrets that you can use in your code to read and write to Beneath in a safe way.

A `service` has the following properties:
- You grant it custom access permissions (on a stream level) that are not tied to the permissions of a specific user
- You can create secrets for the service, which you embed in your code to use Beneath
- You get usage metrics (reads and writes) for the service
- You can set usage limits (reads and writes) for the service on a monthly basis

**Relations:**
- A `service` belongs to one `organization`, which handles billing for the `service`.
- A `service` has many (zero or more) `secret`s.

**Access management:**
- A `service` can be granted access to a `stream`.

**View:**
- *Terminal:* (Not available)
- *CLI:* Run `beneath service --help` for details.

## `secret`

**Definition:** A `secret` is a token that you can use to authenticate to Beneath (some products call it an *API token*). It belongs to either a `user` or a `service`. When you authenticate with a secret, you get the same access permissions as the parent `user` or `service` (with the caveat that you can create special read-only secrets for a `user`).

If you need to expose a secret publicly (e.g. in your front-end code), make sure it belongs to a service with sensible usage quotas and only read-only permissions.

**Relations:**
- A `secret` belongs to *either* a `user` or a `service`.

**Access management:**
- A `user` can create `secret`s for themself.
- A `user` can create `secret`s for `service`s that belong to an organization that they have `admin` permissions on.

**View:**
- *Terminal:* For user-owned `secret`s, go to `https://beneath.dev/USERNAME/-/secrets`. (Not available for service-owned `secret`s.)
- *CLI:* For service-owned `secret`s, run `beneath service --help` for details. (Not available for user-owned `secret`s.)
