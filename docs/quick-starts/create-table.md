---
title: Create a table and write data
description: A guide to creating a real-time table and writing data to it
menu:
  docs:
    parent: quick-starts
    weight: 300
weight: 300
---

In this quick-start, we create a table in Beneath and write some data to it using Python.

## Install the Beneath SDK

If you haven't already, follow the [Install the Beneath SDK]({{< ref "/docs/quick-starts/install-sdk" >}}) quick start to install and authenticate Beneath on your local computer.

## Create a Beneath project

On Beneath, every table lives in a project. You can create many projects, and a project can contain many tables. You can think of projects as Git repositories.

To create a project from the command-line, run the following command:

```bash
beneath project create USERNAME/NEW_PROJECT_NAME
```

Replace `USERNAME` with your Beneath username or organization, and `NEW_PROJECT_NAME` with your project name.

If you prefer, you can also create a project straight from the web console using the [Create project](https://beneath.dev/-/create/project) page.

## Create a real-time table

In this example, we will create a table of movies using Python. If you prefer, you can also create a table using the CLI or from the web console using [Create table](https://beneath.dev/-/create/table).

In Beneath, every table has a _schema_ that defines its fields and their data types, and a _key_ consisting of one or more fields that uniquely identify a record. Schemas are based on GraphQL, and you can read more about them [here]({{< ref "/docs/reading-writing-data/schema-definition" >}}).

We suggest you run the code in a Jupyter notebook, but you can use any Python environment you prefer:

```python
import beneath
client = beneath.Client()
await client.start()

table = await client.create_table("USERNAME/PROJECT_NAME/movies", schema="""
    " A table of movies "
    type Movie @schema {
        title: String! @key
        released_on: Timestamp! @key
        director: String
        budget_usd: Int
        rating: Float
    }
""", update_if_exists=True)

# Stop the client at the end of your script
# await client.stop()
```

To run the above example, you need to:

1. Replace the table path with your username and project name
2. Run the code in a context that supports `await` (`asyncio`). Notebooks support these by default, but if you're using plain Python, you need to wrap it in an [async function](https://docs.python.org/3/library/asyncio-task.html#coroutines).
3. Optionally adapt the schema to suit your use case

That's it! You just created a table on Beneath.

## Write data

Now, let's write some data to the table.

First, navigate to your table in the Beneath web [console](https://beneath.dev), where the URL will look like `beneath.dev/USERNAME/PROJECT/table:movies`. Keep this window open, so you can see the records flow as you write data.

Then run the following Python code, which writes a single record to the table:

```python
from datetime import datetime
await table.write({
    "title":       "Star Wars: Episode IV",
    "released_on": datetime(year=1977, month=5, day=25),
    "director":    "George Lucas",
    "budget_usd":  11000000,
    "rating":      8.6,
})
```

Did you see it tick in? Try tweaking and running the code several times to write multiple records.

## Writing more data

Let's write some more data to our table. We'll fetch a dataset of movies from Github and write some random movies to our table.

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
for i in range(n):
    # get a random movie
    movie = random.choice(movies)

    # transform and write the movie
    await table.write({
        "title":       str(movie["Title"]),
        "released_on": datetime.strptime(movie["Release Date"], "%b %d %Y"),
        "director":    movie["Director"],
        "budget_usd":  movie["Production Budget"],
        "rating":      movie["IMDB Rating"],
    })

    # wait a while
    await asyncio.sleep(5)
```

## Clean up

If you'd like to clean up, you can delete the table and project you created in this tutorial. To delete the movies table from the command-line:

```bash
beneath table delete USERNAME/PROJECT_NAME/movies
```

Then you can delete the empty project:

```bash
beneath project delete USERNAME/PROJECT_NAME
```

## More info

Under the "API" tab on the table's page in the web console, you will find several more guides for reading and writing to the table.

Other examples of the Python SDK in action can be found in the [examples repo](https://github.com/beneath-hq/beneath/tree/master/examples). For full details about available classes and functions, check out the [Python client API reference](https://python.docs.beneath.dev/).
