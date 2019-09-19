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
	return p.queueProto(ctx, p.WriteRequestsTopic, req)
}

// QueueWriteReport implements beneath.StreamsDriver
func (p *Pubsub) QueueWriteReport(ctx context.Context, rep *pb.WriteRecordsReport) error {
	return p.queueProto(ctx, p.WriteReportsTopic, rep)
}

// QueueTask implements beneath.StreamsDriver
func (p *Pubsub) QueueTask(ctx context.Context, t *pb.QueuedTask) error {
	return p.queueProto(ctx, p.TaskQueueTopic, t)
}

func (p *Pubsub) queueProto(ctx context.Context, topic *pubsub.Topic, pb proto.Message) error {
	// encode message
	msg, err := proto.Marshal(pb)
	if err != nil {
		panic(err)
	}

	// check encoded message size
	if len(msg) > p.GetMaxMessageSize() {
		return fmt.Errorf(
			"message has invalid size <%d> (max message size is <%d bytes>)",
			len(msg), p.GetMaxMessageSize(),
		)
	}

	// push
	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte(msg),
	})

	// blocks until ack'ed by pubsub
	_, err = result.Get(ctx)
	return err
}

// ReadWriteRequests implements beneath.StreamsDriver
func (p *Pubsub) ReadWriteRequests(fn func(context.Context, *pb.WriteRecordsRequest) error) error {
	sub := p.getSubscription(p.WriteRequestsTopic, p.config.WriteRequestsSubscription)
	cctx, cancel := context.WithCancel(context.Background())

	// receive messages forever (or until error occurs)
	return sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		// unmarshal write request from pubsub
		req := &pb.WriteRecordsRequest{}
		err := proto.Unmarshal(msg.Data, req)
		if err != nil {
			log.S.Errorf("couldn't unmarshal write records request: %s", err.Error())
			cancel()
			return
		}

		// trigger callback function
		err = fn(ctx, req)
		if err != nil {
			// TODO: we'll want to keep the pipeline going in the future when things are stable
			log.S.Errorf("couldn't process write request: %s", err.Error())
			cancel()
			return
		}

		// ack after processing
		msg.Ack()
	})
}

// ReadWriteReports implements beneath.StreamsDriver
func (p *Pubsub) ReadWriteReports(fn func(context.Context, *pb.WriteRecordsReport) error) error {
	// prepare context
	cctx, cancel := context.WithCancel(context.Background())

	// prepare subscription (seek to now, skipping messages from downtime)
	subname := fmt.Sprintf("%s-%s", p.config.WriteReportsSubscriptionPrefix, p.config.SubscriberID)
	sub := p.getSubscription(p.WriteReportsTopic, subname)
	err := sub.SeekToTime(cctx, time.Now())
	if err != nil {
		status, ok := status.FromError(err)
		if ok && status.Code() == codes.Unimplemented && p.config.EmulatorHost != "" {
			// Seek not implemented on Emulator, ignore
		} else {
			panic(fmt.Errorf("error seeking on subscription '%s': %v", subname, err))
		}
	}

	// receive messages forever (or until error occurs)
	return sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		// ack first -- don't want reprocessing on failure
		msg.Ack()

		// unmarshal write report
		rep := &pb.WriteRecordsReport{}
		err := proto.Unmarshal(msg.Data, rep)
		if err != nil {
			log.S.Errorf("couldn't unmarshal write report: %s", err.Error())
			cancel()
			return
		}

		// trigger callback function
		err = fn(ctx, rep)
		if err != nil {
			// TODO: we'll want to keep the pipeline going in the future when things are stable
			log.S.Errorf("couldn't process write report: %s", err.Error())
			cancel()
			return
		}
	})
}

// ReadTasks reads queued tasks
func (p *Pubsub) ReadTasks(fn func(context.Context, *pb.QueuedTask) error) error {
	// prepare subscription and context
	sub := p.getSubscription(p.TaskQueueTopic, p.config.TaskQueueTopic)
	cctx, cancel := context.WithCancel(context.Background())

	// receive messages forever (or until error occurs)
	return sub.Receive(cctx, func(ctx context.Context, msg *pubsub.Message) {
		// unmarshal task from pubsub
		t := &pb.QueuedTask{}
		err := proto.Unmarshal(msg.Data, t)
		if err != nil {
			log.S.Errorf("couldn't unmarshal queued task: %s", err.Error())
			cancel()
			return
		}

		// trigger callback function
		err = fn(ctx, t)
		if err != nil {
			// TODO: we'll want to keep the pipeline going in the future when things are stable
			log.S.Errorf("couldn't process queued task: %s", err.Error())
			cancel()
			return
		}

		// ack after processing
		msg.Ack()
	})
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
