package data

import (
	"github.com/beneath-hq/beneath/infra/engine"
	"github.com/beneath-hq/beneath/infra/mq"
	"github.com/beneath-hq/beneath/services/permissions"
	"github.com/beneath-hq/beneath/services/table"
	"github.com/beneath-hq/beneath/services/usage"
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
	Tables      *table.Service

	// manages real-time push (for websockets/streaming grpc clients)
	subscriptions subscriptions
}

// New returns a new data service instance
func New(logger *zap.Logger, mq mq.MessageQueue, engine *engine.Engine, usage *usage.Service, permissions *permissions.Service, tables *table.Service) (*Service, error) {
	err := mq.RegisterTopic(writeRequestsTopic, false)
	if err != nil {
		return nil, err
	}

	err = mq.RegisterTopic(writeReportsTopic, false)
	if err != nil {
		return nil, err
	}

	return &Service{
		Logger:      logger.Named("data").Sugar(),
		MQ:          mq,
		Engine:      engine,
		Usage:       usage,
		Permissions: permissions,
		Tables:      tables,
	}, nil
}
