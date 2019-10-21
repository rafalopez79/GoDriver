package util

import (
	"reflect"
	"unsafe"
)

//String from slice
func String(b []byte) (s string) {
	bytes := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	data := (*reflect.StringHeader)(unsafe.Pointer(&s))
	data.Data = bytes.Data
	data.Len = bytes.Len
	return
}
