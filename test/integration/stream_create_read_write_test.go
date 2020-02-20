package integration

import (
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestStreamCreateReadAndWrite(t *testing.T) {
	// create project
	res := queryGQL(`
		mutation CreateProject($organizationID: UUID!) {
			createProject(name: "test", organizationID: $organizationID, public: true) {
				projectID
				name
				users { userID }
			}
		}
	`, map[string]interface{}{"organizationID": testUser.OrganizationID})
	assert.Empty(t, res.Errors)
	project := res.Data["createProject"]
	assert.Equal(t, "test", project["name"])
	assert.NotEmpty(t, project["users"])

	// create stream
	schema := readTestdata("foobar.graphql")
	res = queryGQL(`
		mutation CreateExternalStream($projectID: UUID!, $schema: String!) {
			createExternalStream(projectID: $projectID, schema: $schema, batch: false, manual: false) {
				streamID
				currentStreamInstanceID
				name
				project { projectID }
			}
		}
	`, map[string]interface{}{
		"projectID": project["projectID"],
		"schema":    schema,
	})
	assert.Empty(t, res.Errors)
	stream := res.Data["createExternalStream"]
	assert.Equal(t, "foo_bar", stream["name"])
	assert.Len(t, stream["project"], 1)
	assert.False(t, uuid.FromStringOrNil(stream["currentStreamInstanceID"].(string)) == uuid.Nil)

	// create stream
	// update it
	// write to it
	// replay it with grpc
	// replay it with rest
	// query it with grpc
	// query it with rest
	// subscribe to updates with grpc
	// subscribe to updates with ws
	// 	push updates
	// 	check for updates
}
