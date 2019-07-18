package engine

import pb "github.com/beneath-core/beneath-go/proto"

// StreamsDriver defines the functions necessary to encapsulate Beneath's streaming data needs
type StreamsDriver interface {
	// GetMaxMessageSize returns the maximum accepted message size in bytes
	GetMaxMessageSize() int

	// QueueWriteRequest queues a write request -- concretely, it results in
	// the write request being written to Pubsub, then from there read by
	// the data processing pipeline and written to BigTable and BigQuery
	QueueWriteRequest(req *pb.WriteRecordsRequest) error
}

// TablesDriver defines the functions necessary to encapsulate Beneath's operational datastore needs
type TablesDriver interface {
	// GetMaxKeySize returns the maximum accepted key size in bytes
	GetMaxKeySize() int

	// GetMaxDataSize returns the maximum accepted value size in bytes
	GetMaxDataSize() int
}

// WarehouseDriver defines the functions necessary to encapsulate Beneath's data archiving needs
type WarehouseDriver interface {
	// GetMaxDataSize returns the maximum accepted row size in bytes
	GetMaxDataSize() int
}
