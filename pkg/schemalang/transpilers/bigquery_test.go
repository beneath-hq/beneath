package transpilers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/beneath-hq/beneath/pkg/schemalang"
)

func TestBigQuery(t *testing.T) {
	s, err := FromAvro(UserSchemaAvroWithDoc)
	assert.Nil(t, err)

	// bigquery transpilation is lossy, so we do an extra round-trip before getting avro for comparison
	// note: doesn't detect if ToBigQuery and FromBigQuery implements the same bug commutatively

	// lossy
	bq := ToBigQuery(s, true)
	s, err = FromBigQuery(bq)
	assert.Nil(t, err)

	err = schemalang.Check(s)
	assert.Nil(t, err)

	avroExpected := ToAvro(s, true)

	// non-lossy
	bq = ToBigQuery(s, true)
	s, err = FromBigQuery(bq)
	assert.Nil(t, err)

	err = schemalang.Check(s)
	assert.Nil(t, err)

	avroActual := ToAvro(s, true)

	assert.Equal(t, avroExpected, avroActual)
}
