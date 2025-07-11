package enum

import (
	"fmt"
	"testing"
)

func TestNewEnum(t *testing.T) {
	enum := NewEnum[string]("a", "b", "c")

	if len(enum.Options()) != 3 {
		t.Errorf("Expected 3 options, got %d", len(enum.Options()))
	}
}

func TestEnum_Options(t *testing.T) {
	enum := NewEnum[int](1, 2, 3)
	options := enum.Options()

	if len(options) != 3 {
		t.Fatalf("Expected 3 options, but got %d", len(options))
	}

	if options[0] != 1 || options[1] != 2 || options[2] != 3 {
		t.Errorf("Options are not as expected: %v", options)
	}
}

func TestEnum_Validate(t *testing.T) {
	enum := NewEnum[string]("a", "b", "c")

	if !enum.Validate("a") {
		t.Errorf("Expected 'a' to be valid, but it wasn't")
	}

	if enum.Validate("d") {
		t.Errorf("Expected 'd' to be invalid, but it was")
	}
}

func TestEnum_Parse(t *testing.T) {
	enum := NewEnum[int](1, 2, 3)

	// Test valid parse
	val, err := enum.Parse(1)
	if err != nil {
		t.Errorf("Parse of valid value returned an unexpected error: %v", err)
	}
	if val != 1 {
		t.Errorf("Parse of valid value returned wrong value. Expected: %v, Got: %v", 1, val)
	}

	// Test invalid parse
	val, err = enum.Parse(4)
	if err == nil {
		t.Errorf("Parse of invalid value did not return an error")
	}
	var zeroU int // Zero value for int (U is int here)
	if val != zeroU {
		t.Errorf("Parse of invalid value returned non-zero value. Expected: %v, Got: %v", zeroU, val)
	}
	expectedError := "invalid value 4; it must be one of [1 2 3]"
	if err.Error() != expectedError {
		t.Errorf("Parse of invalid value returned wrong error message.\nExpected: %q\nGot: %q", expectedError, err.Error())
	}
}

func ExampleNewEnum() {
	enum := NewEnum[string]("a", "b", "c")
	fmt.Println(enum.Options())
	// Output: [a b c]
}

func Example_enum_Options() {
	enum := NewEnum[int](10, 20, 30)
	options := enum.Options()
	fmt.Println(options)
	// Output: [10 20 30]
}

func Example_enum_Validate() {
	enum := NewEnum[int](1, 2, 3)
	valid := enum.Validate(2)
	fmt.Println(valid)
	// Output: true
}

func Example_enum_Parse() {
	enum := NewEnum[string]("apple", "banana", "cherry")
	value, err := enum.Parse("banana")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(value)
	}
	// Output: banana
}

type Color string

func ExampleNewEnum_type_alias() {
	enum := NewEnum[string, Color]("red", "green", "blue")
	fmt.Println(enum.Options())
	// Output: [red green blue]
}

func Example_enum_Options_type_alias() {
	enum := NewEnum[string, Color]("red", "green", "blue")
	options := enum.Options()
	fmt.Println(options)
	// Output: [red green blue]
}

func Example_enum_Validate_type_alias() {
	enum := NewEnum[string, Color]("red", "green", "blue")
	valid := enum.Validate("green")
	fmt.Println(valid)
	// Output: true
}

func Example_enum_Parse_type_alias() {
	enum := NewEnum[string, Color]("red", "green", "blue")
	value, err := enum.Parse("blue")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println(value)
	}
	// Output: blue
}
