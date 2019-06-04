package checks

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
)

//Errors
var (
	ErrValueRequired   = errors.New("value required")
	ErrValueUnexpected = errors.New("unexpected value")
	ErrBadSyntax       = errors.New("bad syntax")
	ErrSkip            = errors.New("skip")
)

const (
	modeFirst mode = iota
	modeAll
)

type (
	mode int

	//Checker is the interface that wraps the basic Check method.
	Checker interface {
		Check() error
	}

	checker struct {
		mode mode
	}
)

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Func:
		return v.IsNil()
	case reflect.Map, reflect.Array, reflect.Slice:
		return v.IsNil() || (v.Len() == 0)
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

func checkInterfaceChecker(value reflect.Value) error {
	if !isInterface(value) {
		return nil
	}
	check, ok := value.Interface().(Checker)
	if !ok {
		return nil
	}
	return check.Check()
}

func checkValue(v Value) error {
	value := *v.Value()
	if err := checkInterfaceChecker(value); err != nil {
		return err
	}

	if v.Struct() == nil {
		return nil
	}

	sTag, ok := v.Tag().Lookup("check")
	if !ok {
		return nil
	}

	var errCheck error
	switch sTag {
	case "required":
		errCheck = checkRequired(value, v.Struct())
	case "deprecated":
		errCheck = checkDeprecated(value, v.Struct())
	default:
		if strings.HasPrefix(sTag, "expect:") {
			errCheck = checkExpect(value, v.Struct())
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

func isNil(value reflect.Value) bool {
	return (value.Kind() == reflect.Ptr || value.Kind() == reflect.Interface) && value.IsNil()
}

func isInterface(value reflect.Value) bool {
	return value.IsValid() && !isNil(value) && value.CanInterface()
}

func new(m mode) *checker {
	return &checker{
		mode: m,
	}
}

func (c *checker) check(v interface{}) []error {
	result := make([]error, 0)
	iter := newIterator(v)
	for iter.HasNext() {
		item := iter.Next()
		if err := checkValue(item); err != nil {
			if err == ErrSkip {
				return nil
			}
			result = append(result, err)
			if c.mode == modeFirst {
				break
			}
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

//Check check structure
func Check(v interface{}) error {
	errs := new(modeFirst).check(v)
	if len(errs) == 0 {
		return nil
	}
	return errs[0]
}

//CheckAll check all fields of the structure
func CheckAll(v interface{}) []error {
	return new(modeAll).check(v)
}
