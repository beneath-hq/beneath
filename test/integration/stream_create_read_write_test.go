package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/linkedin/goavro/v2"
	"github.com/mr-tron/base58"
	uuid "github.com/satori/go.uuid"
	assert "github.com/stretchr/testify/require"

	"gitlab.com/beneath-hq/beneath/pkg/timeutil"
	pb "gitlab.com/beneath-hq/beneath/server/data/grpc/proto"
)

func TestStreamCreateReadAndWrite(t *testing.T) {
	startTime := time.Now()

	// create project
	res1 := queryGQL(`
		mutation CreateProject($organizationID: String!) {
			createProject(organizationID: $organizationID, projectName: "test", public: true) {
				projectID
				name
			}
		}
	`, map[string]interface{}{"organizationID": testOrg.OrganizationID})
	assert.Empty(t, res1.Errors)
	project := res1.Result()["createProject"]
	assert.Equal(t, "test", project["name"])

	// create stream
	schema := readTestdata("foobar.graphql")
	res2 := queryGQL(`
		mutation CreateStream($organization: String!, $schema: String!) {
			createStream(
				input: {
					organizationName: $organization,
					projectName: "test",
					streamName: "foo_bar",
					schemaKind: GraphQL,
					schema: $schema,
					updateIfExists: true,
				}
			) {
				streamID
				avroSchema
				primaryStreamInstanceID
				name
				project { projectID }
			}
		}
	`, map[string]interface{}{
		"organization": testOrg.Name,
		"schema":       schema,
	})
	assert.Empty(t, res2.Errors)
	stream := res2.Result()["createStream"]
	assert.Equal(t, "foo_bar", stream["name"])
	assert.Len(t, stream["project"], 1)

	//
	res2_1 := queryGQL(`
		mutation CreateStreamInstance($streamID: UUID!) {
			createStreamInstance(
				input: {
					streamID: $streamID,
					version: 0,
					makePrimary: true,
					updateIfExists: true,
				}
			) {
				streamInstanceID
				streamID
				createdOn
				version
				madePrimaryOn
				madeFinalOn
			}
		}
	`, map[string]interface{}{
		"streamID": stream["streamID"],
	})
	assert.Empty(t, res2_1.Errors)
	instance := res2_1.Result()["createStreamInstance"]
	instanceID := uuid.FromStringOrNil(instance["streamInstanceID"].(string))
	assert.False(t, instanceID == uuid.Nil)
	assert.Equal(t, float64(0), instance["version"])
	assert.NotNil(t, instance["madePrimaryOn"])
	assert.Nil(t, instance["madeFinalOn"])

	// prepare to write to stream
	foobars := nextFoobars(100)
	split := 50
	codec, err := goavro.NewCodec(stream["avroSchema"].(string))
	assert.Nil(t, err)

	// compile grpc records
	subset := foobars[0:split]
	recordsPB := make([]*pb.Record, len(subset))
	for i, foobar := range subset {
		data, err := codec.BinaryFromNative(nil, foobar.AvroNative())
		assert.Nil(t, err)
		recordsPB[i] = &pb.Record{
			AvroData: data,
		}
	}

	// write grpc records
	res3, err := gatewayGRPC.Write(grpcContext(), &pb.WriteRequest{
		InstanceRecords: []*pb.InstanceRecords{
			&pb.InstanceRecords{
				InstanceId: instanceID.Bytes(),
				Records:    recordsPB,
			},
		},
	})
	assert.Nil(t, err)
	assert.Len(t, res3.GetWriteId(), 16)

	// wait to let writes happen
	time.Sleep(200 * time.Millisecond)

	// test grpc log data replay
	res4, err := gatewayGRPC.QueryLog(grpcContext(), &pb.QueryLogRequest{
		InstanceId: instanceID.Bytes(),
		Partitions: 1,
	})
	assert.Nil(t, err)
	assert.Len(t, res4.ReplayCursors, 1)
	assert.Len(t, res4.ChangeCursors, 1)

	// read data page 1
	n := 30
	res5, err := gatewayGRPC.Read(grpcContext(), &pb.ReadRequest{
		Cursor: res4.ReplayCursors[0],
		Limit:  int32(n),
	})
	assert.Nil(t, err)
	assert.Len(t, res5.Records, n)
	assert.True(t, len(res5.NextCursor) > 1)
	for idx, record := range res5.Records {
		assert.Equal(t, record.AvroData, recordsPB[idx].AvroData)
	}

	// read data page 2 (ends here)
	res6, err := gatewayGRPC.Read(grpcContext(), &pb.ReadRequest{
		Cursor: res5.NextCursor,
		Limit:  int32(n),
	})
	assert.Nil(t, err)
	assert.Len(t, res6.Records, 20)
	assert.Len(t, res6.NextCursor, 0)
	for idx, record := range res6.Records {
		assert.Equal(t, record.AvroData, recordsPB[n+idx].AvroData)
	}

	// test grpc index data replay

	// output will be sorted by primary key, so we must sort the expected values
	expected := make([]*pb.Record, len(recordsPB))
	copy(expected, recordsPB)
	sort.Slice(expected, func(i, j int) bool {
		return bytes.Compare(expected[i].AvroData, expected[j].AvroData) < 0
	})

	res4a, err := gatewayGRPC.QueryIndex(grpcContext(), &pb.QueryIndexRequest{
		InstanceId: instanceID.Bytes(),
		Partitions: 1,
	})
	assert.Nil(t, err)
	assert.Len(t, res4a.ReplayCursors, 1)
	assert.Len(t, res4a.ChangeCursors, 1)

	// read data page 1
	n = 30
	res5a, err := gatewayGRPC.Read(grpcContext(), &pb.ReadRequest{
		Cursor: res4a.ReplayCursors[0],
		Limit:  int32(n),
	})
	assert.Nil(t, err)
	assert.Len(t, res5a.Records, n)
	assert.True(t, len(res5a.NextCursor) > 1)
	for idx, record := range res5a.Records {
		assert.Equal(t, record.AvroData, expected[idx].AvroData)
	}

	// read data page 2 (ends here)
	res6a, err := gatewayGRPC.Read(grpcContext(), &pb.ReadRequest{
		Cursor: res5a.NextCursor,
		Limit:  int32(n),
	})
	assert.Nil(t, err)
	assert.Len(t, res6a.Records, 20)
	assert.Len(t, res6a.NextCursor, 0)
	for idx, record := range res6a.Records {
		assert.Equal(t, record.AvroData, expected[n+idx].AvroData)
	}

	// query filtered data with grpc
	res7, err := gatewayGRPC.QueryIndex(grpcContext(), &pb.QueryIndexRequest{
		InstanceId: instanceID.Bytes(),
		Filter:     `{ "a": { "_prefix": "b" } }`,
		Partitions: 1,
	})
	assert.Nil(t, err)
	assert.Len(t, res7.ReplayCursors, 1)
	assert.Len(t, res7.ChangeCursors, 1)

	// the filter will match 2 rows

	// read data page 1 (should return exactly the two rows)
	res8, err := gatewayGRPC.Read(grpcContext(), &pb.ReadRequest{
		Cursor: res7.ReplayCursors[0],
		Limit:  2,
	})
	assert.Nil(t, err)
	assert.Len(t, res8.Records, 2)
	for _, record := range res8.Records {
		assert.NotEqual(t, record.Timestamp, 0)
		native, rem, err := codec.NativeFromBinary(record.AvroData)
		assert.Nil(t, err)
		assert.Len(t, rem, 0)
		assert.Equal(t, byte('b'), native.(map[string]interface{})["a"].(string)[0])
	}

	// read data page 2 (should be empty)
	res9, err := gatewayGRPC.Read(grpcContext(), &pb.ReadRequest{
		Cursor: res8.NextCursor,
		Limit:  10,
	})
	assert.Nil(t, err)
	assert.Len(t, res9.Records, 0)
	assert.Nil(t, res9.NextCursor)

	// get cursor for subscriptions
	res10, err := gatewayGRPC.QueryIndex(grpcContext(), &pb.QueryIndexRequest{
		InstanceId: instanceID.Bytes(),
		Partitions: 1,
	})
	assert.Nil(t, err)
	assert.Len(t, res10.ReplayCursors, 1)
	assert.Len(t, res10.ChangeCursors, 1)

	recvGRPC := false
	recvWS := false

	// subscribe with grpc
	go func() {
		// create sub
		ctx, cancel := context.WithCancel(grpcContext())
		sub, err := gatewayGRPC.Subscribe(ctx, &pb.SubscribeRequest{
			Cursor: res10.ChangeCursors[0],
		})
		panicIf(err)

		// read until we set recvGRPC = true
		for {
			_, err := sub.Recv()
			panicIf(err)
			recvGRPC = true
			cancel()
			break
		}
	}()

	// subscribe with websockets
	go func() {
		// create sub
		u := url.URL{
			Scheme: "ws",
			Host:   gatewayHTTP.Listener.Addr().String(),
			Path:   "/v1/-/ws",
		}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		panicIf(err)
		defer c.Close()

		// authenticate
		err = c.WriteJSON(map[string]interface{}{
			"type": "connection_init",
			"payload": map[string]interface{}{
				"secret": testSecret.Token.String(),
			},
		})
		panicIf(err)

		// subscribe to cursor
		err = c.WriteJSON(map[string]interface{}{
			"type": "start",
			"id":   "testid",
			"payload": map[string]interface{}{
				"instance_id": instanceID.String(),
				"cursor":      res10.ChangeCursors[0],
			},
		})
		panicIf(err)

		// read until we set recvWS = true
		for {
			_, message, err := c.ReadMessage()
			panicIf(err)

			var res map[string]interface{}
			err = json.Unmarshal(message, &res)
			panicIf(err)

			if res["id"] == "testid" {
				recvWS = true
				break
			}
		}
	}()

	// check subscriptions not yet triggered
	assert.False(t, recvGRPC)
	assert.False(t, recvWS)

	// write http records
	subset = foobars[split:]
	code, res11 := queryGatewayHTTP(http.MethodPost, fmt.Sprintf("v1/-/instances/%s", instanceID.String()), subset)
	assert.Equal(t, 200, code)
	assert.NotEqual(t, uuid.Nil, uuid.FromStringOrNil(res11["write_id"].(string)))

	// wait to let writes happen
	time.Sleep(200 * time.Millisecond)

	// check subscriptions triggered
	assert.True(t, recvGRPC)
	assert.True(t, recvWS)

	// query change data with REST
	changeCursor := base58.Encode(res10.ChangeCursors[0])
	code, res12 := queryGatewayHTTP(http.MethodGet, fmt.Sprintf(`v1/%s/test/foo-bar?cursor=%s`, testOrg.Name, changeCursor), nil)
	assert.Equal(t, 200, code)
	assert.Len(t, res12["meta"], 2)
	assert.Len(t, res12["data"], 50)
	for idx, recordT := range res12["data"].([]interface{}) {
		record := recordT.(map[string]interface{})
		assert.NotNil(t, record)
		assert.Len(t, record["@meta"], 2)
		assert.Equal(t, subset[idx].A, record["a"])
		assert.Equal(t, float64(subset[idx].B), record["b"])
		assert.Equal(t, base64.StdEncoding.EncodeToString(subset[idx].C), record["c"])
		assert.Equal(t, float64(timeutil.UnixMilli(subset[idx].D)), record["d"])
		assert.Equal(t, subset[idx].E.FloatString(0), record["e"])
		assert.Equal(t, nil, record["f"])
	}

	// query some filtered data with REST
	// expecting four records (two from each subset)
	code, res13 := queryGatewayHTTP(http.MethodGet, fmt.Sprintf(`v1/%s/test/foo-bar?filter={"a":{"_prefix":"b"}}`, testOrg.Name), nil)
	assert.Equal(t, 200, code)
	assert.Len(t, res13["meta"], 2)
	assert.Len(t, res13["data"], 4)
	for _, recordT := range res13["data"].([]interface{}) {
		record := recordT.(map[string]interface{})
		assert.NotNil(t, record)
		assert.Len(t, record["@meta"], 2)
		assert.Equal(t, byte('b'), record["a"].(string)[0])
	}

	// wait to let metrics get committed
	time.Sleep(200 * time.Millisecond)

	// check metrics have been accurately tracked for stream
	res14 := queryGQL(`
		query GetStreamUsage($streamID: UUID!, $from: Time!) {
			getStreamUsage(streamID: $streamID, period: "month", from: $from) {
				entityID
				time
				readOps
				readBytes
				readRecords
				writeOps
				writeBytes
				writeRecords
			}
		}
	`, map[string]interface{}{
		"streamID": stream["streamID"],
		"from":     timeutil.Floor(startTime, timeutil.PeriodMonth).Format("2006-01-02T15:04:05Z07:00"),
	})
	assert.Empty(t, res14.Errors)
	streamUsage := res14.Results()["getStreamUsage"]
	assert.Len(t, streamUsage, 1)
	assert.Equal(t, stream["streamID"], streamUsage[0]["entityID"])
	assert.Greater(t, int(streamUsage[0]["readOps"].(float64)), 0)
	assert.Greater(t, int(streamUsage[0]["readBytes"].(float64)), 0)
	assert.Greater(t, int(streamUsage[0]["readRecords"].(float64)), 0)
	assert.Equal(t, int(streamUsage[0]["writeOps"].(float64)), 2)
	assert.Greater(t, int(streamUsage[0]["writeBytes"].(float64)), 0)
	assert.Equal(t, int(streamUsage[0]["writeRecords"].(float64)), len(foobars))

	// check metrics have been accurately tracked for the user
	res15 := queryGQL(`
		query GetUserUsage($userID: UUID!, $from: Time!) {
			getUserUsage(userID: $userID, period: "month", from: $from) {
				entityID
				time
				readOps
				readBytes
				readRecords
				writeOps
				writeBytes
				writeRecords
			}
		}
	`, map[string]interface{}{
		"userID": testUser.UserID.String(),
		"from":   timeutil.Floor(startTime, timeutil.PeriodMonth).Format("2006-01-02T15:04:05Z07:00"),
	})
	assert.Empty(t, res15.Errors)
	userUsage := res15.Results()["getUserUsage"]
	assert.Len(t, userUsage, 1)
	assert.Equal(t, testUser.UserID.String(), userUsage[0]["entityID"])
	assert.Greater(t, int(userUsage[0]["readOps"].(float64)), 0)
	assert.Greater(t, int(userUsage[0]["readBytes"].(float64)), 0)
	assert.Greater(t, int(userUsage[0]["readRecords"].(float64)), 0)
	assert.Equal(t, int(userUsage[0]["writeOps"].(float64)), 2)
	assert.Greater(t, int(userUsage[0]["writeBytes"].(float64)), 0)
	assert.Equal(t, int(userUsage[0]["writeRecords"].(float64)), len(foobars))

	// test peek by grpc
	res16, err := gatewayGRPC.QueryLog(grpcContext(), &pb.QueryLogRequest{
		InstanceId: instanceID.Bytes(),
		Peek:       true,
		Partitions: 1,
	})
	assert.Nil(t, err)
	assert.Len(t, res16.ReplayCursors, 1)
	assert.Len(t, res16.ChangeCursors, 1)

	// read peek page 1
	res17, err := gatewayGRPC.Read(grpcContext(), &pb.ReadRequest{
		Cursor: res16.ReplayCursors[0],
		Limit:  60,
	})
	assert.Nil(t, err)
	assert.True(t, len(res17.NextCursor) > 0)
	assert.Len(t, res17.Records, 60)

	// read peek page 2
	res18, err := gatewayGRPC.Read(grpcContext(), &pb.ReadRequest{
		Cursor: res17.NextCursor,
		Limit:  60,
	})
	assert.Nil(t, err)
	assert.Len(t, res18.NextCursor, 0)
	assert.Len(t, res18.Records, 40)
}
