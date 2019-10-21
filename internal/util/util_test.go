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
