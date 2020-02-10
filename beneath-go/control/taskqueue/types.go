package taskqueue

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/beneath-core/beneath-go/core/timeutil"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack"

	pb "github.com/beneath-core/beneath-go/engine/proto"
)

// Task is the abstract interface for a worker task
type Task interface {
	Run(ctx context.Context) error
}

var (
	taskRegistry map[string]reflect.Type
)

// RegisterTask registers a task struct for processing
func RegisterTask(task Task) {
	if taskRegistry == nil {
		taskRegistry = make(map[string]reflect.Type)
	}

	t := reflect.TypeOf(task)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("task is not a pointer to a struct"))
	}

	taskRegistry[t.Elem().Name()] = t.Elem()
}

// EncodeTask converts a task to a protobuf (the task struct must have been registered with RegisterTask)
func EncodeTask(task Task) (*pb.QueuedTask, error) {
	t := reflect.TypeOf(task)
	if t.Kind() != reflect.Ptr || t.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("task is not a pointer to a struct"))
	}

	data, err := msgpack.Marshal(task)
	if err != nil {
		return nil, err
	}

	return &pb.QueuedTask{
		UniqueId:  uuid.NewV4().String(),
		Name:      t.Elem().Name(),
		Timestamp: timeutil.UnixMilli(time.Now()),
		Data:      data,
	}, nil
}

// DecodeTask converts a pb.QueuedTask to a task registered with RegisterTask
func DecodeTask(qt *pb.QueuedTask) (Task, error) {
	t, ok := taskRegistry[qt.Name]
	if !ok {
		return nil, fmt.Errorf("cannot find task with name '%s'", qt.Name)
	}

	tval := reflect.New(t)              // pointer to underlying; underlying is raw struct
	task, ok := tval.Interface().(Task) // pointer to struct
	if !ok {
		return nil, fmt.Errorf("cannot cast type '%s' to Task", qt.Name)
	}

	err := msgpack.Unmarshal(qt.Data, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}
