---
title: Read data into Jupyter notebook
description:
menu:
  docs:
    parent: quick-starts
    weight: 100
weight: 100
---

Time required: 5 minutes.

In this quick-start, we read an event stream from Beneath into a Jupyter notebook. You can read any public data stream or any of the private data streams that you have access to.

## Install the Beneath Python SDK
Install the Python SDK from your command line:
```bash
pip install beneath
```

## Log in to the Terminal
Go to the [Terminal](https://beneath.dev/?noredirect=1), and log in. If you don't yet have an account, create one.

## Create a Read-Only secret

- Navigate to your user profile by clicking on the profile icon in the top right-hand corner of the screen.
- Click on the Secrets tab
- Click "Create new read-only secret" and enter a description
- Save your secret!

## Authorize your local environment
From your command line:
```bash
beneath auth SECRET
``` 
Now your secret has been stored in a hidden folder, `.beneath`, in your home directory

## Navigate to a data stream's API tab

- The Beneath directory structure is USER/PROJECT/STREAM
- In the [Terminal](https://beneath.dev/?noredirect=1), navigate to your desired stream, and click on the API tab

## Copy-paste the Python snippet into a Jupyter notebook
Many data workers choose [Jupyter notebooks](https://jupyter.org/) for ad-hoc analyses. A few short lines of Python code are all you need to import Beneath data into your notebook environment.

Here's the template for Python imports, but on the API tab, the stream's path is automatically populated for you.
 
```python
from beneath import Client
client = Client()
df = await client.easy_read(USER/PROJECT/STREAM)
```

And there's your data in a Pandas DataFrame!
