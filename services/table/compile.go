package table

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/beneath-hq/beneath/models"
	"github.com/beneath-hq/beneath/pkg/schemalang"
	"github.com/beneath-hq/beneath/pkg/schemalang/transpilers"
	uuid "github.com/satori/go.uuid"
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

// CompileToTable compiles the given schema and sets relevant fields on the table.
func (s *Service) CompileToTable(table *models.Table, schemaKind models.TableSchemaKind, newSchema string, newIndexes *string, description *string) error {
	// determine if we're updating an existing table or creating a new
	update := table.TableID != uuid.Nil

	// get schema
	var schema schemalang.Schema
	var indexes schemalang.Indexes
	var err error
	switch schemaKind {
	case models.TableSchemaKindAvro:
		schema, err = transpilers.FromAvro(newSchema)
	case models.TableSchemaKindGraphQL:
		schema, indexes, err = transpilers.FromGraphQL(newSchema)
	default:
		err = fmt.Errorf("Unsupported table schema definition language '%s'", schemaKind)
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
		if canonicalAvro != table.CanonicalAvroSchema {
			return fmt.Errorf("Schema error: Unfortunately we do not currently support structural changes to a table's schema; you can only comments and descriptions")
		}
		if canonicalIndexes != table.CanonicalIndexes {
			return fmt.Errorf("Schema error: Unfortunately we do not currently support updating a table's indexes")
		}
	}

	// TEMPORARY: disable secondary indexes
	if len(indexes) != 1 {
		return fmt.Errorf("Cannot add secondary indexes to table (a table must have exactly one key index)")
	}

	// set indexes
	if !update {
		s.assignIndexes(table, indexes)
	}

	// set missing table fields
	table.SchemaKind = schemaKind
	table.Schema = newSchema
	table.AvroSchema = avro
	table.CanonicalAvroSchema = canonicalAvro
	table.CanonicalIndexes = canonicalIndexes
	if description == nil {
		table.Description = schema.(*schemalang.Record).Doc
	} else {
		table.Description = *description
	}

	return nil
}

// Sets TableIndexes to new TableIndex objects based on def.
// Doesn't execute any DB actions, so doesn't set any IDs.
func (s *Service) assignIndexes(table *models.Table, indexes schemalang.Indexes) {
	indexes.Sort()
	table.TableIndexes = make([]*models.TableIndex, len(indexes))
	for idx, index := range indexes {
		table.TableIndexes[idx] = &models.TableIndex{
			ShortID:   idx,
			Fields:    index.Fields,
			Primary:   index.Key,
			Normalize: index.Normalize,
		}
	}
}
