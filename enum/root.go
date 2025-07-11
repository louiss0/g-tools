// Package enum provides a generic Enum type for defining and working with enumerations.
package enum

import (
	"fmt"
	"reflect"
)

// PureStringIntOrFloat is a constraint that allows underlying types of string, int, float32, and float64,
// including type aliases of these basic types.
type PureStringIntOrFloat interface {
	~string | ~int | ~float32 | ~float64
}

type StringIntOrFloat interface {
	string | int | float32 | float64
}

// enum is a generic type that represents an enumeration of values.
// T is the underlying type of the enumeration values.
// U is the underlying type of the values used for parsing and validation.
type enum[value PureStringIntOrFloat, input StringIntOrFloat] struct {
	values []value
}

// NewEnum creates a new Enum with the given values.
func NewEnum[input StringIntOrFloat, value PureStringIntOrFloat](values ...value) enum[value, input] {
	return enum[value, input]{values: values}
}

// Options returns the list of valid options for the Enum.
func (e enum[value, input]) Options() []value {
	return e.values
}

// Parse checks if the given value is a valid option for the Enum.
// It returns the value if valid, or the zero value of U and an error if not valid.
func (e enum[T, U]) Parse(input U) (T, error) {

	inputReflection := reflect.ValueOf(input)

	for _, v := range e.values {

		valueRelection := reflect.ValueOf(v)

		switch valueRelection.Kind() {

		case reflect.String:
			if inputReflection.String() == valueRelection.String() {
				return valueRelection.Interface().(T), nil
			}
		case reflect.Int:
			if inputReflection.Int() == valueRelection.Int() {
				return valueRelection.Interface().(T), nil
			}
		case reflect.Float32:
			if inputReflection.Float() == valueRelection.Float() {
				return valueRelection.Interface().(T), nil
			}
		case reflect.Float64:
			if inputReflection.Float() == valueRelection.Float() {
				return valueRelection.Interface().(T), nil
			}
		}

	}

	return *new(T), fmt.Errorf("invalid value %v; it must be one of %v", input, e.values)
}

// Validate checks if the given value is a valid option for the Enum.
// It returns true if the value is valid, false otherwise.
func (e enum[T, U]) Validate(input U) bool {

	inputReflection := reflect.ValueOf(input)

	for _, v := range e.values {

		valueRelection := reflect.ValueOf(v)

		switch valueRelection.Kind() {

		case reflect.String:
			if inputReflection.String() == valueRelection.String() {
				return true
			}
		case reflect.Int:
			if inputReflection.Int() == valueRelection.Int() {
				return true
			}
		case reflect.Float32:
			if inputReflection.Float() == valueRelection.Float() {
				return true
			}
		case reflect.Float64:
			if inputReflection.Float() == valueRelection.Float() {
				return true
			}
		}

	}

	return false
}
