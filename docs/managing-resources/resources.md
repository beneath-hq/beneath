---
title: Resources
description: An overview of every resource that can be created and managed in Beneath
menu:
  docs:
    parent: managing-resources
    weight: 300
weight: 300
---

## `User`

**Definition:** A `user` represents a real person who operates on Beneath. 

**Relations:**
- A `user` belongs to one `organization`, which handles billing for the `user`. If the `user` doesn't belong to an `organization`, a "personal" `organization` is automatically and transparently created for the `user` with the same name as the `user`'s username.
- A `user` has many `secret`s.

**Access management:**
- A `user` can be granted access to an `organization`.
- A `user` can be granted access to a `project`.

**View:**
- *Terminal:* Go to `https://beneath.dev/USERNAME`
- *CLI*: (Not available)

## `Organization`s

**Definition:** An `organization` is the top-level owner of users, services and projects. Billing is managed at the `organization` level, which means that every resource that can accrue bills is directly or indirectly linked to exactly one `organization`. 

**Relations:**
- An `organization` has many (one or more) `user`s. Since all `user`s must belong to an `organization`, when a `user` is created, they're added to a "personal" `organization` that is automatically and transparently created with the same name as the username.
- An `organization` has many (zero or more) `project`s.
- An `organization` has many (zero or more) `service`s.

**Access management:**
- A `user` can access an `organization`.
  - The `view` permission grants the `user` permission to browse the members and projects in the organization.
  - The `admin` permission grants the `user` permission to add and delete members, create new projects, and change billing information.

**View:**
- *Terminal:* Go to `https://beneath.dev/ORGANIZATION_NAME`
- *CLI:* Run `beneath organization --help` for details.

## `Project`s

**Definition:** A `project` is a collection of `stream`s and `model`s. You can think of them like repositories in Git.

**Relations:**
- A `project` belongs to one `organization`.
- A `project` has many (zero or more) `stream`s.
- A `project` has many (zero or more) `model`s.

**Access management:**
- A `user` can access a `project`.
  - The `view` permission grants the `user` permission to browse the contents of the project, including viewing and querying records in its `stream`s.
  - The `create` permission grants the `user` permission to create and delete `stream`s and `model`s in the project, including writing data directly to root `stream`s.
  - The `admin` permission grants the `user` permission to add, remove and change permissions for other `user`'s.

**View:**
- *Terminal:* Go to `https://beneath.dev/ORGANIZATION_NAME/PROJECT_NAME`
- *CLI:* Run `beneath project --help` for details.

## `Stream`s

**Definition:** 

**Relations:**
- 

**Access management:**
- 

**View:**
- 

## `StreamInstance`s

**Definition:** 

**Relations:**
- 

**Access management:**
- 

**View:**
- 

## `Model`s

**Definition:** 

**Relations:**
- 

**Access management:**
- 

**View:**
- 


## `Service`s

**Definition:** 

**Relations:**
- 

**Access management:**
- 

**View:**
- 


## `Secret`s

**Definition:** 

**Relations:**
- 

**Access management:**
- 

**View:**
- 

