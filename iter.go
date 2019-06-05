package checks

import (
	"reflect"
)

type (
	//Iterator is the interface iterator
	Iterator interface {
		HasNext() bool
		Next() Value
	}

	//Value is the interface Value of structs
	Value interface {
		Name() string
		Tag() reflect.StructTag
		Value() *reflect.Value
		Parent() Value
		Struct() *reflect.StructField
	}

	iterator struct {
		idx   int
		nodes []*node
	}

	node struct {
		ptr      bool
		parent   *node
		value    *reflect.Value
		strField *reflect.StructField
	}
)

var iterateable = map[reflect.Kind]struct{}{
	reflect.Struct: struct{}{},
	reflect.Slice:  struct{}{},
	reflect.Array:  struct{}{},
	reflect.Map:    struct{}{},
}

func canIterate(v reflect.Value) bool {
	v = reflect.Indirect(v)
	_, ok := iterateable[v.Kind()]
	return ok
}

func (n node) Name() string {
	if n.strField == nil {
		name := reflect.Indirect(*n.value).Type().Name()
		if name == "" {
			name = reflect.Indirect(*n.value).Type().Kind().String()
		}
		return name
	}
	return n.strField.Name
}
func (n node) Tag() reflect.StructTag {
	if n.strField == nil {
		return ""
	}
	return n.strField.Tag
}
func (n node) Value() *reflect.Value { return n.value }

func (n node) Struct() *reflect.StructField { return n.strField }

func (n node) Parent() Value {
	return n.parent
}

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

func (i *iterator) initValue(v reflect.Value, sf *reflect.StructField, parent *node) *node {
	ptr := false
	value := v
	if v.Kind() == reflect.Ptr {
		value = reflect.Indirect(v)
		ptr = true
	}
	item := &node{
		ptr:      ptr,
		value:    &v,
		strField: sf,
		parent:   parent,
	}
	i.nodes = append(i.nodes, item)

	if canIterate(v) {
		i.initInterateable(value, nil, item)
	}
	return item
}

func (i *iterator) initInterateable(v reflect.Value, sf *reflect.StructField, parent *node) {
	switch v.Kind() {
	case reflect.Struct:
		for k := 0; k < v.NumField(); k++ {
			cType := v.Type()
			strFieldCur := cType.Field(k)
			i.initValue(v.Field(k), &strFieldCur, parent)
		}
	case reflect.Slice, reflect.Array:
		for k := 0; k < v.Len(); k++ {
			value := v.Index(k)
			i.initValue(value, sf, parent)
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			value := v.MapIndex(key)
			i.initValue(value, sf, parent)
		}
	}
}

func newIterator(value interface{}) Iterator {
	v := reflect.ValueOf(value)
	if value == nil || isNil(v) {
		return (*iterator)(nil)
	}
	result := &iterator{}
	result.initValue(v, nil, nil)
	return result
}
