package data

import (
	"gitlab.com/beneath-hq/beneath/infrastructure/engine"
	"gitlab.com/beneath-hq/beneath/infrastructure/mq"
	"gitlab.com/beneath-hq/beneath/services/metrics"
	"gitlab.com/beneath-hq/beneath/services/permissions"
	"gitlab.com/beneath-hq/beneath/services/stream"
)

// Service implements the data-plane functionality for handling requests
// (in a protocol-agnostic way used by the http and grpc interfaces), and for
// processing records in the background.
type Service struct {
	MQ          mq.MessageQueue
	Engine      *engine.Engine
	Metrics     *metrics.Broker
	Permissions *permissions.Service
	Streams     *stream.Service

	// manages real-time push (for websockets/streaming grpc clients)
	subscriptions subscriptions
}

// New returns a new data service instance
func New(mq mq.MessageQueue, engine *engine.Engine, metrics *metrics.Broker, permissions *permissions.Service, streams *stream.Service) (*Service, error) {
	err := mq.RegisterTopic(writeRequestsTopic)
	if err != nil {
		return nil, err
	}

	err = mq.RegisterTopic(writeReportsTopic)
	if err != nil {
		return nil, err
	}

	return &Service{
		MQ:          mq,
		Engine:      engine,
		Metrics:     metrics,
		Permissions: permissions,
		Streams:     streams,
	}, nil
}
