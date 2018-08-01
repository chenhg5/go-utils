package utils

import (
	"unsafe"
	"reflect"
)

func ByteSliceToString(bs []byte) string {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	strHeader := &reflect.StringHeader{
		Data: header.Data,
		Len: header.Len,
	}
	return *(*string)(unsafe.Pointer(strHeader))
	//return *(*string)(unsafe.Pointer(&bs))
}

func StringToByteSlice(s string) []byte {
	header := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bytesHeader := &reflect.SliceHeader{
		Data: header.Data,
		Len: header.Len,
		Cap: header.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bytesHeader))
}