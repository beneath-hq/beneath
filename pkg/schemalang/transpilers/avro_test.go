package transpilers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/beneath-hq/beneath/pkg/schemalang"
)

func TestAvro(t *testing.T) {
	s, err := FromAvro(UserSchemaAvroWithDoc)
	assert.Nil(t, err)

	err = schemalang.Check(s)
	assert.Nil(t, err)

	avro := ToAvro(s, false)
	assert.Equal(t, UserSchemaAvro, avro)

	avroWithDoc := ToAvro(s, true)
	assert.Equal(t, UserSchemaAvroWithDoc, avroWithDoc)
}
