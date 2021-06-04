# License

The entire source code for Beneath is publicly available on [Github](https://github.com/beneath-hq/beneath), however, not all of it meets the [definition of "open source"](https://opensource.org/osd). Beneath is free to use for most use cases, and will eventually become open source.

Beneath's core is licensed under a "source available" license, the [Business Source License](https://github.com/beneath-hq/beneath/blob/master/licenses/BSL.txt), which converts to the Apache 2.0 license after four years. It means you cannot run a large scale deployment of Beneath or re-sell Beneath as a managed service without buying a license from us. The idea behind this is simple: we want as many people as possible to benefit from Beneath, and we also want a viable business model that allows us to continue to promote and improve Beneath.

All the [client libraries](https://github.com/beneath-hq/beneath/tree/master/clients) and [examples](https://github.com/beneath-hq/beneath/tree/master/examples) are published under the MIT license and do meet the definition of "open source" – so you can freely integrate with Beneath in any software product, including in commercial software. See the [`LICENSE`](https://github.com/beneath-hq/beneath/blob/master/LICENSE) file for details.

It's also worth mentioning the `ee/` directory (short for "Enterprise Edition"). It contains source code for running Beneath at scale (e.g. code for charging for usage). It's also currently subject to the BSL license, but that may change in the future. Currently, we haven't developed the tools for building a useful non-ee binary, so the details around how to license this are still unclear.

What does all this mean for contributors? It means that if you want to contribute to the core platform, we will need you to sign a [Contributor License Agreement (CLA)](https://en.wikipedia.org/wiki/Contributor_License_Agreement). If you have any feedback or concerns about this, please don't hesitate to [reach out](https://about.beneath.dev/contact/).
