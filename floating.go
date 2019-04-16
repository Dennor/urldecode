package urldecode

import (
	"reflect"
	"strconv"
	"unsafe"
)

func isFloating(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return true
	default:
		return false
	}
}

func (d *Decoder) decodeFloating(value []byte, v reflect.Value) error {
	f, err := strconv.ParseFloat(*(*string)(unsafe.Pointer(&value)), v.Type().Bits())
	if err != nil {
		return err
	}
	v.SetFloat(f)
	return nil
}
