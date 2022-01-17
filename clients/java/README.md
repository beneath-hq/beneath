# Beneath Java Client Library

[![License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://github.com/beneath-hq/beneath/blob/master/clients/LICENSE)

This folder contains the source code for the [Beneath](https://beneath.dev) Java library. 

Thus far, the Beneath Java library has only been used to develop the [Beneath CDC service](https://github.com/beneath-hq/beneath/tree/master/examples/cdc). The primary Beneath SDK is the [Python client](https://github.com/beneath-hq/beneath/tree/master/clients/python).

### Installation

With Gradle:

```groovy
repositories {
  mavenCentral()
}

dependencies {
  implementation "dev.beneath:beneath:1.0.0"
}
```
