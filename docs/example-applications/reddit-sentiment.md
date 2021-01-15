---
title: Reddit sentiment
description: Use Beneath, Kubernetes, Next.js, and Netlify to build and deploy an analytics website
menu:
  docs:
    parent: example-applications
    weight: 200
weight: 200
---

IN PROGRESS.

In this walk-through, we'll create a real-time, data-driven web app to monitor sentiment of Reddit posts. The following things are needed for dashboards like this one:

- Subscribe to a data feed and store it
- Apply an NLP model for sentiment analysis
- Visualize the data in real-time
- Build and deploy a web app so we can share our work

To accomplish these tasks, our full stack includes:

- [Beneath](about.beneath.dev), our data system.
- [Kubernetes](), a to orchestrate the compute that subscribes to Reddit and writes data to Beneath.
- [Next.js](https://nextjs.org/), a minimalistic React framework. The latest and greatest way to build Single Page Applications.
- [Recharts.js](http://recharts.org/en-US/), a React-native graphing library.
- [MaterialUI](https://material-ui.com/), a library for design aesthetics.
- [Netlify](https://www.netlify.com/), a tool to build and deploy our web app.

Given our objectives and stack, why is Beneath a good choice for data storage?

- Little code!
- Store the data in a data warehouse to enable efficient SQL queries
- Store the data in a log to quickly fetch the latest records.
- Very easy React library to fetch data into our frontend
- No need to build an API server if we want to share data with others
- No need to configure or maintain a database(s)

## What we'll build

- First, we'll create this [stream of real-time Reddit posts](https://beneath.dev/epg/reddit/stream:coronavirus-posts-sentiment).
- Then we'll create this [website](https://nextjs.demo.beneath.dev).

For reference, here's all [the frontend code](https://gitlab.com/beneath-hq/beneath/-/tree/master/clients/js-react/examples/reddit-sentiment) and [the backend code](https://gitlab.com/beneath-hq/beneath/-/tree/master/clients/python/examples/reddit-sentiment).

## Prerequisites

In the interest of getting-to-the-point and focusing on Beneath, we'll assume you can set yourself up with the following:

- A [Kubernetes](https://kubernetes.io/docs/home/) cluster. There are [many ways to set up Kubernetes](https://kubernetes.io/docs/setup/). Personally, we use Google's managed Kubernetes offering, [GKE](https://cloud.google.com/kubernetes-engine).
- [Docker Desktop](https://docs.docker.com/desktop/). So you can build and deploy your container images.
- A [Beneath account](https://beneath.dev?noredirect=1). Also [install the SDK](/docs/quick-starts/install-the-sdk/) to follow along.
- A [Reddit Developer account](). Get yourself an API key in order to fetch loans from their servers.
- A [Netlify account]()

Alright, let's start!

## Write real-time data to Beneath

## Initialize a Nextjs + MaterialUI site

### Run `create next-app`

From the [Next.js Getting Started docs](https://nextjs.org/docs/getting-started), in our command line:

```bash
yarn create next-app
```

### Add files for MaterialUI

Next.js provides a fantastic repository of examples. Here's their [Next.js example for MaterialUI](https://github.com/mui-org/material-ui/tree/master/examples/nextjs).

We add the MaterialUI package:

```bash
yarn add @material-ui/core
```

Copied files into my project directory:

- pages/\_app.js
- pages/\_document.js
- src/theme.js

Now we're all ready to use React components and Material UI.

## Read data from Beneath into frontend

### Add the Beneath dependency

From your command line:

```bash
yarn add beneath-react
```

### Read data into your frontend

```javascript
import { useRecords } from "beneath-react";

const {
  records,
  error,
  loading,
  fetchMore,
  fetchMoreChanges,
  subscription,
  truncation,
} = useRecords({
  stream: "epg/reddit/coronavirus-posts-sentiment",
  query: { type: "log", peek: "true" },
  pageSize: 500,
});
```

## Use Recharts to display the data

Recharts has great documentatation. And it's very intuitive to use for anyone with React component experience.

Here's our full code:

```javascript
<ResponsiveContainer width={"100%"} height={400}>
  <LineChart
    data={sortedRecords}
    margin={{ top: 30, right: 30, left: 30, bottom: 30 }}
  >
    <CartesianGrid />
    <XAxis
      type="number"
      dataKey="timestamp"
      domain={["dataMin", "dataMax"]}
      tickFormatter={(unixTime) =>
        moment(unixTime).format("MMMM Do YYYY, h:mm:ss a")
      }
      interval="preserveStartEnd"
    >
      <Label
        value={"Date"}
        position="bottom"
        style={{ textAnchor: "middle" }}
      />
    </XAxis>
    <YAxis dataKey="polarity" domain={[-1, 1]}>
      <Label
        value={"Polarity"}
        position="left"
        angle={-90}
        style={{ textAnchor: "middle" }}
      />
    </YAxis>
    <Tooltip
      formatter={(value) => value.toFixed(2)}
      labelFormatter={(unixTime) =>
        moment(unixTime).format("MMMM Do YYYY, h:mm:ss a")
      }
    />
    <Line dataKey="polarity" stroke="#8884d8" />
    <ReferenceLine y={0} stroke="black" />
  </LineChart>
</ResponsiveContainer>
```

Notes for working with Rechart:

- this shows how to customize spacing of xaxis ticks and how to customize your domain (the range of your xaxis) https://github.com/recharts/recharts/issues/1137
- https://medium.com/swlh/data-visualisation-in-react-part-i-an-introduction-to-recharts-33249e504f50

## Build and deploy site with Netlify
