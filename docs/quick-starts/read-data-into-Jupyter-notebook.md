---
title: Read data into Jupyter notebook
description: A guide to reading your data into an interactive Python environment
menu:
  docs:
    parent: quick-starts
    weight: 500
weight: 500
---

Time required: 5 minutes.

In this quick-start, we read a data stream from Beneath into a [Jupyter notebook](https://jupyter.org/). Jupyter notebooks provide an interactive environment in Python (as well as many other languages) that many data analysts like to use to work with data.

## Install the Beneath Python SDK

Install the latest version of the Python SDK from your command line:
```bash
pip3 install --upgrade beneath
```

## Log in to the Console

Go to the [Console](https://beneath.dev/?noredirect=1), and log in. If you don't yet have an account, create one.

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

In the [Console](https://beneath.dev/?noredirect=1), navigate to your desired stream, and click on the API tab

## Open a Jupyter notebook

If you don't have Jupyter installed, here's a super quick [installation guide](https://jupyter.org/install.html) to get up-and-running. 

From your command line, it's simple to launch a notebook:
```bash
jupyter notebook
```

In the notebook, you can run code interactively. Enter your code in a cell and run it with `shift` + `enter`. Now you're ready to load some data from Beneath. 

## Copy and paste the Python snippet

A few short lines of Python are all you need to import Beneath data into your notebook environment. On the "API" tab of every stream on Beneath, you'll find a Python snippet you can copy and paste to load records into your code.

Here's a snippet taken from the [COVID-19 cases](https://beneath.dev/bem/covid19/cases) stream. Run it to load the data as a [pandas](https://pandas.pydata.org/) dataframe:
 
```python
from beneath import Client
client = Client()
df = await client.easy_read("bem/covid19/cases", to_dataframe=True)
```

And there's your data, all ready for analysis!
