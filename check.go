package checks

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

//Errors
var (
	ErrValueRequired   = errors.New("value required")
	ErrValueUnexpected = errors.New("unexpected value")
	ErrDeprecated      = errors.New("deprecated parameter")
	ErrBadSyntax       = errors.New("bad syntax")
	ErrSkip            = errors.New("skip")
)

const (
	ModeFirst Mode = iota
	ModeAll
)

type (
	Mode int

	//Checker is the interface that wraps the basic Check method.
	Checker interface {
		Check() error
	}

	checker struct {
		mode      Mode
		errorMode Type
	}
)

func isZero(value reflect.Value) bool {
	v := reflect.Indirect(value)
	if value.Kind() == reflect.Ptr && !isNil(value) {
		return false
	}
	switch v.Kind() {
	case reflect.Bool:
		return false
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
		return newError(ErrValueRequired, strField.Name, nil, ErrorType)
	}
	return nil
}

func checkDeprecated(value reflect.Value, strField *reflect.StructField) error {
	if isNil(value) || !value.IsValid() || isZero(value) {
		return nil
	}
	return newError(ErrDeprecated, strField.Name, nil, WarningType)
}

func checkExpect(value reflect.Value, strField *reflect.StructField) error {
	value = reflect.Indirect(value)
	sTag := strField.Tag.Get("check")
	sTagValues := strings.SplitN(sTag, ":", 2)
	if len(sTagValues) != 2 || sTagValues[1] == "" {
		return newError(ErrBadSyntax, strField.Name, sTag, ErrorType)
	}
	sTagValues = strings.Split(sTagValues[1], ";")

	if isNil(value) || !value.IsValid() {
		return newError(ErrValueUnexpected, strField.Name, "<nil>", ErrorType)
	}

	sValue := fmt.Sprintf("%v", value.Interface())
	if !hasValue(sValue, sTagValues) {
		return newError(ErrValueUnexpected, strField.Name, value.Interface(), ErrorType)
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
	value = reflect.Indirect(value)
	return value.IsValid() && !isNil(value) && value.CanInterface()
}

//New returns new checker
func New(m Mode, e Type) *checker {
	return &checker{
		mode:      m,
		errorMode: e,
	}
}

func (c *checker) Check(v interface{}) []error {
	result := make([]error, 0)
	iter := newIterator(v)
	for iter.HasNext() {
		item := iter.Next()
		if err := checkValue(item); err != nil {
			if err == ErrSkip {
				return nil
			}

			if are, ok := err.(ErrorCheckResult); ok {
				typ := are.GetType()
				if (typ & c.errorMode) != typ {
					continue
				}
			}

			result = append(result, err)
			if c.mode == ModeFirst {
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
	errs := New(ModeFirst, ErrorType).Check(v)
	if len(errs) == 0 {
		return nil
	}
	return errs[0]
}

//CheckAll check all fields of the structure
func CheckAll(v interface{}) []error {
	return New(ModeAll, ErrorType).Check(v)
}
