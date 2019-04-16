package urldecode

import (
	"reflect"
)

func isBool(v reflect.Value) bool {
	return v.Kind() == reflect.Bool
}

func equalFold(b1, b2 byte) bool {
	return b1 == b2 || b1 == b2+'a'-'A'
}

func (d *Decoder) decodeBool(v []byte, dst reflect.Value) error {
	switch len(v) {
	case 1:
		dst.SetBool(v[0] == '1' || equalFold('y', v[0]))
	case 3:
		dst.SetBool(equalFold('y', v[0]) && equalFold('e', v[1]) && equalFold('s', v[2]))
	case 4:
		dst.SetBool(equalFold('t', v[0]) && equalFold('r', v[1]) && equalFold('u', v[2]) && equalFold('e', v[3]))
	default:
		dst.SetBool(false)
	}
	return nil
}
