package mysql

import (
	"testing"

	"gotest.tools/assert"
)

func TestMin(t *testing.T) {
	assert.Equal(t, Min(1, 2), 1)
	assert.Equal(t, Min(5, 2), 2)
}

func TestString(t *testing.T) {
	str := "Here is a string...."
	buff := []byte(str)
	assert.Equal(t, String(buff), str)
}

func TestReadFixedLenghtString(t *testing.T) {
	str := "Here is a string...."
	buff := []byte(str)
	assert.Equal(t, ReadFixedLengthString(buff, 3), "Her")
}
