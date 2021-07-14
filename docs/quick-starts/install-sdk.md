---
title: Install the Beneath SDK
description: A guide to connecting from your local machine
menu:
  docs:
    parent: quick-starts
    weight: 200
weight: 200
---

Install and authenticate with Beneath from your local machine to use the CLI (command-line interface) or to use Beneath in Python.

## Install with pip

Run the following command at the terminal to install Beneath. You must have Python 3.7 or greater installed.

```bash
pip3 install --upgrade beneath
```

## Login and authenticate your local environment

Run the following command to authenticate your local environment. If you do not have a Beneath account, it will automatically guide you to create one:

```bash
beneath auth
```

Now, when you use Beneath on your local machine (such as the CLI or in Python), it will automatically authenticate your requests. To access Beneath outside of your local environment (e.g. when deploying code to production), see [Access management]({{< ref "/docs/reading-writing-data/access-management.md" >}}) for details.

## Docs for the CLI

To get an overview of the CLI, run:

```bash
beneath --help
```

To learn more about a specific command, run:

```bash
beneath SUBCOMMAND --help
```

## Docs for the Python client

We recommend you check out the next quick start to get started with the Beneath client for Python, but if you want the full details about available classes and functions, you can also check out the [Python client API reference](https://python.docs.beneath.dev/).
