package data

import (
	"gitlab.com/beneath-hq/beneath/infra/engine"
	"gitlab.com/beneath-hq/beneath/infra/mq"
	"gitlab.com/beneath-hq/beneath/services/permissions"
	"gitlab.com/beneath-hq/beneath/services/stream"
	"gitlab.com/beneath-hq/beneath/services/usage"
	"go.uber.org/zap"
)

// Service implements the data-plane functionality for handling requests
// (in a protocol-agnostic way used by the http and grpc interfaces), and for
// processing records in the background.
type Service struct {
	Logger      *zap.SugaredLogger
	MQ          mq.MessageQueue
	Engine      *engine.Engine
	Usage       *usage.Service
	Permissions *permissions.Service
	Streams     *stream.Service

	// manages real-time push (for websockets/streaming grpc clients)
	subscriptions subscriptions
}

// New returns a new data service instance
func New(logger *zap.Logger, mq mq.MessageQueue, engine *engine.Engine, usage *usage.Service, permissions *permissions.Service, streams *stream.Service) (*Service, error) {
	err := mq.RegisterTopic(writeRequestsTopic)
	if err != nil {
		return nil, err
	}

	err = mq.RegisterTopic(writeReportsTopic)
	if err != nil {
		return nil, err
	}

	return &Service{
		Logger:      logger.Named("data").Sugar(),
		MQ:          mq,
		Engine:      engine,
		Usage:       usage,
		Permissions: permissions,
		Streams:     streams,
	}, nil
}
