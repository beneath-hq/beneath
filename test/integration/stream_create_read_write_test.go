package integration

import (
	"bytes"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/linkedin/goavro/v2"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"

	pb "github.com/beneath-core/gateway/grpc/proto"
)

func TestStreamCreateReadAndWrite(t *testing.T) {
	// create project
	res1 := queryGQL(`
		mutation CreateProject($organizationID: UUID!) {
			createProject(name: "test", organizationID: $organizationID, public: true) {
				projectID
				name
				users { userID }
			}
		}
	`, map[string]interface{}{"organizationID": testUser.OrganizationID})
	assert.Empty(t, res1.Errors)
	project := res1.Data["createProject"]
	assert.Equal(t, "test", project["name"])
	assert.NotEmpty(t, project["users"])

	// create stream
	schema := readTestdata("foobar.graphql")
	res2 := queryGQL(`
		mutation CreateExternalStream($projectID: UUID!, $schema: String!) {
			createExternalStream(projectID: $projectID, schema: $schema, batch: false, manual: false) {
				streamID
				avroSchema
				currentStreamInstanceID
				name
				project { projectID }
			}
		}
	`, map[string]interface{}{
		"projectID": project["projectID"],
		"schema":    schema,
	})
	assert.Empty(t, res2.Errors)
	stream := res2.Data["createExternalStream"]
	instanceID := uuid.FromStringOrNil(stream["currentStreamInstanceID"].(string))
	assert.Equal(t, "foo_bar", stream["name"])
	assert.Len(t, stream["project"], 1)
	assert.False(t, instanceID == uuid.Nil)

	// prepare to write to stream
	foobars := nextFoobars(100)
	split := 50
	codec, err := goavro.NewCodec(stream["avroSchema"].(string))
	assert.Nil(t, err)

	// compile grpc records
	subset := foobars[0:split]
	recordsPB := make([]*pb.Record, len(subset))
	for i, foobar := range subset {
		data, err := codec.BinaryFromNative(nil, foobar.Native())
		assert.Nil(t, err)
		recordsPB[i] = &pb.Record{
			AvroData: data,
		}
	}

	// write grpc records
	res3, err := gatewayGRPC.Write(grpcContext(), &pb.WriteRequest{
		InstanceId: instanceID.Bytes(),
		Records:    recordsPB,
	})
	assert.Nil(t, err)
	assert.Len(t, res3.GetWriteId(), 20)

	// wait to let writes happen
	time.Sleep(200 * time.Millisecond)

	// test grpc data replay (both compact and uncompact)
	for _, compact := range []bool{false, true} {
		// if compact, output will be sorted by primary key, so we must sort the expected values
		expected := recordsPB
		if compact {
			expected = make([]*pb.Record, len(recordsPB))
			copy(expected, recordsPB)
			sort.Slice(expected, func(i, j int) bool {
				return bytes.Compare(expected[i].AvroData, expected[j].AvroData) < 0
			})
		}

		res4, err := gatewayGRPC.Query(grpcContext(), &pb.QueryRequest{
			InstanceId: instanceID.Bytes(),
			Compact:    compact,
			Partitions: 1,
		})
		assert.Nil(t, err)
		assert.Len(t, res4.ReplayCursors, 1)
		assert.Len(t, res4.ChangeCursors, 1)

		// read data page 1
		n := 30
		res5, err := gatewayGRPC.Read(grpcContext(), &pb.ReadRequest{
			InstanceId: instanceID.Bytes(),
			Cursor:     res4.ReplayCursors[0],
			Limit:      int32(n),
		})
		assert.Nil(t, err)
		assert.Len(t, res5.Records, n)
		assert.True(t, len(res5.NextCursor) > 1)
		for idx, record := range res5.Records {
			assert.Equal(t, record.AvroData, expected[idx].AvroData)
		}

		// read data page 2 (ends here)
		res6, err := gatewayGRPC.Read(grpcContext(), &pb.ReadRequest{
			InstanceId: instanceID.Bytes(),
			Cursor:     res5.NextCursor,
			Limit:      int32(n),
		})
		assert.Nil(t, err)
		assert.Len(t, res6.Records, 20)
		assert.Len(t, res6.NextCursor, 0)
		for idx, record := range res6.Records {
			assert.Equal(t, record.AvroData, expected[n+idx].AvroData)
		}
	}

	// // TODO: query filtered data with grpc
	// gatewayGRPC.Query(grpcContext(), &pb.QueryRequest{
	// 	InstanceId:
	// })

	// TODO: subscribe with grpc

	// TODO: subscribe with rest

	// write http records
	subset = foobars[split:]
	code, res7 := queryGatewayHTTP("POST", fmt.Sprintf("streams/instances/%s", instanceID.String()), subset)
	assert.Equal(t, 200, code)
	assert.Nil(t, res7)

	// wait to let writes happen
	time.Sleep(200 * time.Millisecond)

	// TODO: check subscription receives

	// TODO: query with rest
}
