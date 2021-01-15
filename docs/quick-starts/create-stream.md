---
title: Create stream and write data
description: A guide to creating a stream and writing data to it
menu:
  docs:
    parent: quick-starts
    weight: 300
weight: 300
---

In this quick-start, we create a stream in Beneath and write some data to it using Python.

## Install the Beneath SDK

If you haven't already, follow the [Install the Beneath SDK]({{< ref "/docs/quick-starts/install-sdk" >}}) quick start to install and authenticate Beneath on your local computer.

## Create a Beneath project

On Beneath, every data stream lives in a project. You can create many projects, and a project can contain many streams. You can think of projects as Git repositories.

To create a project from the command-line, run the following command:
```bash
beneath project create USERNAME/NEW_PROJECT_NAME
```
Replace `USERNAME` with your Beneath username or organization, and `NEW_PROJECT_NAME` with your project name.

If you prefer, you can also create a project straight from the web console using the [Create project](https://beneath.dev/-/create/project) page.

## Create a stream

In this example, we will create a stream of movies using Python. We suggest you run the code in a Jupyter notebook, but you can use any Python environment you prefer.
```python
import beneath
client = beneath.Client()

stream = await client.create_stream("USERNAME/PROJECT_NAME/movies", schema="""
    " A stream of movies "
    type Movie @schema {
        title: String! @key
        released_on: Timestamp! @key 
        director: String!
        description: String
    }
""", update_if_exists=True)
```
To run the above example, you need to:
1. Replace the stream path with your username and project name
- Run the code in a context that supports `await` (`asyncio`). Notebooks support these by default, but if you're using plain Python, you need to wrap it in an [async function](https://docs.python.org/3/library/asyncio-task.html#coroutines).
- Optionally adapt the schema to suit your use case

In Beneath, every stream has a *schema* that defines its fields and their data types, and a *key* consisting of one or more fields that uniquely identify a record. Schemas are based on GraphQL, and you can read more about them [here]({{< ref "/docs/reading-writing-data/schema-definition" >}}).

Great! You just created a stream on Beneath.

## Write data

Now, let's write some data to the stream.

First, navigate to your stream in the Beneath web [console](https://beneath.dev), where the URL will look like `beneath.dev/USERNAME/PROJECT/stream:movies`. Keep this window open, so you can see the records flow as you write data.

Then run the following Python code, which writes a single record to the stream:
```python
from datetime import datetime

async with stream.primary_instance.writer() as w:
    await w.write({
        "title":       "Star Wars: Episode IV",
        "released_on": datetime(year=1977, month=5, day=25),
        "director":    "George Lucas",
        "budget_usd":  11000000,
        "rating":      8.6,
    })
```

Did you see it tick in? Try tweaking and running the code several times to write multiple records.

## Writing more data

Let's write some more data to our stream. We'll fetch a dataset of movies from Github and write some random movies to our stream. 

Here's the Python code:

```python
import aiohttp
import asyncio
from datetime import datetime
import random

# load a dataset of movies
url = "https://raw.githubusercontent.com/vega/vega/master/docs/data/movies.json"
async with aiohttp.ClientSession() as session:
    async with session.get(url) as res:
        movies = await res.json(content_type=None)

# write a 100 random movies one-by-one
n = 100
async with stream.primary_instance.writer() as w:
    for i in range(n):
        # get a random movie
        movie = random.choice(movies) 

        # transform and write the movie
        await w.write({
            "title":       str(movie["Title"]),
            "released_on": datetime.strptime(movie["Release Date"], "%b %d %Y"),
            "director":    movie["Director"],
            "budget_usd":  movie["Production Budget"],
            "rating":      movie["IMDB Rating"],
        })

        # wait about a second
        await asyncio.sleep(1.1)
```

## Clean up

If you'd like to clean up, you can delete the stream and project you created in this tutorial.

To delete the movies stream from the command-line:
```bash
beneath stream delete USERNAME/PROJECT_NAME/movies
```

You can also delete the empty project:
```bash
beneath project delete USERNAME/PROJECT_NAME
```

## Further examples

More examples of the Python SDK in action can be found in the [examples repo](https://gitlab.com/beneath-hq/beneath/-/tree/master/examples).
