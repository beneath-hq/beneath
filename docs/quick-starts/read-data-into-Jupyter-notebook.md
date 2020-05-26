---
title: Read data into Jupyter notebook
description: A guide to reading your data into an interactive Python environment
menu:
  docs:
    parent: quick-starts
    weight: 200
weight: 200
---

Time required: 5 minutes.

In this quick-start, we read a data stream from Beneath into a [Jupyter notebook](https://jupyter.org/). Jupyter notebooks provide an interactive environment in Python (as well as many other languages) that many data analysts like to use to work with data.

## Install the Beneath Python SDK

Install the Python SDK from your command line:
```bash
pip install --upgrade beneath
```

## Log in to the Terminal

Go to the [Terminal](https://beneath.dev/?noredirect=1), and log in. If you don't yet have an account, create one.

## Create a command-line secret

- Navigate to your user profile by clicking on the profile icon in the top right-hand corner of the screen.
- Click on the Secrets tab
- Click "Create new command-line secret"
- Copy and save your secret!

## Authorize your local environment

From your command line:
```bash
beneath auth SECRET
``` 

Now your secret has been stored in a hidden folder, `.beneath`, in your home directory. When you use Beneath from your local machine, it will automatically authenticate with this secret.

## Navigate to a data stream's API tab

You can read any public data stream (for example, check out the [featured projects](https://beneath.dev/?noredirect=1)) or any of the private data streams that you have access to. 

The Beneath directory structure is USER/PROJECT/STREAM.

In the [Terminal](https://beneath.dev/?noredirect=1), navigate to your desired stream, and click on the API tab

## Open a Jupyter notebook and paste the snippet

If you don't have Jupyter installed, here's a super quick [installation guide](https://jupyter.org/install.html) to get up-and-running. 

From your command line, it's simple to launch a notebook:
```bash
jupyter notebook
```

In the notebook, you can run code interactively. Enter your code in a cell and run it with `shift` + `enter`. Now you're ready to load some data from Beneath. 

## Copy the Python snippet

A few short lines of Python are all you need to import Beneath data into your notebook environment. Run the following snippet to get a [pandas](https://pandas.pydata.org/) dataframe of COVID-19 data.
 
```python
from beneath import Client
client = Client()
df = await client.easy_read("bem/covid19/cases", to_dataframe=True)
```

And there's your data, all ready for analysis!
