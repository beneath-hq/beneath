package stream

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	uuid "github.com/satori/go.uuid"
	"gitlab.com/beneath-hq/beneath/models"
	"gitlab.com/beneath-hq/beneath/pkg/schemalang"
	"gitlab.com/beneath-hq/beneath/pkg/schemalang/transpilers"
)

// ComputeSchemaMD5 returns a non-canonical (but useful) digest of a schema and its indexes
func (s *Service) ComputeSchemaMD5(schema string, indexes *string) []byte {
	input := []byte(strings.TrimSpace(schema))
	if indexes != nil {
		input = append(input, []byte(strings.TrimSpace(*indexes))...)
	}
	digest := md5.Sum(input)
	return digest[:]
}

// CompileToStream compiles the given schema and sets relevant fields on the stream.
func (s *Service) CompileToStream(stream *models.Stream, schemaKind models.StreamSchemaKind, newSchema string, newIndexes *string, description *string) error {
	// determine if we're updating an existing stream or creating a new
	update := stream.StreamID != uuid.Nil

	// get schema
	var schema schemalang.Schema
	var indexes schemalang.Indexes
	var err error
	switch schemaKind {
	case models.StreamSchemaKindAvro:
		schema, err = transpilers.FromAvro(newSchema)
	case models.StreamSchemaKindGraphQL:
		schema, indexes, err = transpilers.FromGraphQL(newSchema)
	default:
		err = fmt.Errorf("Unsupported stream schema definition language '%s'", schemaKind)
	}
	if err != nil {
		return err
	}

	// parse newIndexes if set (overrides output from transpiler)
	if newIndexes != nil {
		err = json.Unmarshal([]byte(*newIndexes), &indexes)
		if err != nil {
			return fmt.Errorf("error parsing indexes: '%s'", err.Error())
		}
	}

	// check schema
	err = schemalang.Check(schema)
	if err != nil {
		return err
	}

	// check indexes
	err = indexes.Check(schema)
	if err != nil {
		return err
	}

	// normalize indexes
	canonicalIndexes := indexes.CanonicalJSON()

	// get avro schemas
	canonicalAvro := transpilers.ToAvro(schema, false)
	avro := transpilers.ToAvro(schema, true)

	// if update, check canonical avro is the same
	if update {
		if canonicalAvro != stream.CanonicalAvroSchema {
			return fmt.Errorf("Schema error: Unfortunately we do not currently support structural changes to a stream's schema; you can only comments and descriptions")
		}
		if canonicalIndexes != stream.CanonicalIndexes {
			return fmt.Errorf("Schema error: Unfortunately we do not currently support updating a stream's indexes")
		}
	}

	// TEMPORARY: disable secondary indexes
	if len(indexes) != 1 {
		return fmt.Errorf("Cannot add secondary indexes to stream (a stream must have exactly one key index)")
	}

	// set indexes
	if !update {
		s.assignIndexes(stream, indexes)
	}

	// set missing stream fields
	stream.SchemaKind = schemaKind
	stream.Schema = newSchema
	stream.AvroSchema = avro
	stream.CanonicalAvroSchema = canonicalAvro
	stream.CanonicalIndexes = canonicalIndexes
	if description == nil {
		stream.Description = schema.(*schemalang.Record).Doc
	} else {
		stream.Description = *description
	}

	return nil
}

// Sets StreamIndexes to new StreamIndex objects based on def.
// Doesn't execute any DB actions, so doesn't set any IDs.
func (s *Service) assignIndexes(stream *models.Stream, indexes schemalang.Indexes) {
	indexes.Sort()
	stream.StreamIndexes = make([]*models.StreamIndex, len(indexes))
	for idx, index := range indexes {
		stream.StreamIndexes[idx] = &models.StreamIndex{
			ShortID:   idx,
			Fields:    index.Fields,
			Primary:   index.Key,
			Normalize: index.Normalize,
		}
	}
}
