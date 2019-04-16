// Package decodes url encoded values into struct

package urldecode

import (
	"bytes"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"sync"
)

type decoderPool struct {
	sync.Pool
}

func (d *decoderPool) Get() *Decoder {
	dec := d.Pool.Get().(*Decoder)
	dec.r = nil
	dec.query.Reset()
	dec.read.Reset()
	return dec
}

func (d *decoderPool) Put(enc *Decoder) {
	d.Pool.Put(enc)
}

func newDecoderPool() *decoderPool {
	return &decoderPool{
		sync.Pool{
			New: func() interface{} {
				return &Decoder{}
			},
		},
	}
}

var decoderStatePool = newDecoderPool()

// Decoder reads values from url encoded values string
// sets them in target value
type Decoder struct {
	scratch scratchBuffer
	query   bytes.Buffer
	read    bytes.Buffer
	r       io.Reader
}

func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}

func unhex(c byte) byte {
	switch {
	case '0' <= c && c <= '9':
		return c - '0'
	case 'a' <= c && c <= 'f':
		return c - 'a' + 10
	case 'A' <= c && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}

// copy paste from https://golang.org/src/net/url/url.go with unencoded query buffering
// and removed unnecessary checks, as we are only interested in query encoding
func (d *Decoder) unescape(b []byte) error {
	d.query.Reset()
	// Count %, check that they're well-formed.
	n := 0
	for i := 0; i < len(b); {
		switch b[i] {
		case '%':
			n++
			if i+2 >= len(b) || !ishex(b[i+1]) || !ishex(b[i+2]) {
				b = b[i:]
				if len(b) > 3 {
					b = b[:3]
				}
				return url.EscapeError(string(b))
			}
			i += 3
		default:
			i++
		}
	}
	d.query.Reset()
	d.query.Grow(len(b) - 2*n)
	j := 0
	for i := 0; i < len(b); {
		switch b[i] {
		case '%':
			d.scratch[j] = unhex(b[i+1])<<4 | unhex(b[i+2])
			i += 3
		case '+':
			d.scratch[j] = ' '
			i++
		default:
			d.scratch[j] = b[i]
			i++
		}
		j++
		if j == len(d.scratch) {
			d.query.Write(d.scratch[:j])
			j = 0
		}
	}
	d.query.Write(d.scratch[:j])
	return nil
}

func (d *Decoder) decodeToken(value []byte, v reflect.Value) error {
	if v.Type().Implements(unmarshalerType) {
		return d.decodeUnmarshaler(value, v)
	}
	if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			if v.Kind() == reflect.Interface {
				return TypeError{fmt.Errorf("can't decode into interface type")}
			}
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	switch {
	case isBool(v):
		return d.decodeBool(value, v)
	case isInteger(v):
		return d.decodeInteger(value, v)
	case isUnsigned(v):
		return d.decodeUnsigned(value, v)
	case isFloating(v):
		return d.decodeFloating(value, v)
	case isString(v):
		return d.decodeString(value, v)
	}
	return TypeError{fmt.Errorf("can't decode into %s type", v.Type().Name())}
}

func splitShift(b []byte, iat int) ([]byte, []byte) {
	if iat == -1 {
		return b, b[len(b):]
	}
	return b[:iat], b[iat+1:]
}

func checkAndAlloc(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
	}
	return v
}

func (d *Decoder) decodeQuery(v reflect.Value) error {
	var err error
	if d.r != nil {
		_, err = io.Copy(&d.read, d.r)
	}
	if err != nil {
		return err
	}
	fields := cachedTypeFields(v.Type())
	if err := d.unescape(d.read.Bytes()); err != nil {
		return err
	}
	b := d.query.Bytes()
	for len(b) > 0 {
		var token []byte
		token, b = splitShift(b, bytes.IndexByte(b, '&'))
		if len(token) == 0 {
			continue
		}
		var key []byte
		key, token = splitShift(token, bytes.IndexByte(token, '='))
		var f *field
		for i := range fields {
			if fields[i].foldFunc(fields[i].nameBytes, key) {
				f = &fields[i]
				break
			}
		}
		if f != nil {
			fv := checkAndAlloc(v.Field(f.index[0]))
			for i := 1; i < len(f.index); i++ {
				if fv.Kind() == reflect.Ptr {
					fv = fv.Elem()
				}
				fv = checkAndAlloc(fv.Field(f.index[i]))
			}
			d.decodeToken(token, fv)
		}
	}
	return nil
}

func (d *Decoder) decode(v reflect.Value) error {
	if v.Kind() != reflect.Ptr {
		return TypeError{fmt.Errorf("must be a pointer")}
	}
	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return TypeError{fmt.Errorf("must be a struct")}
	}
	return d.decodeQuery(v)
}

// Decode url into struct/url.Values
func (d *Decoder) Decode(v interface{}) error {
	if err := d.decode(reflect.ValueOf(v)); err != nil {
		return err
	}
	return nil
}

// NewDecoder creates new decoder
func NewDecoder(r io.Reader) *Decoder {
	dec := decoderStatePool.Get()
	dec.r = r
	return dec
}

// Unmarshal url encoded data into v
func Unmarshal(data []byte, v interface{}) error {
	dec := decoderStatePool.Get()
	dec.read.Write(data)
	err := dec.decode(reflect.ValueOf(v))
	decoderStatePool.Put(dec)
	return err
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}
