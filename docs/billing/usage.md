---
title: Usage and Billing Model
description: A description of how Beneath calculates usage for billing purposes
menu:
  docs:
    parent: billing
    weight: 100
weight: 100
---

TODO...

For the Professional and Enterprise plans, you are allotted monthly quotas for both Reads and Writes. At the beginning of each month, you are charged your plan's base price. At the end of the month, if your Read or Write usage exceeds your monthly quota, then you will be charged an "overage" fee. 

## The overage formula

The overage fee is simply calculated for both Reads and Writes:<br>
overageFee = (usageGB - quotaGB) * overagePricePerGB

## How usage is calculated

The usage of every row of data is measured as the sum of [Avro](https://en.wikipedia.org/wiki/Apache_Avro)-serialized bytes.

Avro serialization is a way to quickly and cost-efficiently transmit data across the internet. Avro uses JSON encoding to transmit the stream schema and uses binary encoding to compress and transmit the stream data.
