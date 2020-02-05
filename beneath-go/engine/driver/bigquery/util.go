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

func externalDatasetName(projectName string) string {
	return strings.ReplaceAll(projectName, "-", "_")
}

func externalTableName(streamName string, instanceID uuid.UUID) string {
	name := strings.ReplaceAll(streamName, "-", "_")
	return fmt.Sprintf("%s_%s", name, hex.EncodeToString(instanceID[0:4]))
}

func externalStreamViewName(streamName string) string {
	return strings.ReplaceAll(streamName, "-", "_")
}

func fullyQualifiedName(table *bigquery.Table) string {
	return strings.ReplaceAll(table.FullyQualifiedName(), ":", ".")
}
