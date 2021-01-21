---
title: Billing plans and quotas
description: How Beneath handles billing
menu:
  docs:
    parent: usage
    weight: 200
weight: 200
---

## Plans and billing

Every user and organization has a free quota included in their billing plan, and can upgrade to a paid plan to get larger fixed or pay-per-GB quotas.

If you select a pay-per-GB plan, your monthly bill will include the next month's service and the previous month's GB usage. Note that the default pay-per-GB plans have hard caps to prevent you from racking up a planet-scale bill.

All usage from services is counted against the project owner's billing quota. For multi-user organizations, each member user's usage is counted against the organization's billing plan.

## Billing periods

Billing periods run in 31 day cycles (by default) starting on the day you signed up for your billing plan. If you change billing plans, your billing cycle resets to the current time and your quota usage resets to zero.

## Minimum usage billed per request

For users and services, Beneath tracks a minimum of 1 KB of usage per read or write request. For requests that read or write several records, this threshold will easily be exceeded, but it can make a difference for requests that read or write just a single small record. (Note that the Python client library by default bundles writes made within 1 second of each other in a single request.)

## Custom plans or sponsorships

If we don't offer a plan that suits your needs, please [contact us]({{< ref "/contact" >}}) and we'll work something out.

We also sponsor free higher quotas for open-source projects and other useful public data on Beneath, so if you're hitting the limits of the free plan, [contact us]({{< ref "/contact" >}}) and we'll set up a sponsorship for your project.
