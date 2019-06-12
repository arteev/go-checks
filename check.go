package checks

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

//Errors
var (
	ErrValueRequired        = errors.New("value required")
	ErrValueUnexpected      = errors.New("unexpected value")
	ErrDeprecated           = errors.New("deprecated parameter")
	ErrWrongSignatureMethod = errors.New("wrong signature method")
	ErrNoMatch              = errors.New("no matches")
	ErrBadSyntax            = errors.New("bad syntax")
	ErrSkip                 = errors.New("skip")
)

//Known check modes
const (
	ModeFirst Mode = iota
	ModeAll
)

type (
	//Mode check mode
	Mode int

	//Checker is the interface that wraps the basic Check method.
	Checker interface {
		Check() error
	}

	//SimpeChecker implements simple checks: required, expect, deprecated
	SimpeChecker struct {
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

func required(value reflect.Value, strField *reflect.StructField) error {
	if isNil(value) || !value.IsValid() || isZero(value) {
		return newError(ErrValueRequired, strField.Name, nil, ErrorType)
	}
	return nil
}

func deprecated(value reflect.Value, strField *reflect.StructField) error {
	if isNil(value) || !value.IsValid() || isZero(value) {
		return nil
	}
	return newError(ErrDeprecated, strField.Name, nil, WarningType)
}

func expect(value reflect.Value, strField *reflect.StructField) error {
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

func withMethod(root reflect.Value, value reflect.Value, strField *reflect.StructField) error {
	sTag := strField.Tag.Get("check")
	sTagValues := strings.SplitN(sTag, ":", 2)
	if len(sTagValues) != 2 || sTagValues[1] == "" {
		return newError(ErrBadSyntax, strField.Name, sTag, ErrorType)
	}
	methodName := sTagValues[1]
	methodValue := root.MethodByName(methodName)
	if !methodValue.IsValid() {
		return fmt.Errorf("method not found: %s", methodName)
	}

	errValues := methodValue.Call([]reflect.Value{reflect.ValueOf(strField.Name), value})
	errSignature := newError(ErrWrongSignatureMethod, strField.Name, sTag, ErrorType)
	if len(errValues) != 1 {
		return errSignature
	}

	if isNil(errValues[0]) {
		return nil
	}

	err, ok := errValues[0].Interface().(error)
	if !ok {
		return errSignature
	}
	return err
}

func withRegexp(root reflect.Value, value reflect.Value, strField *reflect.StructField) error {
	sTag := strField.Tag.Get("check")
	sTagValues := strings.SplitN(sTag, ":", 2)
	if len(sTagValues) != 2 || sTagValues[1] == "" {
		return newError(ErrBadSyntax, strField.Name, sTag, ErrorType)
	}
	reString := sTagValues[1]

	if isNil(value) || !value.IsValid() {
		return newError(ErrValueUnexpected, strField.Name, "<nil>", ErrorType)
	}

	sValue := fmt.Sprintf("%v", value.Interface())
	matched, err := regexp.MatchString(reString, sValue)
	if err != nil {
		return newError(err, strField.Name, sTag, ErrorType)
	}
	if !matched {
		return newError(ErrNoMatch, strField.Name, sTag, ErrorType)
	}
	return nil
}

func interfaceChecker(value reflect.Value) error {
	if !isInterface(value) {
		return nil
	}
	check, ok := value.Interface().(Checker)
	if !ok {
		return nil
	}
	return check.Check()
}

func checkValue(v Value, parent Value) []error {
	value := v.Value()
	if err := interfaceChecker(value); err != nil {
		return []error{err}
	}

	if v.Struct() == nil {
		return nil
	}

	sTag, ok := v.Tag().Lookup("check")
	if !ok {
		return nil
	}

	tagsChecks := strings.Split(sTag, ",")
	var result []error
	for _, tagCheck := range tagsChecks {
		var errCheck error
		switch tagCheck {
		case "required":
			errCheck = required(value, v.Struct())
		case "deprecated":
			errCheck = deprecated(value, v.Struct())
		default:
			if strings.HasPrefix(tagCheck, "expect:") {
				errCheck = expect(value, v.Struct())
			} else if strings.HasPrefix(tagCheck, "call:") {
				errCheck = withMethod(parent.Value(), value, v.Struct())
			} else if strings.HasPrefix(tagCheck, "re:") {
				errCheck = withRegexp(parent.Value(), value, v.Struct())
			} else {
				errCheck = fmt.Errorf("unknown check: %s", tagCheck)
			}
		}
		if errCheck != nil {
			result = append(result, errCheck)
		}
	}
	if len(result) == 0 {
		return nil
	}

	return result
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
func New(m Mode, e Type) *SimpeChecker {
	return &SimpeChecker{
		mode:      m,
		errorMode: e,
	}
}

//Check checks value
func (c *SimpeChecker) Check(v interface{}) []error {
	result := make([]error, 0)
	iter := newIterator(v)
	for iter.HasNext() {
		item := iter.Next()
		if err := checkValue(item, item.Parent()); err != nil {
			if len(err) != 0 && err[0] == ErrSkip {
				return nil
			}

			for _, e := range err {
				if are, ok := e.(ErrorCheckResult); ok {
					typ := are.GetType()
					if (typ & c.errorMode) != typ {
						continue
					}
				}
				result = append(result, e)
				if c.mode == ModeFirst {
					return result
				}
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
