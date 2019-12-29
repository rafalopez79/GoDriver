package util

import (
	"bytes"
	"math/rand"
	"sync"
)

//BufferPool pool of buffers
type BufferPool struct {
	pool *sync.Pool
}

//NewBufferPool creates a bufferpool
func NewBufferPool() *BufferPool {
	return &BufferPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return new(bytes.Buffer)
			},
		},
	}
}

//Get buffer from pool
func (bufferPool *BufferPool) Get() *bytes.Buffer {
	return bufferPool.pool.Get().(*bytes.Buffer)
}

//Return buffer from pool
func (bufferPool *BufferPool) Return(buffer *bytes.Buffer) {
	buffer.Reset()
	bufferPool.pool.Put(buffer)
}

//RandomBuffer of rand & size
func RandomBuffer(r *rand.Rand, size int) ([]byte, error) {
	buff := make([]byte, size)
	if _, err := r.Read(buff); err != nil {
		return nil, err
	}
	for i, b := range buff {
		if uint8(b) == 0 {
			buff[i] = 'R'
		}
	}
	return buff, nil
}
