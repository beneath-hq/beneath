package bigquery

import (
	"encoding/hex"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
