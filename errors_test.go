package checks

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCheckResult(t *testing.T) {
	err := errors.New("test")
	field := "F1"
	value := struct{}{}
	got := newError(err, field, value, WarningType)
	assert.NotNil(t, got)
	assert.Equal(t, field, got.FieldName)
	assert.Equal(t, value, got.Value)
	assert.Equal(t, err, got.cause)
	assert.Equal(t, WarningType, got.typ)
	assert.Equal(t, got.typ, got.GetType())
	assert.EqualError(t, got.ErrorCheck, "test")
	assert.EqualError(t, got, "test: F1 {}")
	got.Value = nil
	assert.EqualError(t, got, "test: F1")
}

func TestErrorFilter(t *testing.T) {
	err := errors.New("test")
	errors := []error{
		err,
		newError(err, "f1", "123", WarningType),
		newError(err, "f2", 333, ErrorType),
	}

	got := Filter(errors, WarningType)
	assert.Len(t, got, 2)
	assert.EqualError(t, got[0], err.Error())
	assert.EqualError(t, got[1], "test: f1 123")

	got = Filter(errors, ErrorType)
	assert.Len(t, got, 2)
	assert.EqualError(t, got[0], err.Error())
	assert.EqualError(t, got[1], "test: f2 333")

	errors = []error{
		newError(err, "f2", 333, ErrorType),
	}
	got = Filter(errors, ErrorType)
	assert.Len(t, got, 1)
	assert.EqualError(t, got[0], "test: f2 333")

	got = Filter(errors, WarningType)
	assert.Len(t, got, 0)

}
