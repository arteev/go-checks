package checks

import "testing"

func TestCheck2(t *testing.T) {
	type testNested struct {
		ID    int
		SN    *int
		Array []byte `tag_nested_array`
	}
	type testType struct {
		Value          int
		testNested     `tag_str_nested`
		ValueStr       string                `tag_str`
		ArrayStruct    []testNested          `tag_array`
		ArrayStructPtr []testNested          `tag_array_ptr`
		Array          []byte                `tag_array_byte`
		Map            map[string]testNested `tag_map`
	}

	Check2(&testType{
		Value: 1,
		ArrayStruct: []testNested{
			{ID: 123, Array: []byte{1, 2, 3}},
		},
		Array: []byte{6, 7, 8},
		Map: map[string]testNested{
			"729": {ID: 789},
			"999": {ID: 999},
		},
	})
}
