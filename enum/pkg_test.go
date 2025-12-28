package enum

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var tAssert *assert.Assertions

func TestEnum(t *testing.T) {
	tAssert = assert.New(t)
	RunSpecs(t, "Enum Suite")
}

var DescribeNewEnum = Describe("NewEnum", func() {
	It("creates an enum with the provided options", func() {
		enum := NewEnum[string]("a", "b", "c")

		tAssert.Len(enum.Options(), 3)
	})
})

var DescribeEnumOptions = Describe("Enum.Options", func() {
	It("returns options in the provided order", func() {
		enum := NewEnum[int](1, 2, 3)
		options := enum.Options()

		tAssert.Equal([]int{1, 2, 3}, options)
	})
})

var DescribeEnumValidate = Describe("Enum.Validate", func() {
	It("validates whether a value is included", func() {
		enum := NewEnum[string]("a", "b", "c")

		tAssert.True(enum.Validate("a"))
		tAssert.False(enum.Validate("d"))
	})
})

var DescribeEnumParse = Describe("Enum.Parse", func() {
	It("parses values and returns errors for invalid input", func() {
		enum := NewEnum[int](1, 2, 3)

		value, err := enum.Parse(1)
		tAssert.NoError(err)
		tAssert.Equal(1, value)

		value, err = enum.Parse(4)
		hasError := tAssert.Error(err)
		tAssert.Equal(0, value)

		if hasError {
			expectedError := "invalid value 4; it must be one of [1 2 3]"
			tAssert.Equal(expectedError, err.Error())
		}
	})
})

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
