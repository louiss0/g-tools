package enum

import (
	"fmt"
	"testing"
)

func TestNewEnum(t *testing.T) {
	enum := NewEnum[string, string]("a", "b", "c")

	if len(enum.Options()) != 3 {
		t.Errorf("Expected 3 options, got %d", len(enum.Options()))
	}
}

func TestEnum_Options(t *testing.T) {
	enum := NewEnum[int, int](1, 2, 3)
	options := enum.Options()

	if len(options) != 3 {
		t.Fatalf("Expected 3 options, but got %d", len(options))
	}

	if options[0] != 1 || options[1] != 2 || options[2] != 3 {
		t.Errorf("Options are not as expected: %v", options)
	}
}

func TestEnum_Validate(t *testing.T) {
	enum := NewEnum[string, string]("a", "b", "c")

	if !enum.Validate("a") {
		t.Errorf("Expected 'a' to be valid, but it wasn't")
	}

	if enum.Validate("d") {
		t.Errorf("Expected 'd' to be invalid, but it was")
	}
}

func TestEnum_Parse(t *testing.T) {
	enum := NewEnum[int, int](1, 2, 3)

	// Test valid parse
	enum.Parse(1)

	// Test invalid parse
	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The code did not panic")
		}

		if msg, ok := r.(string); ok {
			if msg != "invalid value 4 it must be one of [1 2 3]" {
				t.Errorf("The code panicked with the wrong message: %v", msg)
			}
		} else {
			t.Errorf("The code panicked with an unexpected type: %T, value: %v", r, r)
		}
	}()
	enum.Parse(4)
}

func ExampleNewEnum() {
	enum := NewEnum[string, string]("a", "b", "c")
	fmt.Println(enum.Options())
	// Output: [a b c]
}

func ExampleEnum_Validate() {
	enum := NewEnum[int, int](1, 2, 3)
	valid := enum.Validate(2)
	fmt.Println(valid)
	// Output: true
}
