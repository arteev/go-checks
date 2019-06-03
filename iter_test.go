package checks

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func checkItemsHelper(t *testing.T, iter Iterator, must []string) {
	t.Helper()
	i := 0
	for iter.HasNext() {
		item := iter.Next()
		if i >= len(must) {
			assert.Fail(t, "Must: index out of range", i)
			return
		}

		assert.Equal(t, must[i], item.Name())
		i++
	}
	if i != len(must) {
		assert.Fail(t, "Expected: ", must)
	}
}

func TestNode(t *testing.T) {
	assert.Implements(t, (*Value)(nil), &node{})

	//non struct
	value := reflect.ValueOf("str")
	parent := node{
		strField: nil,
		value:    nil,
	}
	n := node{
		parent: &parent,
		value:  &value,
	}
	assert.Equal(t, "string", n.Name())
	assert.Empty(t, n.Tag())
	assert.Equal(t, n.value, n.Value())
	assert.Equal(t, n.parent, n.Parent())

	//struct
	type tstuct struct {
		ValueStr string `check:"string"`
	}
	value = reflect.ValueOf(tstuct{"Test"})
	sValue := value.Field(0)
	field := value.Type().Field(0)
	n = node{
		parent:   &parent,
		value:    &sValue,
		strField: &field,
	}
	assert.Equal(t, "ValueStr", n.Name())
	assert.Equal(t, `check:"string"`, string(n.Tag()))
	assert.Equal(t, n.strField, n.Struct())
}

func TestIter(t *testing.T) {

	assert.Implements(t, (*Iterator)(nil), &iterator{})

	type testStruct struct {
		Field string
		ID    int
	}

	iter := newIterator(nil)
	assert.Nil(t, iter)
	assert.Implements(t, (*Iterator)(nil), iter)

	type testType struct{}
	iter = newIterator(&testType{})
	checkItemsHelper(t, iter, []string{"testType"})

	//single value string
	iter = newIterator("test")
	assert.NotNil(t, iter)
	checkItemsHelper(t, iter, []string{"string"})

	//single value int ptr
	i := 123
	iter = newIterator(&i)
	assert.NotNil(t, iter)
	checkItemsHelper(t, iter, []string{"int"})

	//slice
	arr := []byte{1, 2}
	iter = newIterator(arr)
	checkItemsHelper(t, iter, []string{"slice", "uint8", "uint8"})

	//struct
	iter = newIterator(testStruct{Field: "str", ID: 1})
	checkItemsHelper(t, iter, []string{"testStruct", "Field", "ID"})
	iter = newIterator(&testStruct{})
	checkItemsHelper(t, iter, []string{"testStruct", "Field", "ID"})

	//struct field
	type testNestedStruct struct {
		ID int
		F1 testStruct
	}
	iter = newIterator(testNestedStruct{ID: 10, F1: testStruct{Field: "str", ID: 1}})
	checkItemsHelper(t, iter, []string{"testNestedStruct", "ID",
		"F1", "Field", "ID"})

	//embedded struct
	type testEmbeddedStruct struct {
		testStruct
		NewID int
	}
	iter = newIterator(testEmbeddedStruct{testStruct: testStruct{Field: "str", ID: 1}, NewID: 123})
	checkItemsHelper(t, iter, []string{"testEmbeddedStruct", "testStruct", "Field", "ID",
		"NewID"})

	//slice of struct
	arrayStructs := []testStruct{{Field: "str", ID: 1}}
	iter = newIterator(arrayStructs)
	checkItemsHelper(t, iter, []string{"slice", "testStruct", "Field", "ID"})

	//map of struct
	mStructs := map[string]testStruct{
		"one": {Field: "str", ID: 1},
	}
	iter = newIterator(mStructs)
	checkItemsHelper(t, iter, []string{"map", "testStruct", "Field", "ID"})

	//slice&map in struct
	type testSliceStruct struct {
		NewID int
		Slice []testStruct
		Map   map[string]testStruct
	}

	iter = newIterator(&testSliceStruct{
		NewID: 123,
		Slice: arrayStructs,
		Map:   mStructs,
	})
	checkItemsHelper(t, iter, []string{"testSliceStruct", "NewID", "Slice", "testStruct",
		"Field", "ID", "Map", "testStruct", "Field", "ID"})

}
