package urldecode

import (
	"reflect"
	"strings"
	"sync"
)

func getTag(field *reflect.StructField) (string, []string) {
	tag := field.Tag.Get("url")
	if tag == "" {
		tag = field.Tag.Get("json")
		if tag == "" {
			return "", nil
		}
	}
	parts := strings.Split(tag, ",")
	return parts[0], parts[1:]
}

var fieldCache sync.Map

type field struct {
	typ       reflect.Type
	nameBytes []byte
	foldFunc  func(s, t []byte) bool
	index     []int
	tagged    bool
}

func typeFields(t reflect.Type) []field {
	current := []field{}
	next := []field{{typ: t}}
	visited := map[reflect.Type]bool{}
	fieldAt := map[string]int{}
	orphans := []int{}
	var fields []field
	var level int
	for len(next) > 0 {
		level++
		current, next = next, current[:0]
		nextCount := map[reflect.Type]bool{}
		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				isUnexported := sf.PkgPath != ""
				if sf.Anonymous {
					t := sf.Type
					if t.Kind() == reflect.Ptr {
						t = t.Elem()
					}
					if isUnexported && t.Kind() != reflect.Struct {
						continue
					}
				} else if isUnexported {
					continue
				}
				tag, _ := getTag(&sf)
				if tag == "-" {
					continue
				}
				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i
				name := tag
				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					if fAt, ok := fieldAt[name]; ok {
						if level > len(fields[fAt].index) {
							continue
						}
						if fields[fAt].tagged || (!tagged && !fields[fAt].tagged) {
							continue
						}
						orphans = append(orphans, fAt)
					}
					fieldAt[name] = len(fields)
					fields = append(fields, field{
						typ:       ft,
						tagged:    tagged,
						nameBytes: []byte(name),
						index:     index,
					})
					continue
				}
				if !nextCount[ft] {
					nextCount[ft] = true
					next = append(next, field{index: index, typ: ft})
				}
			}
		}
	}
	for i, orphan := range orphans {
		fields = append(fields[:orphan-i], fields[orphan-i+1:]...)
	}
	for i := range fields {
		fields[i].foldFunc = foldFunc(fields[i].nameBytes)
	}
	return fields
}

func cachedTypeFields(t reflect.Type) []field {
	if f, ok := fieldCache.Load(t); ok {
		return f.([]field)
	}
	f, _ := fieldCache.LoadOrStore(t, typeFields(t))
	return f.([]field)
}

// For now foldFunc is brute force byte by byte comparison
// that only checks for ASCII chars
func foldFunc(b []byte) func(s, t []byte) bool {
	return func(s, t []byte) bool {
		if len(s) != len(t) {
			return false
		}
		for i, b := range s {
			if b != t[i] {
				return false
			}
		}
		return true
	}
}
