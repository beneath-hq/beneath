package engine

import (
	"context"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"gitlab.com/beneath-hq/beneath/infra/engine/driver"
	"gitlab.com/beneath-hq/beneath/infra/engine/driver/mock"
	"gitlab.com/beneath-hq/beneath/pkg/codec"
)

func TestExpandWarehouseQueryString(t *testing.T) {
	ctx := context.Background()
	e := &Engine{Warehouse: &mock.Mock{}}
	expandedQuery, err := e.ExpandWarehouseQuery(ctx, "select * from `exampleorg/example-proj/example_stream` join `/orgexample/proj-example/stream_example/`", streamResolver)
	assert.Nil(t, err)
	assert.Equal(t, "select * from exampleorg.example_proj.example_stream_00000000 join orgexample.proj_example.stream_example_00000000", expandedQuery)
}

func streamResolver(ctx context.Context, organizationName, projectName, streamName string) (driver.Project, driver.Stream, driver.StreamInstance, error) {
	p := project{
		organization: organizationName,
		project:      projectName,
	}
	s := stream{
		name: streamName,
	}
	si := streamInstance{}
	return p, s, si, nil
}

type project struct {
	id           uuid.UUID
	organization string
	project      string
	public       bool
}

func (p project) GetProjectID() uuid.UUID {
	return p.id
}

func (p project) GetOrganizationName() string {
	return p.organization
}

func (p project) GetProjectName() string {
	return p.project
}

func (p project) GetPublic() bool {
	return p.public
}

type stream struct {
	id   uuid.UUID
	name string
}

func (s stream) GetStreamID() uuid.UUID {
	return s.id
}

func (s stream) GetStreamName() string {
	return s.name
}

func (s stream) GetUseLog() bool {
	return true
}

func (s stream) GetUseIndex() bool {
	return true
}

func (s stream) GetUseWarehouse() bool {
	return true
}

func (s stream) GetLogRetention() time.Duration {
	return time.Duration(0)
}

func (s stream) GetIndexRetention() time.Duration {
	return time.Duration(0)
}

func (s stream) GetWarehouseRetention() time.Duration {
	return time.Duration(0)
}

func (s stream) GetCodec() *codec.Codec {
	return nil
}

type streamInstance struct {
	id uuid.UUID
}

func (i streamInstance) GetStreamInstanceID() uuid.UUID {
	return i.id
}
