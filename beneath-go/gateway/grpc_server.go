package gateway

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/beneath-core/beneath-go/control/entity"
	"github.com/beneath-core/beneath-go/core/jsonutil"
	"github.com/beneath-core/beneath-go/core/log"
	"github.com/beneath-core/beneath-go/core/middleware"
	"github.com/beneath-core/beneath-go/core/queryparse"
	"github.com/beneath-core/beneath-go/core/timeutil"
	"github.com/beneath-core/beneath-go/db"
	pb "github.com/beneath-core/beneath-go/proto"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	version "github.com/hashicorp/go-version"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// see https://github.com/grpc/grpc-go/blob/master/Documentation/encoding.md#using-a-compressor
	_ "google.golang.org/grpc/encoding/gzip"
)

const (
	maxRecvMsgSize = 1024 * 1024 * 10
	maxSendMsgSize = 1024 * 1024 * 50
)

type clientVersionSpec struct {
	RecommendedVersion *version.Version
	WarningVersion     *version.Version
	DeprecatedVersion  *version.Version
}

func (s clientVersionSpec) IsZero() bool {
	return s.DeprecatedVersion == nil && s.WarningVersion == nil && s.RecommendedVersion == nil
}

func newClientVersionSpec(recommended, warning, deprecated string) clientVersionSpec {
	return clientVersionSpec{
		RecommendedVersion: newVersionOrPanic(recommended),
		WarningVersion:     newVersionOrPanic(warning),
		DeprecatedVersion:  newVersionOrPanic(deprecated),
	}
}

func newVersionOrPanic(str string) *version.Version {
	v, err := version.NewVersion(str)
	if err != nil {
		panic(err)
	}
	return v
}

var clientSpecs = map[string]clientVersionSpec{
	"beneath-python": newClientVersionSpec("1.0.2", "1.0.1", "0.0.1"),
}

// ListenAndServeGRPC serves a gRPC API
func ListenAndServeGRPC(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}

	server := grpc.NewServer(
		grpc.MaxRecvMsgSize(maxRecvMsgSize),
		grpc.MaxSendMsgSize(maxSendMsgSize),
		grpc_middleware.WithUnaryServerChain(
			middleware.InjectTagsUnaryServerInterceptor(),
			middleware.LoggerUnaryServerInterceptor(),
			grpc_auth.UnaryServerInterceptor(middleware.AuthInterceptor),
			middleware.RecovererUnaryServerInterceptor(),
		),
		grpc_middleware.WithStreamServerChain(
			middleware.InjectTagsStreamServerInterceptor(),
			middleware.LoggerStreamServerInterceptor(),
			grpc_auth.StreamServerInterceptor(middleware.AuthInterceptor),
			middleware.RecovererStreamServerInterceptor(),
		),
	)
	pb.RegisterGatewayServer(server, &gRPCServer{})

	log.S.Infow("gateway grpc started", "port", port)
	return server.Serve(lis)
}

// gRPCServer implements pb.GatewayServer
type gRPCServer struct{}

func (s *gRPCServer) GetStreamDetails(ctx context.Context, req *pb.StreamDetailsRequest) (*pb.StreamDetailsResponse, error) {
	// to backend names
	req.StreamName = toBackendName(req.StreamName)
	req.ProjectName = toBackendName(req.ProjectName)

	// set log payload
	middleware.SetTagsPayload(ctx, streamDetailsLog{
		Stream:  req.StreamName,
		Project: req.ProjectName,
	})

	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// get instance ID
	instanceID := entity.FindInstanceIDByNameAndProject(ctx, req.StreamName, req.ProjectName)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// get stream details
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Read {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// return
	return &pb.StreamDetailsResponse{
		CurrentInstanceId: instanceID.Bytes(),
		ProjectId:         stream.ProjectID.Bytes(),
		ProjectName:       stream.ProjectName,
		StreamName:        stream.StreamName,
		KeyFields:         stream.Codec.GetKeyFields(),
		AvroSchema:        stream.Codec.GetAvroSchemaString(),
		Public:            stream.Public,
		External:          stream.External,
		Batch:             stream.Batch,
		Manual:            stream.Manual,
	}, nil
}

func (s *gRPCServer) ReadRecords(ctx context.Context, req *pb.ReadRecordsRequest) (*pb.ReadRecordsResponse, error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// read instanceID
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "instance_id not valid UUID")
	}

	// set log payload
	middleware.SetTagsPayload(ctx, readRecordsLog{
		InstanceID: instanceID.String(),
		Limit:      req.Limit,
		Where:      req.Where,
		After:      req.After,
	})

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.External, stream.Public)
	if !perms.Read {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// check quota
	usage := Metrics.GetCurrentUsage(ctx, secret.GetOwnerID())
	ok := secret.CheckReadQuota(usage)
	if !ok {
		return nil, status.Error(codes.ResourceExhausted, "you have exhausted your monthly quota")
	}

	// check limit is valid
	if req.Limit == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "limit cannot be 0")
	} else if req.Limit > maxRecordsLimit {
		return nil, grpc.Errorf(codes.InvalidArgument, fmt.Sprintf("limit exceeds maximum of %d", maxRecordsLimit))
	}

	// get key range where clause (if no where, it will be nil)
	var whereQuery queryparse.Query
	if req.Where != "" {
		// read where
		var where map[string]interface{}
		err := jsonutil.UnmarshalBytes([]byte(req.Where), &where)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "couldn't parse 'where' -- is it valid JSON?")
		}

		// make query
		whereQuery, err = queryparse.JSONToQuery(where)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, fmt.Sprintf("couldn't parse 'where' query: %s", err.Error()))
		}
	}

	// adapt key range based on after (for pagination), if present
	var afterQuery queryparse.Query
	if req.After != "" {
		// read after
		var after map[string]interface{}
		err := jsonutil.UnmarshalBytes([]byte(req.After), &after)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, "couldn't parse 'after' -- is it valid JSON?")
		}

		// make query
		afterQuery, err = queryparse.JSONToQuery(after)
		if err != nil {
			return nil, grpc.Errorf(codes.InvalidArgument, fmt.Sprintf("couldn't parse 'after' query: %s", err.Error()))
		}
	}

	// set key range
	keyRange, err := stream.Codec.MakeKeyRange(whereQuery, afterQuery)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// read rows from engine
	response := &pb.ReadRecordsResponse{}
	bytesRead := 0
	err = db.Engine.Tables.ReadRecordRange(ctx, instanceID, keyRange, int(req.Limit), func(avroData []byte, timestamp time.Time) error {
		response.Records = append(response.Records, &pb.Record{
			AvroData:  avroData,
			Timestamp: timeutil.UnixMilli(timestamp),
		})
		bytesRead += len(avroData)
		return nil
	})
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// track read metrics
	Metrics.TrackRead(stream.StreamID, int64(len(response.Records)), int64(bytesRead))
	Metrics.TrackRead(secret.GetOwnerID(), int64(len(response.Records)), int64(bytesRead))

	// done
	return response, nil
}

func (s *gRPCServer) ReadLatestRecords(ctx context.Context, req *pb.ReadLatestRecordsRequest) (*pb.ReadRecordsResponse, error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// read instanceID
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "instance_id not valid UUID")
	}

	// set log payload
	middleware.SetTagsPayload(ctx, readLatestLog{
		InstanceID: instanceID.String(),
		Limit:      req.Limit,
		Before:     req.Before,
	})

	// get cached stream
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Read {
		return nil, grpc.Errorf(codes.PermissionDenied, "token doesn't grant right to read this stream")
	}

	// check isn't batch
	if stream.Batch {
		return nil, grpc.Errorf(codes.InvalidArgument, "cannot get latest records for batch streams")
	}

	// check quota
	usage := Metrics.GetCurrentUsage(ctx, secret.GetOwnerID())
	ok := secret.CheckReadQuota(usage)
	if !ok {
		return nil, status.Error(codes.ResourceExhausted, "you have exhausted your monthly quota")
	}

	// check limit is valid
	if req.Limit == 0 {
		return nil, grpc.Errorf(codes.InvalidArgument, "limit cannot be 0")
	} else if req.Limit > maxRecordsLimit {
		return nil, grpc.Errorf(codes.InvalidArgument, fmt.Sprintf("limit exceeds maximum of %d", maxRecordsLimit))
	}

	// get before as time
	var before time.Time
	if req.Before != 0 {
		before = timeutil.FromUnixMilli(req.Before)
	} else {
		before = time.Time{}
	}

	// read rows from engine
	response := &pb.ReadRecordsResponse{}
	bytesRead := 0
	err := db.Engine.Tables.ReadLatestRecords(ctx, instanceID, int(req.Limit), before, func(avroData []byte, timestamp time.Time) error {
		response.Records = append(response.Records, &pb.Record{
			AvroData:  avroData,
			Timestamp: timeutil.UnixMilli(timestamp),
		})
		bytesRead += len(avroData)
		return nil
	})
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// track read metrics
	Metrics.TrackRead(stream.StreamID, int64(len(response.Records)), int64(bytesRead))
	Metrics.TrackRead(secret.GetOwnerID(), int64(len(response.Records)), int64(bytesRead))

	// done
	return response, nil
}

func (s *gRPCServer) WriteRecords(ctx context.Context, req *pb.WriteRecordsRequest) (*pb.WriteRecordsResponse, error) {
	// get auth
	secret := middleware.GetSecret(ctx)
	if secret == nil {
		return nil, grpc.Errorf(codes.PermissionDenied, "not authenticated")
	}

	// read instanceID
	instanceID := uuid.FromBytesOrNil(req.InstanceId)
	if instanceID == uuid.Nil {
		return nil, status.Error(codes.InvalidArgument, "instance_id not valid UUID")
	}

	// set log payload
	payload := writeRecordsLog{
		InstanceID:   instanceID.String(),
		RecordsCount: len(req.Records),
	}
	middleware.SetTagsPayload(ctx, payload)

	// get stream info
	stream := entity.FindCachedStreamByCurrentInstanceID(ctx, instanceID)
	if stream == nil {
		return nil, status.Error(codes.NotFound, "stream not found")
	}

	// check not already a committed batch stream
	if stream.Batch && stream.Committed {
		return nil, status.Error(codes.FailedPrecondition, "batch has been committed and closed for further writes")
	}

	// check permissions
	perms := secret.StreamPermissions(ctx, stream.StreamID, stream.ProjectID, stream.Public, stream.External)
	if !perms.Write {
		return nil, grpc.Errorf(codes.PermissionDenied, "secret doesn't grant right to write to this stream")
	}

	// check records supplied
	if len(req.Records) == 0 {
		return nil, status.Error(codes.InvalidArgument, "records cannot be empty")
	}

	// check quota
	usage := Metrics.GetCurrentUsage(ctx, secret.GetOwnerID())
	ok := secret.CheckWriteQuota(usage)
	if !ok {
		return nil, status.Error(codes.ResourceExhausted, "you have exhausted your monthly quota")
	}

	// check the batch length is valid
	err := db.Engine.CheckBatchLength(len(req.Records))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("error encoding batch: %v", err.Error()))
	}

	// check each record is valid
	bytesWritten := 0
	for idx, record := range req.Records {
		// set timestamp to current timestamp if it's 0
		if record.Timestamp == 0 {
			record.Timestamp = timeutil.UnixMilli(time.Now())
		}

		// check it decodes
		decodedData, err := stream.Codec.UnmarshalAvro(record.AvroData)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("record at index %d doesn't decode: %v", idx, err.Error()))
		}

		// compute key (only used to check size)
		keyData, err := stream.Codec.MarshalKey(decodedData)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// check size
		err = db.Engine.CheckSize(len(keyData), len(record.AvroData))
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("error encoding record at index %d: %v", idx, err.Error()))
		}

		// increment bytes written
		bytesWritten += len(record.AvroData)
	}

	// update log payload
	payload.BytesWritten = bytesWritten
	middleware.SetTagsPayload(ctx, payload)

	// write request to engine
	err = db.Engine.Streams.QueueWriteRequest(ctx, req)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, err.Error())
	}

	// track write metrics
	Metrics.TrackWrite(instanceID, int64(len(req.Records)), int64(bytesWritten))
	Metrics.TrackWrite(secret.GetOwnerID(), int64(len(req.Records)), int64(bytesWritten))

	return &pb.WriteRecordsResponse{}, nil
}

func (s *gRPCServer) SendClientPing(ctx context.Context, req *pb.ClientPing) (*pb.ClientPong, error) {
	// set log payload
	middleware.SetTagsPayload(ctx, clientPingLog{
		ClientID:      req.ClientId,
		ClientVersion: req.ClientVersion,
	})

	spec := clientSpecs[req.ClientId]
	if spec.IsZero() {
		return nil, grpc.Errorf(codes.InvalidArgument, "unrecognized client ID")
	}

	v, err := version.NewVersion(req.ClientVersion)
	if err != nil {
		return nil, grpc.Errorf(codes.InvalidArgument, "client version is not a valid semver")
	}

	status := ""
	if v.GreaterThanOrEqual(spec.RecommendedVersion) {
		status = "stable"
	} else if v.GreaterThanOrEqual(spec.WarningVersion) {
		status = "warning"
	} else {
		status = "deprecated"
	}

	secret := middleware.GetSecret(ctx)
	return &pb.ClientPong{
		Authenticated:      !secret.IsAnonymous(),
		Status:             status,
		RecommendedVersion: spec.RecommendedVersion.String(),
	}, nil
}
