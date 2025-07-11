package enum

import (
	"fmt"
)

type PureStringIntOrFloat interface {
	string | int | float32 | float64
}

type StringIntOrFloat interface {
	~string | ~int | ~float32 | ~float64
}

type Enum[T StringIntOrFloat, U PureStringIntOrFloat] struct {
	values []T
}

func NewEnum[U PureStringIntOrFloat, T StringIntOrFloat](values ...T) Enum[T, U] {
	return Enum[T, U]{values: values}
}

func (e Enum[T, U]) Options() []T {
	return e.values
}

func (e Enum[T, U]) Parse(value U) {

	for _, v := range e.values {
		if v == any(value) {
			return
		}
	}
	panic(fmt.Sprintf("invalid value %v it must be one of %v", value, e.values))
}

func (e Enum[T, U]) Validate(value U) bool {

	for _, v := range e.values {
		if v == any(value) {
			return true
		}
	}

	return false
}
