The `generator` service uses change data capture to subscribe to a source database (e.g. Postgres) and write the changes to a Beneath table.

The service uses the [Debezium Engine](https://debezium.io/documentation/reference/1.6/development/engine.html) and checkpoints its state to Beneath. Typically, Debezium is deployed via Kafka Connect and checkpoints to Kafka, but we eliminate the Kafka dependency by using Beneath instead.

