---
title: Examples
description: Tutorials and example projects on Beneath
menu:
  docs:
    parent: misc
    weight: 150
weight: 150
---

Beneath runs several public example projects on [beneath.dev/examples](https://beneath.dev/examples).

If you want to use the data, you can simply start consuming the streams directly from [beneath.dev/examples](https://beneath.dev/examples). We plan on keeping these examples running forever.

If you want to use the examples as a starting point for creating your own streams, you can find the source code for all these projects in the `examples/` directory in the [Beneath repo on Github](https://github.com/beneath-hq/beneath/tree/master/examples). Most of the examples contain a `README` that shows how to run the code.

## Examples list

- **`examples/earthquakes`:** Scrapes worldwide earthquakes data in real-time from the U.S. Geological Survey and writes it to Beneath.
  - Main stream: [examples/earthquakes/stream:earthquakes](https://beneath.dev/examples/earthquakes/stream:earthquakes)
  - Project: [examples/earthquakes](https://beneath.dev/examples/earthquakes)
  - [Source code](https://github.com/beneath-hq/beneath/tree/master/examples/earthquakes)
- **`examples/reddit`:** Scrapes posts and comments from a few subreddits. The code is a generic scraper that can extract posts and comments from any subreddit to Beneath.
  - Example stream: [examples/reddit/stream:r-wallstreetbets-comments](https://beneath.dev/examples/reddit/stream:r-wallstreetbets-comments)
  - Project: [examples/reddit](https://beneath.dev/examples/reddit)
  - [Source code](https://github.com/beneath-hq/beneath/tree/master/examples/reddit)
- **`examples/clock`:** Publishes several streams that tick at fixed intervals.
  - Example stream: [examples/clock/stream:clock-1m](https://beneath.dev/examples/clock/stream:clock-1m)
  - Project: [examples/clock](https://beneath.dev/examples/clock)
  - [Source code](https://github.com/beneath-hq/beneath/tree/master/examples/clock)
- **`examples/ethereum`:** Loads blocks from the Ethereum blockchain in real-time.
  - Main stream: [examples/ethereum/stream:blocks-stable](https://beneath.dev/examples/ethereum/stream:blocks-stable)
  - Project: [examples/ethereum](https://beneath.dev/examples/ethereum)
  - [Source code](https://github.com/beneath-hq/beneath/tree/master/examples/ethereum)
