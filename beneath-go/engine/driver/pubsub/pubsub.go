package pubsub

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/beneath-core/beneath-go/core"
	"github.com/beneath-core/beneath-go/core/log"
	pb "github.com/beneath-core/beneath-go/proto"
	"github.com/golang/protobuf/proto"
)

// configSpecification defines the config variables to load from ENV
// See https://github.com/kelseyhightower/envconfig
type configSpecification struct {
	ProjectID                      string `envconfig:"PROJECT_ID" required:"true"`
	InstanceID                     string `envconfig:"INSTANCE_ID" required:"true"`
	SubscriberID                   string `envconfig:"SUBSCRIBER_ID" required:"true"`
	EmulatorHost                   string `envconfig:"EMULATOR_HOST" required:"false"`
	WriteRequestsTopic             string `envconfig:"WRITE_REQUESTS_TOPIC" required:"true"`
	WriteRequestsSubscription      string `envconfig:"WRITE_REQUESTS_SUBSCRIPTION" required:"true"`
	WriteReportsTopic              string `envconfig:"WRITE_REPORTS_TOPIC" required:"true"`
	WriteReportsSubscriptionPrefix string `envconfig:"WRITE_REPORTS_SUBSCRIPTION_PREFIX" required:"true"`
	TaskQueueTopic                 string `envconfig:"TASK_QUEUE_TOPIC" required:"true"`
	TaskQueueSubscription          string `envconfig:"TASK_QUEUE_SUBSCRIPTION" required:"true"`
}

// Pubsub implements beneath.StreamsDriver
type Pubsub struct {
	config             *configSpecification
	Client             *pubsub.Client
	WriteRequestsTopic *pubsub.Topic
	WriteReportsTopic  *pubsub.Topic
	TaskQueueTopic     *pubsub.Topic
}

// New returns a new
func New() *Pubsub {
	// parse config from env
	var config configSpecification
	core.LoadConfig("beneath_engine_pubsub", &config)

	// if EMULATOR_HOST set, configure pubsub for the emulator
	if config.EmulatorHost != "" {
		os.Setenv("PUBSUB_PROJECT_ID", config.ProjectID)
		os.Setenv("PUBSUB_EMULATOR_HOST", config.EmulatorHost)
	}

	// prepare pubsub client
	client, err := pubsub.NewClient(context.Background(), config.ProjectID)
	if err != nil {
		panic(err)
	}

	// create instance
	p := &Pubsub{
		config: &config,
		Client: client,
	}

	// set topics
	p.WriteRequestsTopic = p.makeTopic(config.WriteRequestsTopic)
	p.WriteReportsTopic = p.makeTopic(config.WriteReportsTopic)
	p.TaskQueueTopic = p.makeTopic(config.TaskQueueTopic)

	// done
	return p
}

// GetMaxMessageSize implements beneath.StreamsDriver
func (p *Pubsub) GetMaxMessageSize() int {
	return 10000000
}

// QueueWriteRequest implements beneath.StreamsDriver
func (p *Pubsub) QueueWriteRequest(ctx context.Context, req *pb.WriteRecordsRequest) error {
	// encode message
	msg, err := proto.Marshal(req)
	if err != nil {
		panic(err)
	}

	// check encoded message size
	if len(msg) > p.GetMaxMessageSize() {
		return fmt.Errorf(
			"write request has invalid size <%d> (max message size is <%d bytes>)",
			len(msg), p.GetMaxMessageSize(),
		)
	}

	// push
	result := p.WriteRequestsTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	// blocks until ack'ed by pubsub
	_, err = result.Get(ctx)
	return err
}

// ReadWriteRequests implements beneath.StreamsDriver
func (p *Pubsub) ReadWriteRequests(fn func(context.Context, *pb.WriteRecordsRequest) error) error {
	// prepare subscription and context
	sub := p.getWriteRequestsSubscription()
	cctx, cancel := context.WithCancel(context.Background())

	// receive pubsub messages forever (or until error occurs)
	err := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		// called for every single message on pubsub
		// these messages will have been written by QueueWriteRequest

		// unmarshal write request from pubsub
		req := &pb.WriteRecordsRequest{}
		err := proto.Unmarshal(msg.Data, req)
		if err != nil {
			// standard error handling for pubsub library
			log.S.Errorf("couldn't unmarshal write records request: %s", err.Error())
			cancel()
			return
		}

		// trigger callback function
		err = fn(ctx, req)
		if err != nil {
			// TODO: we'll want to keep the pipeline going in the future when things are stable
			// log error and cancel
			log.S.Errorf("couldn't process write request: %s", err.Error())
			cancel()
			return
		}

		// ack the message -- all processing finished successfully
		msg.Ack()
	})

	// pubsub stopped listening for new messages for some reason
	// return the error
	return err
}

// QueueWriteReport implements beneath.StreamsDriver
func (p *Pubsub) QueueWriteReport(ctx context.Context, rep *pb.WriteRecordsReport) error {
	// encode message
	msg, err := proto.Marshal(rep)
	if err != nil {
		panic(err)
	}

	// check encoded message size
	if len(msg) > p.GetMaxMessageSize() {
		return fmt.Errorf(
			"write report has invalid size <%d> (max message size is <%d bytes>)",
			len(msg), p.GetMaxMessageSize(),
		)
	}

	// push
	result := p.WriteReportsTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	// blocks until ack'ed by pubsub
	_, err = result.Get(ctx)
	return err
}

// ReadWriteReports implements beneath.StreamsDriver
func (p *Pubsub) ReadWriteReports(fn func(context.Context, *pb.WriteRecordsReport) error) error {
	// prepare subscription and context
	sub := p.getWriteReportsSubscription()
	cctx, cancel := context.WithCancel(context.Background())

	// receive pubsub messages forever (or until error occurs)
	err := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		// ack the message -- all processing finished successfully
		msg.Ack()

		// called for every single message on pubsub
		// these messages will have been written by QueueWriteReport

		// unmarshal write report from pubsub
		rep := &pb.WriteRecordsReport{}
		err := proto.Unmarshal(msg.Data, rep)
		if err != nil {
			// standard error handling for pubsub library
			log.S.Errorf("couldn't unmarshal write report: %s", err.Error())
			cancel()
			return
		}

		// trigger callback function
		err = fn(ctx, rep)
		if err != nil {
			// TODO: we'll want to keep the pipeline going in the future when things are stable
			// log error and cancel
			log.S.Errorf("couldn't process write report: %s", err.Error())
			cancel()
			return
		}
	})

	// pubsub stopped listening for new messages for some reason
	// return the error
	return err
}

// QueueTask queues a task for processing
func (p *Pubsub) QueueTask(ctx context.Context, t *pb.QueuedTask) error {
	// encode message
	msg, err := proto.Marshal(t)
	if err != nil {
		panic(err)
	}

	// check encoded message size
	if len(msg) > p.GetMaxMessageSize() {
		return fmt.Errorf(
			"task has invalid size <%d> (max message size is <%d bytes>)",
			len(msg), p.GetMaxMessageSize(),
		)
	}

	// push
	result := p.TaskQueueTopic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	// blocks until ack'ed by pubsub
	_, err = result.Get(ctx)
	return err
}

// ReadTasks reads queued tasks
func (p *Pubsub) ReadTasks(fn func(context.Context, *pb.QueuedTask) error) error {
	// prepare subscription and context
	sub := p.getTaskQueueSubscription()
	cctx, cancel := context.WithCancel(context.Background())

	// receive pubsub messages forever (or until error occurs)
	err := sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		// called for every single message on pubsub
		// these messages will have been written by QueueTask

		// unmarshal task from pubsub
		t := &pb.QueuedTask{}
		err := proto.Unmarshal(msg.Data, t)
		if err != nil {
			// standard error handling for pubsub library
			log.S.Errorf("couldn't unmarshal queued task: %s", err.Error())
			cancel()
			return
		}

		// trigger callback function
		err = fn(ctx, t)
		if err != nil {
			// TODO: we'll want to keep the pipeline going in the future when things are stable
			// log error and cancel
			log.S.Errorf("couldn't process write report: %s", err.Error())
			cancel()
			return
		}

		// ack the message -- all processing finished successfully
		msg.Ack()
	})

	// pubsub stopped listening for new messages for some reason
	// return the error
	return err
}

// makeTopic returns a topic and creates it if it doesn't exist
func (p *Pubsub) makeTopic(name string) *pubsub.Topic {
	topic, err := p.Client.CreateTopic(context.Background(), name)
	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			panic(err)
		} else {
			topic = p.Client.Topic(name)
		}
	}
	return topic
}

// getSubscription finds or creates a topic subscription
func (p *Pubsub) getSubscription(topic *pubsub.Topic, subname string) *pubsub.Subscription {
	// create/get subscriber to topic
	subscription, err := p.Client.CreateSubscription(context.Background(), subname, pubsub.SubscriptionConfig{
		Topic:       topic,
		AckDeadline: 20 * time.Second,
	})

	if err != nil {
		status, ok := status.FromError(err)
		if !ok || status.Code() != codes.AlreadyExists {
			panic(fmt.Errorf("error creating subscription '%s': %v", subname, err))
		} else {
			subscription = p.Client.Subscription(subname)
		}
	}

	return subscription
}

// get subscription for WriteRequests
func (p *Pubsub) getWriteRequestsSubscription() *pubsub.Subscription {
	return p.getSubscription(p.WriteRequestsTopic, p.config.WriteRequestsSubscription)
}

// get subscription for TaskQueue
func (p *Pubsub) getTaskQueueSubscription() *pubsub.Subscription {
	return p.getSubscription(p.TaskQueueTopic, p.config.TaskQueueTopic)
}

// get Metrics subscription for the relevant subscriber
func (p *Pubsub) getWriteReportsSubscription() *pubsub.Subscription {
	subname := fmt.Sprintf("%s-%s", p.config.WriteReportsSubscriptionPrefix, p.config.SubscriberID)
	sub := p.getSubscription(p.WriteReportsTopic, subname)
	err := sub.SeekToTime(context.Background(), time.Now())
	if err != nil {
		status, ok := status.FromError(err)
		if ok && status.Code() == codes.Unimplemented && p.config.EmulatorHost != "" {
			// Seek not implemented on Emulator, ignore
		} else {
			panic(fmt.Errorf("error seeking on subscription '%s': %v", subname, err))
		}
	}
	return sub
}
