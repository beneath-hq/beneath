package engine

import (
	"context"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	"github.com/beneath-hq/beneath/infra/engine/driver/mock"
	"github.com/beneath-hq/beneath/pkg/codec"
)

func TestExpandWarehouseQueryString(t *testing.T) {
	ctx := context.Background()
	e := &Engine{Warehouse: &mock.Mock{}}
	expandedQuery, err := e.ExpandWarehouseQuery(ctx, "select * from `exampleorg/example-proj/table:example_table` join `/orgexample/proj-example/table_example/`", tableResolver)
	assert.Nil(t, err)
	assert.Equal(t, "select * from exampleorg.example_proj.example_table_00000000 join orgexample.proj_example.table_example_00000000", expandedQuery)
}

func tableResolver(ctx context.Context, organizationName, projectName, tableName string) (driver.Project, driver.Table, driver.TableInstance, error) {
	p := project{
		organization: organizationName,
		project:      projectName,
	}
	s := table{
		name: tableName,
	}
	si := tableInstance{}
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

type table struct {
	id   uuid.UUID
	name string
}

func (s table) GetTableID() uuid.UUID {
	return s.id
}

func (s table) GetTableName() string {
	return s.name
}

func (s table) GetUseLog() bool {
	return true
}

func (s table) GetUseIndex() bool {
	return true
}

func (s table) GetUseWarehouse() bool {
	return true
}

func (s table) GetLogRetention() time.Duration {
	return time.Duration(0)
}

func (s table) GetIndexRetention() time.Duration {
	return time.Duration(0)
}

func (s table) GetWarehouseRetention() time.Duration {
	return time.Duration(0)
}

func (s table) GetCodec() *codec.Codec {
	return nil
}

type tableInstance struct {
	id uuid.UUID
}

func (i tableInstance) GetTableInstanceID() uuid.UUID {
	return i.id
}
