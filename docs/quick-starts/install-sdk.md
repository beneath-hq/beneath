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

## Create a command-line secret

To authenticate with Beneath from the command-line, you need a command-line secret:

1. Go to the web [Console](https://beneath.dev/?noredirect=1), and log in or create an account
2. Navigate to the "Secrets" page by clicking on your profile icon in the top right-hand corner of the screen
3. Under "Create personal secret", enter a description and create a secret with "Full (CLI)" access
4. Highlight and copy your secret from the green box

## Authenticate your local environment

Run the following command, replacing `SECRET` with the secret you just obtained from the web console:
```bash
beneath auth SECRET
``` 

Now, when you use Beneath on your local machine (such as the CLI or in Python), it will automatically authenticate with this secret.

## Docs for the CLI

To get an overview of the CLI, run:
```bash
beneath --help
```

To learn more about a specific command, run:
```bash
beneath SUBCOMMAND --help
```
