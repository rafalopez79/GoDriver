package mysql

import (
	"testing"

	"gotest.tools/assert"
)

func TestReadFixedLenghtString(t *testing.T) {
	str := "Here is a string...."
	buff := []byte(str)
	assert.Equal(t, ReadFixedLengthString(buff, 3), "Her")
}
