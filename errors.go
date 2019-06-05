package checks

import "fmt"

//Type errors
const (
	ErrorType Type = 1 << iota
	WarningType

	ErrorAll Type = ErrorType | WarningType
)

type (
	//Type errors
	Type int

	//ErrorCheck check
	ErrorCheck struct {
		cause error
		typ   Type
	}

	//ErrorCheckResult implements concrete check result
	ErrorCheckResult struct {
		ErrorCheck
		FieldName string
		Value     interface{}
	}
)

func newError(err error, field string, value interface{}, typ Type) ErrorCheckResult {
	return ErrorCheckResult{
		ErrorCheck: ErrorCheck{
			cause: err,
			typ:   typ,
		},
		FieldName: field,
		Value:     value,
	}
}

func (e ErrorCheck) Error() string {
	return e.cause.Error()
}

//GetType returns error type
func (e ErrorCheck) GetType() Type {
	return e.typ
}

func (e ErrorCheckResult) Error() string {
	if e.Value != nil {
		return fmt.Sprintf("%v: %s %v", e.cause, e.FieldName, e.Value)
	}
	return fmt.Sprintf("%v: %s", e.cause, e.FieldName)
}

//Filter returns filtred slice error by Type
func Filter(errors []error, typ Type) []error {
	result := make([]error, 0)
	for _, e := range errors {
		if are, ok := e.(ErrorCheckResult); !ok || ok && (are.GetType()&typ == typ) {
			result = append(result, e)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
