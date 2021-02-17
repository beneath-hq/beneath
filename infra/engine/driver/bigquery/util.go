package bigquery

import (
	"encoding/hex"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/pkg/codec"
	"gitlab.com/beneath-hq/beneath/pkg/schemalang"
)

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	status, _ := status.FromError(err)
	return status.Code() == codes.AlreadyExists || strings.Contains(err.Error(), "Error 409: Already Exists:")
}

func isExpiredETag(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Error 400: Precondition check failed")
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "Error 404: Not found")
}

func instanceTableName(instanceID uuid.UUID) string {
	return fmt.Sprintf("%s", hex.EncodeToString(instanceID[:]))
}

func parseInstanceTableName(tableName string) (uuid.UUID, error) {
	if len(tableName) < 2*uuid.Size {
		return uuid.Nil, fmt.Errorf("bad table name: %s", tableName)
	}

	idBytes, err := hex.DecodeString(tableName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("bad table name (corrupt instance id): %s", tableName)
	}

	instanceID, err := uuid.FromBytes(idBytes)
	if err != nil {
		return uuid.Nil, fmt.Errorf("bad table name (corrupt instance id): %s", tableName)
	}

	return instanceID, nil
}

func fullyQualifiedName(table *bigquery.Table) string {
	return strings.ReplaceAll(table.FullyQualifiedName(), ":", ".")
}

// BigQuery has some limitations for clustering (see https://cloud.google.com/bigquery/docs/creating-clustered-tables#limitations):
// - it clusters on a maximum of 4 columns
// - it does not cluster on bytes types
func computeFieldsForClustering(codec *codec.Codec) []string {
	recordSchema := codec.Schema.(*schemalang.Record)
	indexFields := codec.PrimaryIndex.GetFields()
	var fieldsForClustering []string
	maxClusterFields := 4

	for i, indexField := range indexFields {
		if i == maxClusterFields {
			break
		}
		var indexFieldType schemalang.Type
		for _, field := range recordSchema.Fields {
			if field.Name == indexField {
				indexFieldType = field.Type.GetType()
				break
			}
		}
		if indexFieldType == schemalang.BytesType || indexFieldType == schemalang.FixedType {
			break
		}
		fieldsForClustering = append(fieldsForClustering, indexField)
	}

	return fieldsForClustering
}
