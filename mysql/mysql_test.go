package mysql

import (
	"testing"

	"gotest.tools/assert"
)

func TestMin(t *testing.T) {
	assert.Equal(t, Min(1, 2), 1)
	assert.Equal(t, Min(5, 2), 2)
}
