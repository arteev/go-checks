package checks

import (
	"bytes"
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testNestedChecker struct {
	Error error
}

type testNestedNoChecker struct {
	Error error
}

type testStruct struct {
	Error         error
	NestedCheck   testNestedChecker
	NestedNoCheck testNestedNoChecker

	NestedCheckPtr   *testNestedChecker
	NestedNoCheckPtr *testNestedNoChecker

	SliceCheckers []interface{}
	MapChecker    map[string]interface{}
}

type testNoCheck struct{}

type testNoCheckNested struct {
	Nested    testNoCheck
	NestedPtr *testNoCheck
}

func (c testStruct) Check() error {
	return c.Error
}
func (c testNestedChecker) Check() error {
	return c.Error
}

func TestCheckOverFunc(t *testing.T) {

	//no checker
	err := Check(nil)
	assert.NoError(t, err)

	var vNil *testStruct
	err = Check(vNil)
	assert.NoError(t, err)

	err = Check(testNoCheck{})
	assert.NoError(t, err)

	err = Check(&testNoCheck{})
	assert.NoError(t, err)

	err = Check(&testNoCheckNested{
		NestedPtr: &testNoCheck{},
	})
	assert.NoError(t, err)

	//with checker
	err = Check(&testStruct{})
	assert.NoError(t, err)

	err = Check(testStruct{
		Error: errors.New("err1"),
	})
	assert.EqualError(t, err, "err1")

	err = Check(&testStruct{
		NestedCheck: testNestedChecker{
			Error: errors.New("nested err1"),
		},
	})
	assert.EqualError(t, err, "nested err1")

	err = Check(&testStruct{
		NestedCheckPtr: &testNestedChecker{
			Error: errors.New("nested ptr err1"),
		},
	})
	assert.EqualError(t, err, "nested ptr err1")

	//slice
	err = Check(&testStruct{
		SliceCheckers: []interface{}{
			testNestedChecker{
				Error: errors.New("nested slice err1"),
			},
		},
	})
	assert.EqualError(t, err, "nested slice err1")

	err = Check(&testStruct{
		SliceCheckers: []interface{}{
			testNestedChecker{},
			testNestedChecker{
				Error: errors.New("nested slice err2"),
			},
			&testNoCheckNested{},
		},
	})
	assert.EqualError(t, err, "nested slice err2")

	err = Check(&testStruct{
		SliceCheckers: []interface{}{
			testNestedChecker{},
			nil,
			&testNestedChecker{},
			&testNoCheckNested{},
			testNoCheckNested{},
		},
	})
	assert.NoError(t, err)

	//map
	err = Check(&testStruct{
		MapChecker: map[string]interface{}{},
	})
	assert.NoError(t, err)

	err = Check(&testStruct{
		MapChecker: map[string]interface{}{
			"0": nil,
			"1": testNestedChecker{
				Error: errors.New("nested map err1"),
			},
			"2": &testNestedChecker{},
		},
	})
	assert.EqualError(t, err, "nested map err1")

	err = Check(&testStruct{
		MapChecker: map[string]interface{}{
			"0": nil,
			"1": "string",
			"2": &testNestedChecker{},
			"3": testNestedNoChecker{},
		},
	})
	assert.NoError(t, err)
}

type testTags struct {
	NoTag            string
	FieldRequired    string  `check:"required"`
	FieldRequiredPtr *string `check:"required"`
}

func TestCheckOverTagsRequired(t *testing.T) {
	err := Check(&testTags{})
	assert.EqualError(t, err, "value required: FieldRequired")

	err = Check(testTags{})
	assert.EqualError(t, err, "value required: FieldRequired")

	err = Check(testTags{
		FieldRequired: "123",
	})
	assert.EqualError(t, err, "value required: FieldRequiredPtr")

	s := "21312"
	err = Check(testTags{
		FieldRequired:    "123",
		FieldRequiredPtr: &s,
	})
	assert.NoError(t, err)
}

func TestCheckOverTagsExpect(t *testing.T) {
	//bad syntax
	type testBadSyntax struct {
		Field string `check:"expect:"`
	}
	err := Check(&testBadSyntax{})
	assert.EqualError(t, err, `bad syntax: "Field" expect:`)

	//string empty
	type testTagsExpect struct {
		FieldExpectStr string `check:"expect:bar;foo;"` //Can empty, last ;
	}

	err = Check(&testTagsExpect{
		FieldExpectStr: "bzz",
	})
	assert.EqualError(t, err, `unexpected value: FieldExpectStr "bzz"`)

	err = Check(&testTagsExpect{
		FieldExpectStr: "",
	})
	assert.NoError(t, err)

	err = Check(&testTagsExpect{
		FieldExpectStr: "bar",
	})
	assert.NoError(t, err)

	type testTagsExpectNoempty struct {
		FieldExpectStrNoEmpty string `check:"expect:warn;error"`
	}
	err = Check(&testTagsExpectNoempty{})
	assert.EqualError(t, err, `unexpected value: FieldExpectStrNoEmpty ""`)

	//string ptr
	type testTagsPtr struct {
		FieldStrPtr *string `check:"expect:bar;foo;baz"`
	}
	err = Check(&testTagsPtr{})
	assert.EqualError(t, err, `unexpected value: FieldStrPtr <nil>`)

	s := "test"
	err = Check(&testTagsPtr{
		FieldStrPtr: &s,
	})
	assert.EqualError(t, err, `unexpected value: FieldStrPtr "test"`)

	s = "foo"
	err = Check(&testTagsPtr{
		FieldStrPtr: &s,
	})
	assert.NoError(t, err)

	type testInt struct {
		Field int `check:"expect:1;2;3;50;23"`
	}
	err = Check(&testInt{})
	assert.EqualError(t, err, `unexpected value: Field "0"`)

	err = Check(&testInt{Field: 50})
	assert.NoError(t, err)
}

func TestCheckDeprecated(t *testing.T) {

	buf := &bytes.Buffer{}

	log.SetOutput(buf)

	type testDep struct {
		Field1 int  `check:"deprecated"`
		Field2 bool `check:"deprecated"`
	}
	err := Check(&testDep{})
	assert.NoError(t, err)
	assert.Zero(t, buf.Len())

	err = Check(&testDep{
		Field1: 1,
	})
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), `deprecated parameter "Field1" discouraged from using, because it is dangerous, or because a better alternative exists`)

	buf = &bytes.Buffer{}
	log.SetOutput(buf)
	err = Check(&testDep{
		Field1: 1,
		Field2: true,
	})
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), `deprecated parameter "Field1" discouraged from using, because it is dangerous, or because a better alternative exists`)
	assert.Contains(t, buf.String(), `deprecated parameter "Field2" discouraged from using, because it is dangerous, or because a better alternative exists`)

}

type testTagAndChecker struct {
	Enabled bool
	err     error
	Field   fieldUint `check:"expect:1;3;5;7"`
}
type fieldUint uint

func (t fieldUint) Check() error {
	if t == 4 {
		return errors.New("test error 4")
	}
	return nil
}

func (t testTagAndChecker) Check() error {
	if !t.Enabled {
		return ErrSkip
	}
	return nil
}

func TestCheckTagsAndChecker(t *testing.T) {
	err := Check(&testTagAndChecker{Enabled: true})
	assert.EqualError(t, err, `unexpected value: Field "0"`)

	err = Check(&testTagAndChecker{
		Enabled: true,
		Field:   4,
	})
	assert.EqualError(t, err, "test error 4")

	err = Check(&testTagAndChecker{
		Enabled: false,
	})
	assert.NoError(t, err)
}
