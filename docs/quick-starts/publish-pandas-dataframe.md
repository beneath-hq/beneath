---
title: Publish a Pandas DataFrame
description: A guide to loading a Pandas DataFrame into Beneath and sharing it with others
menu:
  docs:
    parent: quick-starts
    weight: 800
weight: 800
---

This quick start will show you how to publish a [Pandas DataFrame](https://pandas.pydata.org/pandas-docs/stable/reference/api/pandas.DataFrame.html) on Beneath. Share the DataFrame with others, or use it as reference data in a Beneath data pipeline or SQL query.

## Install the Beneath SDK

If you haven't already, follow the [Install the Beneath SDK]({{< ref "/docs/quick-starts/install-sdk" >}}) quick start to install and authenticate Beneath on your local computer.

## Create a Pandas DataFrame

In this example, we'll share a csv file of S&P 500 stocks that we've saved locally. Here we create a DataFrame from the csv:

```python
import pandas as pd
df = pd.read_csv("data/s-and-p-500-constituents.csv")
```

## Create a Beneath project

From the command line, we create a project named `financial-reference-data`:

```bash
beneath project create USERNAME/financial-reference-data
```

## Load the DataFrame into Beneath

Use the Beneath Python SDK to write the data into Beneath. We can optionally provide a key to [index the data]({{< ref "/docs/reading-writing-data/index-filters" >}}) and a little description:

```python
import beneath
await beneath.write_full(
    table_path="USERNAME/financial-reference-data/s-and-p-500-constituents",
    records=df,
    key=["symbol"],
    description="The full list of companies in the S&P 500 (updated March 31st, 2021)"
)
```

After loading the data, go take a look at the data in the web console: https://beneath.dev/examples/financial-reference-data/table:s-and-p-500-constituents

## Give your teammates access

Share the project with your friends:

```bash
beneath project update-permissions USERNAME/PROJECT_NAME FRIEND_USERNAME --view --create
```

Or make your project completely public:

```bash
beneath project update USERNAME/PROJECT_NAME --public
```
