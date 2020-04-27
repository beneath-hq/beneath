---
title: Replay and derive streams
description: How Beneath uses event logs to make it easy to derive streams and synchronize systems
menu:
  docs:
    parent: overview
    weight: 400
weight: 400
---

In Beneath, there's an *event log* at the heart of every data stream. An event log keeps real-time, ordered track of every record written to a stream, allowing you to read new data starting from any point in time *without missing a single change*. 

In practice, it means you can replay the history of a stream as if you had been subscribed since its beginning, then stay subscribed to get every new change within milliseconds of it happening. And if you're unsubscribed for a while, you can get every change that happened during your downtime once you resubscribe (it's an *at-least-once guarantee*).

With event logs, it becomes really easy to create transformations that derive from one stream and produce a new one. Here we give just a few examples that illustrate the virtually endless opportunities that affords.

## Examples

### Applying a machine learning model to uploaded data

Let's say you've created a machine learning model that can tell if an image has been manipulated. Now you want to turn it into a production service where users can upload their own images. Every time a user uploads an image, you write it to a stream called `images`. Separately, you run a small service that uses the `images` event log to get every new image, run the model on it to determine if the image is has been manipulated, then writes the result to a new stream `labeled-images`. Since Beneath automatically [provides indexed lookups]({{< ref "/docs/overview/unified-data-system" >}}), your UI can just ping the `labeled-images` index for the result, which will likely appear within milliseconds. 

This workflow also makes it easy to improve your machine learning model and update your results. When you've trained an improved model, just update your service and reset its subscription to replay every past image into the model!

This may all sound a little complicated, but Beneath really handles most of it for you. Your only responsibility is to write images into Beneath and keep your model running (and even if it breaks, the event log ensures you don't forget to label any images uploaded while its down!).

### Filtering data in real-time based on criteria

If, for example, you're writing live IoT sensor data into a stream in Beneath, you can run a transformation that reads every observation, checks if it exceeds a certain threshold, and if so, writes it to a special stream of flagged observations.

### Testing data quality

If you want to keep tabs on your data quality, you can write a service that subscribes to every new record in a stream and checks that its values are within the expected ranges. If a record fails your tests, you can write it to a special stream of flagged observations – or even have the code send you an email!

### Synchronizing to an external system

Lets say you need to synchronize every sale you make to your accounting software. If you create a stream of sales in Beneath, you can setup a service that subscribes to it and automatically makes an API call to your accounting software for every sale. And even if your service or accounting software is down, the at-least-once guarantees ensure you account for every sale.

## Learn more

Event logs are arguably the most useful primitive in production data systems. That's exactly why we put them at the center of Beneath! If you want to try them out in action, use one of our [Quick starts](https://about.beneath.dev/docs/quick-starts/). If you would like to learn more about event logs in general, these are the best articles we've ever read on the topic:

- [The Log: What every software engineer should know about real-time data's unifying abstraction](https://engineering.linkedin.com/distributed-systems/log-what-every-software-engineer-should-know-about-real-time-datas-unifying)
- [Online Event Processing](https://queue.acm.org/detail.cfm?id=3321612)
