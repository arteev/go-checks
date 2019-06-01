package checks

import (
	"reflect"
)

type Iterator interface {
	HasNext() bool
	Next() Value
}

type iterator struct {
	idx   int
	nodes []node
}

type node struct {
	value    reflect.Value
	strField *reflect.StructField
}

type Value interface {
	Name() string
	Tag() reflect.StructTag
	Value() reflect.Value
}

func canIterate(v reflect.Value) bool {
	kind := v.Kind()
	return kind == reflect.Struct ||
		kind == reflect.Slice ||
		kind == reflect.Array ||
		kind == reflect.Map
}

func (n node) Name() string {
	if n.strField == nil {
		return n.value.Type().Name()
	}
	return n.strField.Name
}
func (n node) Tag() reflect.StructTag {
	if n.strField == nil {
		return ""
	}
	return n.strField.Tag
}
func (n node) Value() reflect.Value { return n.value }

func (i *iterator) HasNext() bool {
	if i == nil {
		return false
	}
	return i.idx < len(i.nodes)
}

func (i *iterator) Next() Value {
	result := i.nodes[i.idx]
	i.idx++
	return result
}

func (i *iterator) initValue(v reflect.Value, sf *reflect.StructField) {
	//cType := value.Type()
	//strFieldCur := cType.Field(i)

	//if err := walkCheck(value.Field(i), &strFieldCur); err != nil {
	//return err
	//}

	if canIterate(v) {
		i.nodes = append(i.nodes, node{
			value:    v,
			strField: sf,
		})

		i.init(v, sf)
		return
	}
	i.nodes = append(i.nodes, node{
		value:    v,
		strField: sf,
	})
	//log.Println(v, sf.Name, sf.Type.Name())
}

func (i *iterator) init(v reflect.Value, sf *reflect.StructField) {
	switch v.Kind() {
	case reflect.Struct:
		for k := 0; k < v.NumField(); k++ {
			cType := v.Type()
			strFieldCur := cType.Field(k)
			i.initValue(v.Field(k), &strFieldCur)
		}
	case reflect.Slice, reflect.Array:
		for k := 0; k < v.Len(); k++ {
			value := v.Index(k)
			if !isNil(value) {
				value = reflect.Indirect(value)
				i.init(value, sf)
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)
			if !isNil(value) {
				value = reflect.Indirect(value)
				i.init(value, sf)
			}
		}
	default:
		i.initValue(v, nil)
	}

}

func newIterator(v reflect.Value) Iterator {
	if !canIterate(v) {
		return (*iterator)(nil)
	}
	result := &iterator{}
	result.init(v, nil)
	return result
}
