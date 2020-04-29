---
title: Command-line interface (CLI)
description: A guide to the Beneath CLI, which is the command-line interface to Beneath
menu:
  docs:
    parent: managing-resources
    weight: 200
weight: 200
---

The Beneath CLI allows you to interact with Beneath directly from the command-line. While the Beneath Terminal is focused on providing a great overview of resources, the Beneath CLI is especially focused on creating and manipulating resources on Beneath. It's your go-to tool for things like creating projects, staging new streams, changing access permissions, creating a service, issuing service secrets, and more.

## Installation

You need to have Python 3 installed on your computer to install the Beneath CLI. Then simply install it with:

```
pip3 install beneath
```

You can verify that the Beneath CLI installed successfully by running:

```
beneath --version
```

Most features in the CLI require you to be authenticated. To authenticate, issue a "command-line secret" in the ["Secrets" tab of your profile page](https://beneath.dev/-/redirects/secrets), then run:

```
beneath auth COMMAND_LINE_SECRET
```

The secret is stored in a hidden folder, `.beneath`, in your home directory. The CLI and most client libraries will automatically load and use it on your computer if you don't explicitly provide another secret.

## Documentation

You can get help and reference documentation for the Beneath CLI by running the command:

```
beneath --help
```

The `--help` parameter also works for subcommands, for example:

```
beneath project create --help
```

In addition, many documentation pages contain examples of how to use the Beneath CLI (such as [Quick starts]({{< ref "/docs/quick-starts" >}})).
