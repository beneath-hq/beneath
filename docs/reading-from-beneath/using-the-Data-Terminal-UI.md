---
title: Using the Data Terminal UI
description:
menu:
  docs:
    parent: reading-from-beneath
    weight: 100
weight: 100
---
The Explore tab provides a quick, in-browser way to explore the data on Beneath. If you’d like to look at stream schema, validate data entry, or filter for a subset of data, this tab is a great place to start.

##### Filters

Here are example queries for the cryptocurrency token transfers table.

**Filter for a single key** 

```json
{ "symbol": "BTC" }
```

**Filter for a compound key**

```json
{ "symbol": "BTC", "day": "2019-07-20" }
```


**Filter for a compound key with constraints** 

```json
{ "symbol": "BTC", "day": { "_gte": "2019-07-01", "_lt": "2019-08-01" } }
```

“_gte” means “greater than or equal to” <br>
“_lt” means “less than”

**Filter for multiple keys**

```json
{ "symbol": { "_prefix": "B" } }
```

Returns all prices for all tokens starting with "B"

**Filter for multiple compound keys with constraints**

```json
{ "symbol": { "_prefix": "B" }, "day": { "_gte": "2019-07-01", "_lt": "2019-08-01" } }
```

Returns an error! See "lexicographic order."

##### Primary keys

It is important to note that you are only able to filter data based on the stream’s key(s). You will not be able to filter on non-key columns.

##### Lexicographic order

It is important to note that the keys are lexicographically sorted. You are only able to filter for data that is stored in one adjacent block. The last query above fails because the query attempts to extract two non-adjacent blocks of data.
