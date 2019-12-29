package util

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"sync"
	"unsafe"
)

const (
	oneByte   byte = 0XFB
	twoByte   byte = 0XFC
	threeByte byte = 0XFD
	eightByte byte = 0XFE
)

//String from slice
func String(b []byte) (s string) {
	bytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	data := (*reflect.StringHeader)(unsafe.Pointer(&s))
	data.Data = bytes.Data
	data.Len = bytes.Len
	return
}

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

//util bytes

//WriteNullTerminatedString append the nts to the buffer
func WriteNullTerminatedString(buffer *bytes.Buffer, s string) (err error) {
	_, err = buffer.WriteString(s)
	if err != nil {
		return err
	}
	return buffer.WriteByte(0)
}

//WriteFixedLengthString append the nts to the buffer
func WriteFixedLengthString(buffer *bytes.Buffer, s string) (err error) {
	_, err = buffer.WriteString(s)
	return err
}

//WriteRLEString append the nts to the buffer
func WriteRLEString(buffer *bytes.Buffer, s string) (err error) {
	err = WriteRLEInt(buffer, uint64(len(s)))
	if err != nil {
		return err
	}
	_, err = buffer.WriteString(s)
	return err
}

//WriteBytes writes byte to buffer
func WriteBytes(buffer *bytes.Buffer, bytes ...byte) (err error) {
	for _, b := range bytes {
		err = buffer.WriteByte(b)
		if err != nil {
			return err
		}
	}
	return nil
}

//Write writes byte to buffer
func Write(buffer *bytes.Buffer, bytes []byte) (err error) {
	n, err := buffer.Write(bytes)
	if err != nil {
		return err
	} else if n != len(bytes) {
		return fmt.Errorf("Write failed. only %v bytes written while %v expected", n, len(bytes))
	}
	return nil
}

//WriteInt4 writes int4 to buffer
func WriteInt4(buffer *bytes.Buffer, b uint32) (err error) {
	return WriteBytes(buffer, byte(b), byte(b>>8), byte(b>>16), byte(b>>24))
}

//WriteInt3 writes int3 to buffer
func WriteInt3(buffer *bytes.Buffer, b uint32) (err error) {
	return WriteBytes(buffer, byte(b), byte(b>>8), byte(b>>16))
}

//WriteInt2 writes int2 to buffer
func WriteInt2(buffer *bytes.Buffer, b uint16) (err error) {
	return WriteBytes(buffer, byte(b), byte(b>>8))
}

//Read read byte from buffer
func Read(buffer *bytes.Buffer, bytes []byte) (err error) {
	n, err := buffer.Read(bytes)
	if err != nil {
		return err
	} else if n != len(bytes) {
		return fmt.Errorf("Read failed. only %v bytes read while %v expected", n, len(bytes))
	}
	return nil
}

//ReadInt4 writes int4 to buffer
func ReadInt4(buffer *bytes.Buffer) (b uint32, err error) {
	var data [4]byte
	err = Read(buffer, data[:])
	if err != nil {
		return 0, err
	}
	b = uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16 + uint32(data[3])<<24
	return b, nil
}

//ReadInt3 reads int3 to buffer
func ReadInt3(buffer *bytes.Buffer) (b uint32, err error) {
	var data [3]byte
	err = Read(buffer, data[:])
	if err != nil {
		return 0, err
	}
	b = uint32(data[0]) + uint32(data[1])<<8 + uint32(data[2])<<16
	return b, nil
}

//ReadInt2 reads int2 to buffer
func ReadInt2(buffer *bytes.Buffer) (b uint16, err error) {
	var data [2]byte
	err = Read(buffer, data[:])
	if err != nil {
		return 0, err
	}
	b = uint16(data[0]) + uint16(data[1])<<8
	return b, nil
}

//WriteRLEInt writes int to buffer
func WriteRLEInt(buffer *bytes.Buffer, n uint64) (err error) {
	switch {
	case n <= 250:
		return buffer.WriteByte(byte(n))
	case n <= 0xffff:
		b0 := byte(n)
		b1 := byte(n >> 8)
		return WriteBytes(buffer, twoByte, b0, b1)
	case n <= 0xffffff:
		b0 := byte(n)
		b1 := byte(n >> 8)
		b2 := byte(n >> 16)
		return WriteBytes(buffer, threeByte, b0, b1, b2)
	case n <= 0xffffffffffffffff:
		b0 := byte(n)
		b1 := byte(n >> 8)
		b2 := byte(n >> 16)
		b3 := byte(n >> 24)
		b4 := byte(n >> 32)
		b5 := byte(n >> 40)
		b6 := byte(n >> 48)
		b7 := byte(n >> 56)
		return WriteBytes(buffer, eightByte, b0, b1, b2, b3, b4, b5, b6, b7)
	}
	return fmt.Errorf("Wrong uint64: %v", n)
}

//WriteRLEIntNUL writes bull int to buffer
func WriteRLEIntNUL(buffer *bytes.Buffer) error {
	return buffer.WriteByte(oneByte)
}

//ReadRLEInt reads int to buffer
func ReadRLEInt(buffer *bytes.Buffer) (num int64, null bool, err error) {
	var mark byte
	mark, err = buffer.ReadByte()
	if mark < oneByte {
		num := int64(mark)
		return num, false, err
	} else if mark == oneByte {
		return 0, true, nil
	} else if mark == twoByte {
		b1, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b2, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		return int64(b1) + int64(b2)<<8, false, nil
	} else if mark == threeByte {
		b1, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b2, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b3, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		return int64(b1) + int64(b2)<<8 + int64(b3)<<16, false, nil
	} else if mark == eightByte {
		b1, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b2, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b3, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b4, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b5, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b6, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		b7, err := buffer.ReadByte()
		if err != nil {
			return 0, false, err
		}
		return int64(b1) + int64(b2)<<8 + int64(b3)<<16 + int64(b4)<<24 + int64(b4)<<32 + int64(b5)<<40 + int64(b6)<<48 + int64(b7)<<56, false, nil
	}
	return 0, false, fmt.Errorf("Wrong RLE Integer")
}

//ReadFixedLengthString reads string from byte[]
func ReadFixedLengthString(buff []byte, l int) string {
	return String(buff[:Min(len(buff), l)])
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
