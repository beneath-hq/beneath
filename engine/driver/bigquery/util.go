package bigquery

import (
	"encoding/hex"
	"fmt"
	"strings"

	"cloud.google.com/go/bigquery"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.com/beneath-hq/beneath/engine/driver"
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

func externalDatasetName(p driver.Project) string {
	return strings.ReplaceAll(p.GetOrganizationName()+"__"+p.GetProjectName(), "-", "_")
}

func externalTableName(streamName string, instanceID uuid.UUID) string {
	name := strings.ReplaceAll(streamName, "-", "_")
	return fmt.Sprintf("%s_%s", name, hex.EncodeToString(instanceID[:]))
}

func parseExternalTableName(tableName string) (string, uuid.UUID, error) {
	if len(tableName) < 2*uuid.Size+2 {
		return "", uuid.Nil, fmt.Errorf("bad table name: %s", tableName)
	}

	idx := len(tableName) - 2*uuid.Size

	idBytes, err := hex.DecodeString(tableName[idx:])
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("bad table name (corrupt instance id): %s", tableName)
	}

	instanceID, err := uuid.FromBytes(idBytes)
	if err != nil {
		return "", uuid.Nil, fmt.Errorf("bad table name (corrupt instance id): %s", tableName)
	}

	return tableName[:(idx - 1)], instanceID, nil
}

func externalStreamViewName(streamName string) string {
	return strings.ReplaceAll(streamName, "-", "_")
}

func fullyQualifiedName(table *bigquery.Table) string {
	return strings.ReplaceAll(table.FullyQualifiedName(), ":", ".")
}
