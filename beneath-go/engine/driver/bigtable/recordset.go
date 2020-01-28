package bigtable

import (
	"time"

	"cloud.google.com/go/bigtable"
)

// map[string]internalRecord
// type RecordSet struct {

// }

// func LoadExistingRecords([]driver.Records) -> map
// func LoadRecordsFromSecondaryKeys() -> list
// (internally a map)
// func LoadRecordsFromPrimaryKeys()
// func LoadRecordsFromOffset(offset, limit) -> list
// (internally a map)

// internalRecord implements driver.Record
type internalRecord struct {
	Offset       int64
	AvroData     []byte
	PrimaryKey   []byte
	SecondaryKey []byte
	Timestamp    bigtable.Timestamp
}

func (r internalRecord) IsNil() bool {
	return r.Offset == 0 && len(r.AvroData) == 0 && len(r.PrimaryKey) == 0 && len(r.SecondaryKey) == 0 && r.Timestamp == 0
}

func (r internalRecord) GetTimestamp() time.Time {
	return r.Timestamp.Time()
}

func (r internalRecord) GetAvro() []byte {
	return r.AvroData
}

func (r internalRecord) GetStructured() map[string]interface{} {
	return nil
	// structured, err := r.Stream.GetCodec().UnmarshalAvro(r.AvroData)
	// if err != nil {
	// 	panic(err)
	// }

	// structured, err = stream.GetCodec().ConvertFromAvroNative(structured, false)
	// if err != nil {
	// 	panic(err)
	// }

	// return record{
	// 	Proto:      proto,
	// 	Structured: structured,
	// }
}
