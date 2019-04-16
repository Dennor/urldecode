package urldecode

import (
	"reflect"
	"strconv"
	"unsafe"
)

var (
	integerPrefix = []byte("i:")
	integerSuffix = []byte(";")
)

func isInteger(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func isUnsigned(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func (d *Decoder) decodeInteger(value []byte, v reflect.Value) error {
	i, err := strconv.ParseInt(*(*string)(unsafe.Pointer(&value)), 10, v.Type().Bits())
	if err != nil {
		return err
	}
	v.SetInt(i)
	return nil
}

func (d *Decoder) decodeUnsigned(value []byte, v reflect.Value) error {
	i, err := strconv.ParseUint(*(*string)(unsafe.Pointer(&value)), 10, v.Type().Bits())
	if err != nil {
		return err
	}
	v.SetUint(i)
	return nil
}
