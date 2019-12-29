package util

import (
	"testing"
)

func TestBufferPool(t *testing.T) {
	bp := NewBufferPool()
	buff := bp.Get()
	bp.Return(buff)
}
