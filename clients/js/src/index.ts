import BrowserClient from "./BrowserClient";
export default {
  BrowserClient,
};

/*

USE CASES

- Create one to X mapping model in Node
- Load or write data from server-side Node.js
- Load data directly from front-end
  - Filtered query, loading more and more
  - Peek, loading more and more
  - In react, with useRecords

REST VS GRPC

- WS vs SUB
- PEEK and QUERY return records
- No ping

Client(secret, provider="rest"|"grpc")
this.connection
this.find_stream(project, name)

RESTConnection

GRPCConnection

RESTStream

GRPCStream

*/

