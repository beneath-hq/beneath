---
title: Schema language
description:
menu:
  docs:
    parent: writing-to-beneath
    weight: 210
weight: 210
---
When writing data to Beneath, you must first provide a file with the stream's schema. By explicitly defining the schema, Beneath's infrastructure knows exactly what to expect. The schema file should follow the format of the “GraphQL schema language.” 

##### Examples

**blocks.graphql**
```graphql
"Blocks loaded from Ethereum"
type Block @stream(name: "blocks", key: "number") {
  "Block number"
  number: Int!

  "Block timestamp"
  timestamp: Timestamp!

  "Block hash"
  hash: Bytes32!

  "Hash of parent block"
  parentHash: Bytes32!

  "Address of block miner"
  miner: Bytes20!

  "Size of block in bytes"
  size: Int!

  "Number of transactions in block"
  transactions: Int!

  "Block difficulty"
  difficulty: Numeric!

  "Extra data embedded in block by miner"
  extraData: Bytes!

  "Gas limit"
  gasLimit: Int!

  "Gas used"
  gasUsed: Int!
}
```

**dai-transfers.graphql**
```graphql
"Dai transfers including mints and burns"
type DaiTransfers @stream(name: "dai-transfers", key: ["time", "index"]) {
  transactionHash: String!
  time: Timestamp! 
  index: Int!
  name: String!
  from: String
  to: String
  value: String!
}
```

**dai-daily-activity.graphql**
```graphql
"Metrics on daily activity in Dai"
type DaiDailyActivity @stream(name: "dai-daily-activity", key: "day") {
  day: Timestamp!
  users: Int!
  senders: Int!
  receivers: Int!
  transfers: Int!
  minted: Float!
  burned: Float!
  turnover: Float!
}
```

##### Required and optional fields
An exclamation point after the type (e.g. “Float!”) indicates that the field is required. Fields without exclamation points (e.g. "Float") indicate that the field is not required.
