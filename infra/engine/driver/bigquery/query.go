package bigquery

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/gogo/protobuf/proto"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/iterator"

	"github.com/beneath-hq/beneath/infra/engine/driver"
	pb "github.com/beneath-hq/beneath/infra/engine/driver/bigquery/proto"
	"github.com/beneath-hq/beneath/pkg/codec"
	"github.com/beneath-hq/beneath/pkg/schemalang"
	"github.com/beneath-hq/beneath/pkg/schemalang/transpilers"
)

const (
	maxMaxBytesScanned = 1000000000000 // 1 TB
)

// GetWarehouseTableName implements beneath.WarehouseService
func (b BigQuery) GetWarehouseTableName(p driver.Project, s driver.Table, i driver.TableInstance) string {
	return fmt.Sprintf("`%s.%s`", b.InstancesDataset.DatasetID, instanceTableName(i.GetTableInstanceID()))
}

// AnalyzeWarehouseQuery implements beneath.WarehouseService
func (b BigQuery) AnalyzeWarehouseQuery(ctx context.Context, query string) (driver.WarehouseJob, error) {
	q := b.Client.Query(query)
	q.AllowLargeResults = false
	q.DryRun = true

	bqJob, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}

	return b.bqJobToDriverJob(ctx, bqJob, true)
}

// RunWarehouseQuery implements beneath.WarehouseService
func (b BigQuery) RunWarehouseQuery(ctx context.Context, jobID uuid.UUID, query string, partitions int, timeoutMs int, maxBytesScanned int) (driver.WarehouseJob, error) {
	if maxBytesScanned > maxMaxBytesScanned {
		return nil, fmt.Errorf("max_bytes_scanned=%d exceeds maximum of %d", maxBytesScanned, maxMaxBytesScanned)
	} else if maxBytesScanned <= 0 {
		maxBytesScanned = maxMaxBytesScanned
	}

	q := b.Client.Query(query)
	q.AllowLargeResults = false
	q.MaxBytesBilled = int64(maxBytesScanned)
	q.JobID = jobID.String()
	// TODO: handle timeoutMs

	job, err := q.Run(ctx)
	if err != nil {
		return nil, err
	}

	return b.bqJobToDriverJob(ctx, job, false)
}

// PollWarehouseJob implements beneath.WarehouseService
func (b BigQuery) PollWarehouseJob(ctx context.Context, jobID uuid.UUID) (driver.WarehouseJob, error) {
	job, err := b.Client.JobFromID(ctx, jobID.String())
	if err != nil {
		return nil, err
	}
	return b.bqJobToDriverJob(ctx, job, false)
}

func (b BigQuery) bqJobToDriverJob(ctx context.Context, bqJob *bigquery.Job, dry bool) (*queryJob, error) {
	// get status
	status := bqJob.LastStatus()
	queryStats, ok := bqJob.LastStatus().Statistics.Details.(*bigquery.QueryStatistics)
	if !ok {
		return nil, fmt.Errorf("Expected QueryStatistics from query job")
	}

	// check it's a select
	if queryStats.StatementType != "SELECT" {
		return nil, fmt.Errorf("Job is not a SELECT query")
	}

	// build job
	job := &queryJob{}

	// set job id
	if !dry {
		jobID, err := uuid.FromString(bqJob.ID())
		if err != nil {
			return nil, fmt.Errorf("Bad job id, couldn't parse as UUID")
		}
		job.JobID = jobID
	}

	// set state
	switch status.State {
	case bigquery.Pending:
		job.Status = driver.PendingWarehouseJobStatus
	case bigquery.Running:
		job.Status = driver.RunningWarehouseJobStatus
	case bigquery.Done:
		job.Status = driver.DoneWarehouseJobStatus
	default:
		panic(fmt.Errorf("unhandled job state '%v'", status.State))
	}

	// return if error
	if status.Err() != nil {
		job.Error = status.Err()
		return job, nil
	}

	// set bytes scanned
	if queryStats.TotalBytesProcessed > queryStats.TotalBytesBilled {
		job.BytesScanned = queryStats.TotalBytesProcessed
	} else {
		job.BytesScanned = queryStats.TotalBytesBilled
	}

	// extract referenced instances, if possible (ReferencedTables is only set for dry jobs)
	if queryStats.ReferencedTables != nil {
		job.ReferencedInstances = make([]driver.TableInstance, len(queryStats.ReferencedTables))

		for idx, table := range queryStats.ReferencedTables {
			if table.ProjectID != b.ProjectID {
				return nil, fmt.Errorf("Query references bad table: %s.%s.%s", table.ProjectID, table.DatasetID, table.TableID)
			}

			instanceID, err := parseInstanceTableName(table.TableID)
			if err != nil {
				return nil, fmt.Errorf("Query references bad table: %s.%s.%s (%s)", table.ProjectID, table.DatasetID, table.TableID, err.Error())
			}

			job.ReferencedInstances[idx] = tableInstance{InstanceID: instanceID}
		}
	}

	// extract avro schema, if possible (queryStats.Schema is only set for dry jobs)
	if queryStats.Schema != nil {
		avro, err := b.convertToAvroSchema(queryStats.Schema, true, true)
		if err != nil {
			return nil, err
		}
		job.ResultAvroSchema = avro
	}

	// stop now unless it's a completed non-dry job
	if dry || job.Status != driver.DoneWarehouseJobStatus {
		return job, nil
	}

	// the job is done and we should return cursors

	// query config contains the result table
	conf, err := bqJob.Config()
	if err != nil {
		return nil, fmt.Errorf("internal error: error getting job config: %s", err.Error())
	}
	queryConf, ok := conf.(*bigquery.QueryConfig)
	if !ok {
		return nil, fmt.Errorf("internal error: job is not a query")
	}

	// load info about the result table
	resultTable, err := queryConf.Dst.Metadata(ctx)
	if err != nil {
		return nil, fmt.Errorf("internal error: error getting metadata: %s", err.Error())
	}

	// set result stats
	job.ResultSizeBytes = int64(resultTable.NumBytes)
	job.ResultSizeRecords = int64(resultTable.NumRows)

	// set result avro schema
	resultAvro, err := b.convertToAvroSchema(resultTable.Schema, true, true)
	if err != nil {
		return nil, err
	}
	job.ResultAvroSchema = resultAvro

	// get cursor avro schema (includes internal fields that'll be trimmed during read)
	cursorAvro, err := b.convertToAvroSchema(resultTable.Schema, false, false)
	if err != nil {
		return nil, err
	}

	// set cursors
	cursor := &pb.Cursor{
		Dataset:    queryConf.Dst.DatasetID,
		Table:      queryConf.Dst.TableID,
		AvroSchema: cursorAvro,
	}
	compiled, err := proto.Marshal(cursor)
	if err != nil {
		panic(err)
	}
	job.ReplayCursors = [][]byte{compiled}

	return job, nil
}

func (b BigQuery) convertToAvroSchema(bqSchema bigquery.Schema, checkSchema bool, trimInternalFields bool) (string, error) {
	// trim internal fields if present
	trimmed := bqSchema
	if trimInternalFields {
		trimmed = make(bigquery.Schema, 0, len(bqSchema))
		for _, field := range bqSchema {
			if len(field.Name) < 2 || field.Name[0:2] != "__" {
				trimmed = append(trimmed, field)
			}
		}
	}

	// transpile and check
	schema, err := transpilers.FromBigQuery(trimmed)
	if err != nil {
		return "", err
	}

	if checkSchema {
		err := schemalang.Check(schema)
		if err != nil {
			return "", err
		}
	}

	return transpilers.ToAvro(schema, false), nil
}

// ReadWarehouseCursor implements beneath.WarehouseService
func (b BigQuery) ReadWarehouseCursor(ctx context.Context, cursor []byte, limit int) (driver.RecordsIterator, error) {
	// parse cursor
	parsed := &pb.Cursor{}
	err := proto.Unmarshal(cursor, parsed)
	if err != nil {
		return nil, fmt.Errorf("corrupt cursor: %s", err.Error())
	}

	// extract schema from cursor
	schema, err := transpilers.FromAvro(parsed.AvroSchema)
	if err != nil {
		return nil, fmt.Errorf("corrupt schema in cursor: %s", err.Error())
	}
	bqSchema := transpilers.ToBigQuery(schema, false)
	if bqSchema == nil {
		return nil, fmt.Errorf("corrupt schema in cursor: %s", parsed.AvroSchema)
	}

	// cursor schema has internal fields, so we (re)compute result schema without internal fields for encoding output
	outAvroSchema, err := b.convertToAvroSchema(bqSchema, false, true)
	if err != nil {
		return nil, fmt.Errorf("corrupt schema in cursor, error: %s", err.Error())
	}

	// prepare iterator
	table := b.Client.Dataset(parsed.Dataset).Table(parsed.Table)
	it := table.Read(ctx)
	it.Schema = bqSchema // setting this prevents the library from making an extra network request to get the schema
	it.PageInfo().Token = parsed.Token
	it.PageInfo().MaxSize = limit

	// create codec
	coder, err := codec.New(outAvroSchema, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("corrupt schema, error creating codec: %s", err.Error())
	}

	// get records
	records := make([]driver.Record, 0, limit)
	for {
		// get next row
		var dst map[string]bigquery.Value
		err := it.Next(&dst)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("corrupt cursor: invalid token: %s", err.Error())
		}

		// get timestamp
		ts := time.Now()
		if t, ok := dst["__timestamp"].(time.Time); ok {
			ts = t
		}

		// remove internal fields
		for fieldName := range dst {
			if len(fieldName) >= 2 && fieldName[0:2] == "__" {
				delete(dst, fieldName)
			}
		}

		// save
		records = append(records, resultRecord{
			row:   dst,
			ts:    ts,
			coder: coder,
		})

		if it.PageInfo().Remaining() == 0 {
			break
		}
	}

	// build next cursor
	nextToken := it.PageInfo().Token
	var nextCursor []byte
	if nextToken != "" {
		next := &pb.Cursor{
			Dataset:    parsed.Dataset,
			Table:      parsed.Table,
			Token:      nextToken,
			AvroSchema: parsed.AvroSchema,
		}
		compiled, err := proto.Marshal(next)
		if err != nil {
			panic(err)
		}
		nextCursor = compiled
	}

	// done
	return &recordsIterator{
		records:    records,
		nextCursor: nextCursor,
	}, nil

}

func bqValueToInterface(val bigquery.Value) interface{} {
	switch t := val.(type) {
	case map[string]bigquery.Value:
		res := make(map[string]interface{}, len(t))
		for k, v := range t {
			res[k] = bqValueToInterface(v)
		}
		return res
	case []bigquery.Value:
		res := make([]interface{}, len(t))
		for idx, v := range t {
			res[idx] = bqValueToInterface(v)
		}
		return res
	default:
		return interface{}(t)
	}
}

type resultRecord struct {
	row   map[string]bigquery.Value
	ts    time.Time
	coder *codec.Codec
}

func (r resultRecord) GetTimestamp() time.Time {
	return r.ts
}

func (r resultRecord) GetAvro() []byte {
	val := bqValueToInterface(r.row).(map[string]interface{})
	avro, err := r.coder.MarshalAvro(val)
	if err != nil {
		panic(fmt.Errorf("error converting bigquery result: %s", err.Error()))
	}
	return avro
}

func (r resultRecord) GetStructured() map[string]interface{} {
	return bqValueToInterface(r.row).(map[string]interface{})
}

func (r resultRecord) GetJSON() map[string]interface{} {
	data, err := r.coder.ConvertToJSONTypes(r.GetStructured())
	if err != nil {
		panic(err)
	}
	return data
}

func (r resultRecord) GetPrimaryKey() []byte {
	panic(fmt.Errorf("warehouse query result doesn't have primary key"))
}

type tableInstance struct {
	InstanceID uuid.UUID
}

func (i tableInstance) GetTableInstanceID() uuid.UUID {
	return i.InstanceID
}

type queryJob struct {
	JobID               uuid.UUID
	Status              driver.WarehouseJobStatus
	Error               error
	ResultAvroSchema    string
	ReplayCursors       [][]byte
	ReferencedInstances []driver.TableInstance
	BytesScanned        int64
	ResultSizeBytes     int64
	ResultSizeRecords   int64
}

// GetJobID implements driver.WarehouseJob
func (j *queryJob) GetJobID() uuid.UUID {
	return j.JobID
}

// GetStatus implements driver.WarehouseJob
func (j *queryJob) GetStatus() driver.WarehouseJobStatus {
	return j.Status
}

// GetError implements driver.WarehouseJob
func (j *queryJob) GetError() error {
	return j.Error
}

// GetResultAvroSchema implements driver.WarehouseJob
func (j *queryJob) GetResultAvroSchema() string {
	return j.ResultAvroSchema
}

// GetReplayCursors implements driver.WarehouseJob
func (j *queryJob) GetReplayCursors() [][]byte {
	return j.ReplayCursors
}

// GetReferencedInstances implements driver.WarehouseJob
func (j *queryJob) GetReferencedInstances() []driver.TableInstance {
	return j.ReferencedInstances
}

// GetBytesScanned implements driver.WarehouseJob
func (j *queryJob) GetBytesScanned() int64 {
	return j.BytesScanned
}

// GetResultSizeBytes implements driver.WarehouseJob
func (j *queryJob) GetResultSizeBytes() int64 {
	return j.ResultSizeBytes
}

// GetResultSizeRecords implements driver.WarehouseJob
func (j *queryJob) GetResultSizeRecords() int64 {
	return j.ResultSizeRecords
}

// recordsIterator implements driver.RecordsIterator
type recordsIterator struct {
	idx        int
	records    []driver.Record
	nextCursor []byte
}

// Next implements driver.RecordsIterator
func (i *recordsIterator) Next() bool {
	i.idx++
	return len(i.records) >= i.idx
}

// Record implements driver.RecordsIterator
func (i *recordsIterator) Record() driver.Record {
	if i.idx == 0 || i.idx > len(i.records) {
		panic("invalid call to Record")
	}
	return i.records[i.idx-1]
}

// NextCursor implements driver.RecordsIterator
func (i *recordsIterator) NextCursor() []byte {
	return i.nextCursor
}
