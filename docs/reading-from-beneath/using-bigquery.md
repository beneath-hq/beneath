---
title: Using BigQuery
description:
menu:
  docs:
    parent: reading-from-beneath
    weight: 300
weight: 300
---
All data streams on Beneath are stored in BigQuery and are accessible via BigQuery’s web console or via BigQuery’s easy-to-use API.

##### BigQuery web console

Find BigQuery’s console here: <https://bigquery.cloud.google.com/>

The BigQuery console is a great playground for querying data with SQL. Here’s are a couple examples for accessing Beneath data:

```sql
SELECT * FROM `beneath.ethereum.blocks`
```

```sql
SELECT extract(date from _time) as day, count(distinct _from) as active_addresses

FROM `beneath.ethereum.dai_transfers`

Group By 1

Order By 1
```

##### BigQuery API

Here is a basic example using BigQuery’s Python client library:

```python
from google.cloud import bigquery

client = bigquery.Client()

QUERY = ('SELECT * FROM `beneathcrypto.ethereum.blocks`')

query_job = client.query(QUERY)  # API request

rows = query_job.result()  # Waits for query to finish

for row in rows:
    print(row)
```
For more information, you can check out Google’s BigQuery documentation for Python [here](https://googleapis.github.io/google-cloud-python/latest/bigquery/index.html) or for other client libraries [here](https://cloud.google.com/bigquery/docs/quickstarts/quickstart-client-libraries).
