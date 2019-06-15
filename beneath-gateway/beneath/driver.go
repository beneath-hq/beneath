package beneath

// StreamsDriver defines the functions necessary to encapsulate Beneath's streaming data needs
type StreamsDriver interface {
	GetName() string
	GetMaxMessageSize() int
	PushWriteRequest([]byte) error
}

// TablesDriver defines the functions necessary to encapsulate Beneath's operational datastore needs
type TablesDriver interface {
	GetName() string
	GetMaxKeySize() int
	GetMaxDataSize() int
	// TODO: Add table functions
}
