package checks

import (
	"fmt"
	"log"
	"reflect"

	"strings"

	"github.com/go-errors/errors"
)

//Errors
var (
	ErrValueRequired   = errors.New("value required")
	ErrValueUnexpected = errors.New("unexpected value")
	ErrBadSyntax       = errors.New("bad syntax")
	ErrSkip            = errors.New("skip")
)

//Checker is the interface that wraps the basic Check method.
type Checker interface {
	Check() error
}

//Check check structure
func Check(v interface{}) error {
	value := reflect.ValueOf(v)
	if v == nil || isNil(value) {
		return nil
	}
	return walkCheck(value, nil)
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Func:
		return v.IsNil()
	case reflect.Map, reflect.Slice:
		return v.IsNil() || (v.Len() == 0)
	case reflect.Array:
		z := true
		for i := 0; i < v.Len(); i++ {
			z = z && isZero(v.Index(i))
		}
		return z
	case reflect.Struct:
		z := true
		for i := 0; i < v.NumField(); i++ {
			z = z && isZero(v.Field(i))
		}
		return z
	}
	z := reflect.Zero(v.Type())
	return v.Interface() == z.Interface()
}

func checkRequired(value reflect.Value, strField *reflect.StructField) error {
	if isNil(value) || !value.IsValid() || isZero(value) {
		return fmt.Errorf("%s: %s", ErrValueRequired, strField.Name)
	}
	return nil
}

func checkDeprecated(value reflect.Value, strField *reflect.StructField) error {
	if isNil(value) || !value.IsValid() || isZero(value) {
		return nil
	}
	log.Printf("deprecated parameter %q discouraged from using, because it is dangerous, or because a better alternative exists", strField.Name)
	return nil
}

func checkExpect(value reflect.Value, strField *reflect.StructField) error {
	sTag := strField.Tag.Get("check")
	sTagValues := strings.SplitN(sTag, ":", 2)
	if len(sTagValues) != 2 || sTagValues[1] == "" {
		return fmt.Errorf("%s: %q %v", ErrBadSyntax, strField.Name, sTag)
	}
	sTagValues = strings.Split(sTagValues[1], ";")

	if isNil(value) || !value.IsValid() {
		return fmt.Errorf("%s: %s %v", ErrValueUnexpected, strField.Name, nil)
	}

	sValue := fmt.Sprintf("%v", value.Interface())
	if !hasValue(sValue, sTagValues) {
		return fmt.Errorf("%s: %s %q", ErrValueUnexpected, strField.Name, sValue)
	}
	return nil
}

func isNil(value reflect.Value) bool {
	return (value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface) && value.IsNil()
}

func IsInterface(value reflect.Value) bool {
	return value.IsValid() && !isNil(value) && value.CanInterface()
}

func checkInterfaceChecker(value reflect.Value) error {
	if !IsInterface(value) {
		return nil
	}
	check, ok := value.Interface().(Checker)
	if !ok {
		return nil
	}
	return check.Check()
}

func checkValue(value reflect.Value, strField *reflect.StructField) error {
	if err := checkInterfaceChecker(value); err != nil {
		return err
	}

	if strField == nil {
		return nil
	}

	sTag, ok := strField.Tag.Lookup("check")
	if !ok {
		return nil
	}

	var errCheck error
	switch sTag {
	case "required":
		errCheck = checkRequired(value, strField)
	case "deprecated":
		errCheck = checkDeprecated(value, strField)
	default:
		if strings.HasPrefix(sTag, "expect:") {
			errCheck = checkExpect(value, strField)
		}
	}
	if errCheck != nil {
		return errCheck
	}

	return nil
}

func hasValue(value string, values []string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func checkStruct(value reflect.Value) error {
	for i := 0; i < value.NumField(); i++ {
		cType := value.Type()
		strFieldCur := cType.Field(i)
		if err := walkCheck(value.Field(i), &strFieldCur); err != nil {
			return err
		}
	}
	return nil
}

func checkSlice(value reflect.Value) error {
	for i := 0; i < value.Len(); i++ {
		if err := walkCheck(value.Index(i), nil); err != nil {
			return err
		}
	}
	return nil
}

func checkMap(value reflect.Value) error {
	for _, key := range value.MapKeys() {
		if err := walkCheck(value.MapIndex(key), nil); err != nil {
			return err
		}
	}
	return nil
}

func walkCheck(value reflect.Value, strField *reflect.StructField) error {
	value = reflect.Indirect(value)
	if err := checkValue(value, strField); err != nil {
		if err == ErrSkip {
			return nil
		}
		return err
	}
	switch value.Kind() {
	case reflect.Ptr, reflect.Interface:
		if value.IsNil() {
			return nil
		}
	case reflect.Struct:
		return checkStruct(value)
	case reflect.Slice, reflect.Array:
		return checkSlice(value)
	case reflect.Map:
		return checkMap(value)
	default:
		return nil
	}
	return nil
}

func Check2(v interface{}) error {
	value := reflect.ValueOf(v)
	if v == nil || isNil(value) {
		return nil
	}
	value = reflect.Indirect(value)

	iter := newIterator(value)
	for iter.HasNext() {
		item := iter.Next()
		log.Println(item.Name(), "TAG:", item.Tag(), "VALUE:", item.Value())

		//check next()
	}

	return nil
}
