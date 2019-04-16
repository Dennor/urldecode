package urldecode

import (
	"encoding"
	"reflect"
)

var (
	unmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

func (d *Decoder) decodeUnmarshaler(s []byte, v reflect.Value) error {
	unmarshaler := v.Interface().(encoding.TextUnmarshaler)
	return unmarshaler.UnmarshalText(s)
}
