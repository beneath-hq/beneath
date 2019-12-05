package bigquery

import (
	"os"
	"testing"
)

func TestBigQuery(t *testing.T) {
	os.Setenv("PROJECT_ID", "beneathcrypto")

	// b := New()
	// ctx := context.Background()

	// name := "ericsss"
	// displayName := "Eric"
	// description := "A person"

	// dataset := b.Client.Dataset(externalDatasetName(name))
	// md1, err := dataset.Metadata(ctx)
	// log.Printf("etag: %v, err: %v\n", md1, err)

	// md2, err := dataset.Update(ctx, bq.DatasetMetadataToUpdate{
	// 	Name:        displayName,
	// 	Description: description,
	// 	Access:      makeAccess(true, md1.Access),
	// }, md1.ETag)
	// log.Printf("etag: %s, err: %v\n", md2.ETag, err)

	// md3, err := dataset.Update(ctx, bq.DatasetMetadataToUpdate{
	// 	Name:        displayName,
	// 	Description: description,
	// 	Access:      makeAccess(true, md1.Access),
	// }, md1.ETag)
	// log.Printf("etag: %v, err: %v\n", md3, err)

	// projectID := uuid.FromStringOrNil("216dc245-9f29-4e73-8096-07f0ecdfe279")
	// projectName := "test-project"
	// streamID := uuid.FromStringOrNil("a556a809-5a11-4c18-922c-53d9960283d1")
	// streamName := "test-stream"
	// instanceID := uuid.FromStringOrNil("1d628bc4-0bc6-4d08-86bf-f7472c2029bd")
	// schema := `[{"mode":"REQUIRED","name":"a","type":"STRING"},{"mode":"REQUIRED","name":"b","type":"TIMESTAMP"},{"mode":"REQUIRED","name":"__key","type":"BYTES"},{"mode":"REQUIRED","name":"__timestamp","type":"TIMESTAMP"}]`
	// keyFields := []string{"a", "b"}

	// err := b.RegisterProject(projectID, true, projectName, "Test Project", "It's a test project")
	// assert.Nil(t, err)

	// err := b.RegisterStreamInstance(projectID, projectName, streamID, streamName, "It's a test stream", schema, keyFields, instanceID)
	// assert.Nil(t, err)
}
