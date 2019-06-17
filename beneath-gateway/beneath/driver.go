package beneath

import pb "github.com/beneath-core/beneath-gateway/beneath/beneath_proto"

// StreamsDriver defines the functions necessary to encapsulate Beneath's streaming data needs
type StreamsDriver interface {
	GetMaxMessageSize() int
	PushWriteRequest(req *pb.WriteInternalRecordsRequest) error
}

// TablesDriver defines the functions necessary to encapsulate Beneath's operational datastore needs
type TablesDriver interface {
	GetMaxKeySize() int
	GetMaxDataSize() int
	// TODO: Add table functions
}
