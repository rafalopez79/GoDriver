package util

import (
	"testing"

	"gotest.tools/assert"
)

func TestString(t *testing.T) {
	str := "Here is a string...."
	buff := []byte(str)
	assert.Equal(t, String(buff), str)
}

func TestBufferPool(t *testing.T) {
	bp := NewBufferPool()
	buff := bp.Get()
	bp.Return(buff)
}

func TestReadFixedLenghtString(t *testing.T) {
	str := "Here is a string...."
	buff := []byte(str)
	assert.Equal(t, ReadFixedLengthString(buff, 3), "Her")
}
