/*
Package ws provides a modified implementation of the Websockets-based
RPC protocol used in [apollographql/subscriptions-transport-ws](https://github.com/apollographql/subscriptions-transport-ws/blob/master/PROTOCOL.md).

The main difference is that this implementation accepts any query payload, not just GraphQL queries.

If we ever need to increase performance, this [blog post](https://github.com/eranyanay/1m-go-websockets) was very inspiring.
*/
package ws
