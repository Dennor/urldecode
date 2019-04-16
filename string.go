package urldecode

import (
	"reflect"
)

func isString(v reflect.Value) bool {
	return v.Kind() == reflect.String
}

func (d *Decoder) decodeString(value []byte, v reflect.Value) error {
	v.SetString(string(value))
	return nil
}
