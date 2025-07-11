// Package enum provides a generic Enum type for defining and working with enumerations.
package enum

import (
	"fmt"
)

type PureStringIntOrFloat interface {
	string | int | float32 | float64
}

// StringIntOrFloat is a constraint that allows underlying types of string, int, float32, and float64.
type StringIntOrFloat interface {
	~string | ~int | ~float32 | ~float64
}

// Enum is a generic type that represents an enumeration of values.
// T is the underlying type of the enumeration values.
// U is the underlying type of the values used for parsing and validation.
type Enum[T StringIntOrFloat, U PureStringIntOrFloat] struct {
	values []T
}

// NewEnum creates a new Enum with the given values.
func NewEnum[U PureStringIntOrFloat, T StringIntOrFloat](values ...T) Enum[T, U] {
	return Enum[T, U]{values: values}
}

// Options returns the list of valid options for the Enum.
func (e Enum[T, U]) Options() []T {
	return e.values
}

// Parse checks if the given value is a valid option for the Enum.
// It panics if the value is not valid.
// It returns it if it's true
func (e Enum[T, U]) Parse(value U) U {

	for _, v := range e.values {
		if v == any(value) {
			return value
		}
	}
	panic(fmt.Sprintf("invalid value %v it must be one of %v", value, e.values))
}

// Validate checks if the given value is a valid option for the Enum.
// It returns true if the value is valid, false otherwise.
func (e Enum[T, U]) Validate(value U) bool {

	for _, v := range e.values {
		if v == any(value) {
			return true
		}
	}

	return false
}
