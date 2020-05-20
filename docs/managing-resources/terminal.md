---
title: Terminal
description: A guide to the Beneath Terminal, which is Beneath's graphical user interface
menu:
  docs:
    parent: managing-resources
    weight: 100
weight: 100
---

We have often found that it can be frustrating and difficult to maintain and communicate an overview of complex data systems. Beneath changes this with the *Terminal*, a web-based UI for Beneath. It's designed to make it easy to:

- Discover and explore data on Beneath
- Monitor the flow and health of streams
- Understand which streams that are derived or transformed from other streams
- Share data safely with less experienced users (who might not have a background in data science)

The Terminal is currently focused on displaying streams, integrations and resources. To create new streams or other resources, use the [Beneath CLI]({{< ref "/docs/managing-resources/cli" >}}).

## Browsing the Terminal

You access the Beneath Terminal by visiting [https://beneath.dev](https://beneath.dev?noredirect=1) or clicking the "Terminal" button on the top-right corner of this website. You don't have to be logged in to use most parts of the Terminal.

Accessing the Terminal will take you to the *springboard*, the entry point to the terminal. If you are not logged in, it will show you interesting data to explore. If you are logged in, you will see an overview of your usage and your projects.

Resources in the Terminal follow the `/organization/project/stream` hierarchy described in [Resources]({{< ref "/docs/managing-resources/resources" >}}). Note that if you're not part of an *organization*, your *username* is your organization name.

## Managing public information

User and organization profiles are public on Beneath (similar to Github). You can customize your name, biography and profile picture on your profile page, which is found at `beneath.dev/<USERNAME OR ORGANIZATION NAME>`.

## Issuing secrets

To log into the Beneath CLI or to access streams from your code, you need a *secret*. You can issue new secrets from the "Secrets" tab of your user or organization page. You can learn more about secrets in the sections [Resources]({{< ref "/docs/managing-resources/resources" >}}) and [Access management]({{< ref "/docs/managing-resources/access-management" >}}).

## Monitoring usage

Use the "Monitoring" tab on the page of any user, organization, service or stream to view usage statistics. This is useful to monitor the health of your streams and to understand costs. [Click here]({{< ref "/docs/billing/usage" >}}) to read more about the usage statistics Beneath tracks. 

## Exploring streams

You can browse the data in a stream in the Beneath Terminal. It provides different representations of the data, mirroring the [different underlying data systems]({{< ref "/docs/overview/unified-data-system" >}}) where the data is stored. In short:

- **Log** representation lets you browse records in the order they were written. You can sort records by "oldest" or "newest".
- **Index** representation lets you browse records sorted by their [key]({{< ref "/docs/reading-writing-data" >}}). It only contains the newest record for each key (unlike the log, which contains every version). You can provide a filter to quickly lookup records by their key. See the [Reading and Writing Data]({{< ref "/docs/reading-writing-data" >}}) section for details on the filter query syntax.

The Terminal automatically subscribes to real-time updates when you view a stream, so the data will update within milliseconds in response to new writes.
