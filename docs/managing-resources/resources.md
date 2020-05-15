---
title: Resources
description: An overview of every resource that can be created and managed in Beneath
menu:
  docs:
    parent: managing-resources
    weight: 300
weight: 300
---

## Users

**Definition:** A [user]({{< relref "#users" >}}) represents a real person who operates on Beneath. 

**Relations:**
- A [user]({{< relref "#users" >}}) belongs to one [organization]({{< relref "#organizations" >}}), which handles billing for the [user]({{< relref "#users" >}}). If the [user]({{< relref "#users" >}}) doesn't belong to an [organization]({{< relref "#organizations" >}}), a "personal" [organization]({{< relref "#organizations" >}}) is automatically and transparently created for the [user]({{< relref "#users" >}}) with the same name as the [user]({{< relref "#users" >}})'s username.
- A [user]({{< relref "#users" >}}) has many (zero or more) [secrets]({{< relref "#secrets" >}}).

**Access management:**
- A [user]({{< relref "#users" >}}) can be granted access to an [organization]({{< relref "#organizations" >}}).
- A [user]({{< relref "#users" >}}) can be granted access to a [project]({{< relref "#projects" >}}).

**Terminal:** Go to `https://beneath.dev/USERNAME`

**CLI:** (Not available)

## Organizations

**Definition:** An [organization]({{< relref "#organizations" >}}) is the top-level owner of [users]({{< relref "#users" >}}), [services]({{< relref "#services" >}}) and [projects]({{< relref "#projects" >}}). Billing is managed at the [organization]({{< relref "#organizations" >}}) level, which means that every resource that can accrue bills is directly or indirectly linked to exactly one [organization]({{< relref "#organizations" >}}). 

**Relations:**
- An [organization]({{< relref "#organizations" >}}) has many (one or more) [users]({{< relref "#users" >}}). Since all [users]({{< relref "#users" >}}) must belong to an [organization]({{< relref "#organizations" >}}), when a [user]({{< relref "#users" >}}) is created, they're added to a "personal" [organization]({{< relref "#organizations" >}}) that is automatically and transparently created with the same name as the username.
- An [organization]({{< relref "#organizations" >}}) has many (zero or more) [projects]({{< relref "#projects" >}}).
- An [organization]({{< relref "#organizations" >}}) has many (zero or more) [services]({{< relref "#services" >}}).

**Access management:**
- A [user]({{< relref "#users" >}}) can access an [organization]({{< relref "#organizations" >}}).
  - The `view` permission grants the [user]({{< relref "#users" >}}) permission to browse the members and projects in the organization.
  - The `admin` permission grants the [user]({{< relref "#users" >}}) permission to add and delete members, create new projects, create services and change billing information.

**Terminal:** Go to `https://beneath.dev/ORGANIZATION_NAME`

**CLI:** Run `beneath organization --help` for details.

## Projects

**Definition:** A [project]({{< relref "#projects" >}}) is a collection of [streams]({{< relref "#streams" >}}). You can think of them like repositories in Git.

**Relations:**
- A [project]({{< relref "#projects" >}}) belongs to one [organization]({{< relref "#organizations" >}}).
- A [project]({{< relref "#projects" >}}) has many (zero or more) [streams]({{< relref "#streams" >}}).

**Access management:**
- A [user]({{< relref "#users" >}}) can access a [project]({{< relref "#projects" >}}).
  - The `view` permission grants the [user]({{< relref "#users" >}}) permission to browse the contents of the project, including viewing and querying records in its [streams]({{< relref "#streams" >}}).
  - The `create` permission grants the [user]({{< relref "#users" >}}) permission to create and delete [streams]({{< relref "#streams" >}}) in the project, including writing data directly to root [streams]({{< relref "#streams" >}}).
  - The `admin` permission grants the [user]({{< relref "#users" >}}) permission to add, remove and change permissions for other [users]({{< relref "#users" >}}).

**Terminal:** Go to `https://beneath.dev/ORGANIZATION_NAME/PROJECT_NAME`

**CLI:** Run `beneath project --help` for details.

## Streams

**Definition:** A [stream]({{< relref "#streams" >}}) is the prototype of a collection of records with a common schema. For more information about [streams]({{< relref "#streams" >}}) and the related [stream instances]({{< relref "#stream-instances" >}}), see [Streams]({{< ref "/docs/reading-writing-data/streams" >}}).

**Relations:**
- A [stream]({{< relref "#streams" >}}) belongs to one [project]({{< relref "#projects" >}}).
- A [stream]({{< relref "#streams" >}}) has many (zero or more) [stream instances]({{< relref "#stream-instances" >}}).

**Access management:**
- A [user]({{< relref "#users" >}}) can access a [stream]({{< relref "#streams" >}}) through its permissions for the parent [project]({{< relref "#projects" >}}).
  - The `view` permission on [project]({{< relref "#projects" >}}) grants permission to view and query records.
  - The `create` permission on [project]({{< relref "#projects" >}}) grants permission to write records.
- A [service]({{< relref "#services" >}}) can access a [stream]({{< relref "#streams" >}}). (Note the difference: [services]({{< relref "#services" >}}) have direct permissions for a [stream]({{< relref "#streams" >}}), while [users]({{< relref "#users" >}}) get indirect permissions on a [project]({{< relref "#projects" >}})-level)
  - The `read` permission grants the [service]({{< relref "#services" >}}) permission to read and query records.
  - The `write` permission grants the [service]({{< relref "#services" >}}) permission to write records.

**Terminal:** Go to `https://beneath.dev/ORGANIZATION_NAME/PROJECT_NAME/STREAM_NAME`

**CLI:** Run `beneath stream --help` for details.

## Stream instances

**Definition:** A [stream instance]({{< relref "#stream-instances" >}}) represents a single version of a [stream]({{< relref "#streams" >}}). For more information, see [Streams]({{< ref "/docs/reading-writing-data/streams" >}}).

**Relations:**
- A [stream instance]({{< relref "#stream-instances" >}}) belongs to one [stream]({{< relref "#streams" >}}).

**Access management:** A [stream instance]({{< relref "#stream-instances" >}}) inherits the permissions of its parent [stream]({{< relref "#streams" >}}).

**Terminal:** Go to `https://beneath.dev/ORGANIZATION_NAME/PROJECT_NAME/STREAM_NAME` (only shows the primary [stream instance]({{< relref "#stream-instances" >}}))

**CLI:** Run `beneath stream instance --help` for details.

## Services

**Definition:** A [service]({{< relref "#services" >}}) represents a system with access to read or write data to Beneath. You can think of a [services]({{< relref "#services" >}}) as a [user]({{< relref "#users" >}}) for your code. They're especially useful for creating secrets that you can use in your code to read and write to Beneath in a safe way.

A [service]({{< relref "#services" >}}) has the following properties:
- You grant it custom access permissions (on a stream level) that are not tied to the permissions of a specific user
- You can create secrets for the service, which you embed in your code to use Beneath
- You get usage metrics (reads and writes) for the service
- You can set usage limits (reads and writes) for the service on a monthly basis

**Relations:**
- A [service]({{< relref "#services" >}}) belongs to one [organization]({{< relref "#organizations" >}}), which handles billing for the [service]({{< relref "#services" >}}).
- A [service]({{< relref "#services" >}}) has many (zero or more) [secrets]({{< relref "#secrets" >}}).

**Access management:**
- A [service]({{< relref "#services" >}}) can be granted access to a [stream]({{< relref "#streams" >}}).

**Terminal:** (Not available)

**CLI:** Run `beneath service --help` for details.

## Secrets

**Definition:** A [secret]({{< relref "#secrets" >}}) is a token that you can use to authenticate to Beneath (some products call it an *API token*). It belongs to either a [user]({{< relref "#users" >}}) or a [service]({{< relref "#services" >}}). When you authenticate with a secret, you get the same access permissions as the parent [user]({{< relref "#users" >}}) or [service]({{< relref "#services" >}}) (with the caveat that you can create special read-only secrets for a [user]({{< relref "#users" >}})).

If you need to expose a secret publicly (e.g. in your front-end code), make sure it belongs to a service with sensible usage quotas and only read-only permissions.

**Relations:**
- A [secret]({{< relref "#secrets" >}}) belongs to *either* a [user]({{< relref "#users" >}}) or a [service]({{< relref "#services" >}}).

**Access management:**
- A [user]({{< relref "#users" >}}) can create [secrets]({{< relref "#secrets" >}}) for themself.
- A [user]({{< relref "#users" >}}) can create [secrets]({{< relref "#secrets" >}}) for [services]({{< relref "#services" >}}) that belong to an organization that they have `admin` permissions on.

**Terminal:** For user-owned [secrets]({{< relref "#secrets" >}}), go to `https://beneath.dev/USERNAME/-/secrets`. (Not available for service-owned [secrets]({{< relref "#secrets" >}}).)

**CLI:** For service-owned [secrets]({{< relref "#secrets" >}}), run `beneath service --help` for details. (Not available for user-owned [secrets]({{< relref "#secrets" >}}).)
