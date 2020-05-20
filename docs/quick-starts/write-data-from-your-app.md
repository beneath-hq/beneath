---
title: Write data from your app
description:
menu:
  docs:
    parent: quick-starts
    weight: 300
weight: 300
---

Time required: 5 minutes.

In this quick-start, we write a data stream to Beneath. The stream might originate from your application or from some external API that you would like to consume.

You can follow this quick start, or check out [this Jupyter notebook](https://gitlab.com/beneath-hq/beneath/-/blob/master/clients/python/examples/notebooks/covid19.ipynb) that scrapes COVID-19 data and writes it to Beneath.

## Install the Beneath Python SDK
If you haven't already, install the Python SDK from your command line:
```bash
pip install beneath
```

## Log in to the Beneath Terminal
Go to the [Terminal](https://beneath.dev/?noredirect=1), and log in. If you don't yet have an account, create one.

## Create a Command-Line secret

- Navigate to your user profile by clicking on the profile icon in the top right-hand corner of the screen.
- Click on the Secrets tab
- Click "Create new command-line secret" and enter a description
- Save your secret!

## Authorize your local environment
From your command line:
```bash
beneath auth SECRET
``` 
Now your secret has been stored in a hidden folder, `.beneath`, in your home directory

## Create a Beneath project
On Beneath, every data stream lives in a project. Like on GitHub, every code file lives in a repository.

Additionally, every project is assigned to an organization. This ensures that all project resources have someone assigned for billing purposes.

**When you sign up for Beneath, you are automatically assigned an organization with the same name as your username.** The organization is automatically assigned to a Free billing plan.

On your command line, provide your username and the name you'd like for your project:
```bash
beneath project create USER/PROJECT
```

## Initialize the Python client and stage your stream
Either in your application code, or, to test, in a Jupyter notebook, connect to your stream with the following Python code. Use your command-line secret.
providing your username, project name, and the name you'd like for your stream

Every stream requires a schema. If needed, there is more information about the Beneath schema language [here](/docs/schema-language).

```python
import beneath
client = beneath.Client()
stream = await client.stage_stream(USER/PROJECT/STREAM, """
type SchemaTemplate @stream() @key(fields: ["my_integer_index"]) {
  "This is also the stream's index, as the key defines above."
  my_integer_index: Int!

  my_timestamp: Timestamp!
  my_bytes: Bytes32!
  my_big_number: Numeric!

  "This field is not required, as denoted by the lack of exclamation point."
  my_optional_string: String
} 
""")
```

Great! You just created a Beneath stream. The next step is to write data to the stream.

## Generate your data
Generate or capture the data of interest and ensure it aligns with your stream's schema that you defined above. If the data does not conform to the defined schema, Beneath will reject the write request.

```python
from datetime import datetime
import secrets
import sys

def generate_record(n):
  return {
    "my_integer_index": n,
    "my_timestamp": datetime.now(),
    "my_bytes": secrets.token_bytes(4),
    "my_big_number": secrets.randbelow(sys.maxsize),
    "my_optional_string": None
  }

n = 25
records = [generate_record(n) for _ in range(n)]
```

## Write your data to Beneath
The write function accepts a list of data records. Here, we write a list of size 1. We use the await syntax so that we do not block our application while we wait for Beneath's response.

```python
await stream.write(records)
```

## Check out the Beneath Terminal to see data arrive in realtime 
Navigate to your stream in the Terminal and go to the Explore tab to see your data arrive in realtime. The url will look like: https://beneath.dev/USER/PROJECT/STREAM

## Further examples
More examples of the Python SDK in action can be found [here](https://gitlab.com/beneath-hq/beneath/-/tree/master/clients/python/examples).
