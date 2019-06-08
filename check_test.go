package checks

import (
	"errors"
	"fmt"
	"reflect"
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
	FieldBool        bool    `check:"required"`
	FieldBoolPtr     *bool   `check:"required"`
}

func TestCheckOverTagsRequired(t *testing.T) {
	err := Check(&testTags{})
	assert.EqualError(t, err, "value required: FieldRequired")

	err = Check(testTags{
		FieldRequired: "123",
	})
	assert.EqualError(t, err, "value required: FieldRequiredPtr")

	s := ""
	err = Check(testTags{
		FieldRequired:    "123",
		FieldRequiredPtr: &s,
	})
	assert.EqualError(t, err, "value required: FieldBoolPtr")

	s = "21312"
	b := false
	err = Check(testTags{
		FieldRequired:    "123",
		FieldRequiredPtr: &s,
		FieldBoolPtr:     &b,
	})
	assert.NoError(t, err)
}

func TestCheckOverTagsExpect(t *testing.T) {
	//bad syntax
	type testBadSyntax struct {
		Field string `check:"expect:"`
	}
	err := Check(&testBadSyntax{})
	assert.EqualError(t, err, `bad syntax: Field expect:`)

	//string empty
	type testTagsExpect struct {
		FieldExpectStr string `check:"expect:bar;foo;"` //Can empty, last ;
	}

	err = Check(&testTagsExpect{
		FieldExpectStr: "bzz",
	})
	assert.EqualError(t, err, `unexpected value: FieldExpectStr bzz`)

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
	assert.EqualError(t, err, `unexpected value: FieldExpectStrNoEmpty `)

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
	assert.EqualError(t, err, `unexpected value: FieldStrPtr test`)

	s = "foo"
	err = Check(&testTagsPtr{
		FieldStrPtr: &s,
	})
	assert.NoError(t, err)

	type testInt struct {
		Field int `check:"expect:1;2;3;50;23"`
	}
	err = Check(&testInt{})
	assert.EqualError(t, err, `unexpected value: Field 0`)

	err = Check(&testInt{Field: 50})
	assert.NoError(t, err)
}

func TestCheckDeprecated(t *testing.T) {
	type testDep struct {
		Field1 int    `check:"deprecated"`
		Field2 string `check:"deprecated"`
		Field3 *byte  `check:"deprecated"`
		Field4 []byte `check:"deprecated"`
	}
	err := New(ModeFirst, WarningType).Check(&testDep{})
	assert.Len(t, err, 0)

	err = New(ModeFirst, WarningType).Check(&testDep{
		Field1: 1,
	})
	assert.EqualError(t, err[0], `deprecated parameter: Field1`)

	err = New(ModeFirst, WarningType).Check(&testDep{
		Field2: "123",
		Field4: []byte{1},
	})
	assert.EqualError(t, err[0], `deprecated parameter: Field2`)
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
	assert.EqualError(t, err, `unexpected value: Field 0`)

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

func TestChecker(t *testing.T) {
	c := New(ModeAll, ErrorAll)
	assert.NotNil(t, c)
	assert.Equal(t, ModeAll, c.mode)
}

func TestCheckAll(t *testing.T) {
	type testStrAll struct {
		A int    `check:"expect:1;2;3"`
		B *int   `check:"required"`
		C string `check:"deprecated"`
	}
	errs := CheckAll(&testStrAll{A: 10})
	assert.Len(t, errs, 2)
	assert.EqualError(t, errs[0], `unexpected value: A 10`)
	assert.EqualError(t, errs[1], `value required: B`)

	errs = CheckAll(&testStrAll{A: 1})
	assert.Len(t, errs, 1)
	assert.EqualError(t, errs[0], `value required: B`)

	b := 12
	errs = CheckAll(&testStrAll{A: 1, B: &b})
	assert.Len(t, errs, 0)

}

func TestIsZero(t *testing.T) {
	type testZero struct {
		A int
		B int
	}
	strEmpty := ""
	zeros := []struct {
		v    interface{}
		must bool
	}{
		{nil, true},
		{0, true},
		{"", true},
		{0.00, true},
		{(func())(nil), true},
		{(map[string]string)(nil), true},
		{([]byte)(nil), true},
		{[]uint{}, true},
		{(*string)(nil), true},

		{[]uint{0}, false},

		{&testZero{}, false},
		{testZero{}, true},
		{false, false},
		{true, false},
		{strEmpty, true},
		{&strEmpty, false},
	}

	for _, test := range zeros {
		got := isZero(reflect.ValueOf(test.v))
		assert.Equal(t, test.must, got, "value: %v", test.v)
	}
}

func TestCheckerErrorTypes(t *testing.T) {
	type testTypes struct {
		A int    `check:"expect:1;2;3"`
		B string `check:"deprecated"`
	}
	value := testTypes{
		A: 10,
		B: "123",
	}
	c := New(ModeAll, ErrorAll)

	errs := c.Check(value)
	assert.Len(t, errs, 2)
	assert.EqualError(t, errs[0], "unexpected value: A 10")
	assert.EqualError(t, errs[1], "deprecated parameter: B")

	c = New(ModeAll, ErrorType)
	errs = c.Check(value)
	assert.Len(t, errs, 1)
	assert.EqualError(t, errs[0], "unexpected value: A 10")

	c = New(ModeAll, WarningType)
	errs = c.Check(value)
	assert.Len(t, errs, 1)
	assert.EqualError(t, errs[0], "deprecated parameter: B")
}

func TestCheckerMultiCheck(t *testing.T) {
	type testTypes struct {
		A int `check:"required,expect:1;2;3"`
	}
	value := testTypes{}
	c := New(ModeAll, ErrorAll)
	errs := c.Check(value)
	assert.Len(t, errs, 2)
	assert.EqualError(t, errs[0], "value required: A")
	assert.EqualError(t, errs[1], "unexpected value: A 0")

	value = testTypes{A: 10}
	errs = c.Check(value)
	assert.Len(t, errs, 1)
	assert.EqualError(t, errs[0], "unexpected value: A 10")

	value = testTypes{A: 2}
	errs = c.Check(value)
	assert.Len(t, errs, 0)
}

type TestMethod struct {
	Value  string `check:"call:CheckValue"`
	IntPtr *int   `check:"call:CheckValueIntPtr"`
}

func (m TestMethod) CheckValue(field, value string) error {
	if value == "valid" {
		return nil
	}
	return errors.New("error from method")
}

func (m TestMethod) CheckValueIntPtr(field string, value *int) error {
	if value == nil || *value != 100 {
		return fmt.Errorf("required value")
	}
	return nil
}

type testMethodWrong struct {
	Value int `check:"call:CheckValue"`
}

func (m testMethodWrong) CheckValue(field string, value int) {}

type testMethodWrongRet struct {
	Value int `check:"call:CheckValue"`
}

func (m testMethodWrongRet) CheckValue(field string, value int) int { return 1 }

func TestOverMethod(t *testing.T) {
	err := Check(struct {
		Value int `check:"call:"`
	}{})
	assert.EqualError(t, err, "bad syntax: Value call:")

	err = Check(struct {
		Value int `check:"call:method"`
	}{})
	assert.EqualError(t, err, "method not found: method")

	err = Check(testMethodWrong{})
	assert.EqualError(t, err, "wrong signature method: Value call:CheckValue")

	err = Check(testMethodWrongRet{})
	assert.EqualError(t, err, "wrong signature method: Value call:CheckValue")

	v := TestMethod{}
	err = Check(v)
	assert.EqualError(t, err, "error from method")

	v.Value = "valid"
	err = Check(v)
	assert.EqualError(t, err, "required value")

	i := 100
	v.Value = "valid"
	v.IntPtr = &i
	err = Check(v)
	assert.NoError(t, err)

	//nested
	type TestMethodWrap struct {
		Nested TestMethod
	}
	vw := TestMethodWrap{
		Nested: TestMethod{},
	}
	err = Check(vw)
	assert.EqualError(t, err, "error from method")
}
