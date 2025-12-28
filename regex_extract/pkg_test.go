package regex_extract

import (
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	testifyassert "github.com/stretchr/testify/assert"
)

var tAssert *testifyassert.Assertions

func TestRegexExtract(t *testing.T) {
	RunSpecs(t, "Regex Extract Suite")
}

var DescribeExtractGroups = Describe("ExtractGroups", func() {
	BeforeEach(func() {
		tAssert = testifyassert.New(GinkgoT())
	})

	It("extracts named groups into a map", func() {
		pattern := `^(?P<number>\d+)(?P<word>\w+)`

		input := "100w"
		expected := map[string]string{
			"number": "100",
			"word":   "w",
		}
		actual, err := ExtractGroups(input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("preserves mixed-case group names", func() {
		pattern := `^(?P<number>\d+)(?P<Word>\w+)`

		input := "100w"
		expected := map[string]string{
			"number": "100",
			"Word":   "w",
		}
		actual, err := ExtractGroups(input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("returns an error when there is no match", func() {
		pattern := `^(?P<NUMBER>\d+)(?P<WORD>\w+)`

		input := "nope"
		_, err := ExtractGroups(input, pattern)
		tAssert.ErrorIs(err, ErrNoMatch)
	})
})

var DescribeExtractSlice = Describe("ExtractSlice", func() {
	BeforeEach(func() {
		tAssert = testifyassert.New(GinkgoT())
	})

	It("extracts string groups into a slice", func() {
		pattern := `^(\w+)-(\d+)$`

		input := "note-42"
		expected := []string{"note", "42"}

		actual, err := ExtractSlice[string](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("extracts numeric groups into a typed slice", func() {
		pattern := `^(\d+)-(\d+)$`

		input := "10-20"
		expected := []uint8{10, 20}

		actual, err := ExtractSlice[uint8](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("returns an error when conversion fails", func() {
		pattern := `^(\d+)$`

		input := "300"
		_, err := ExtractSlice[uint8](input, pattern)
		tAssert.Error(err)
	})

	It("returns an error when there is no match", func() {
		pattern := `^(\d+)$`

		input := "nope"
		_, err := ExtractSlice[int](input, pattern)
		tAssert.ErrorIs(err, ErrNoMatch)
	})
})

var DescribeExtractTypedNamedGroups = Describe("ExtractTypedNamedGroups", func() {
	BeforeEach(func() {
		tAssert = testifyassert.New(GinkgoT())
	})

	It("extracts typed values for named groups", func() {
		pattern := `^(?P<word>\w+)-(?P<number>\d+)-(?P<float>\d+\.\d+)-(?P<dots>\.+)$`

		input := "hello-42-3.14-..."
		expected := map[string]any{
			"word":   "hello",
			"number": uint8(42),
			"float":  float32(3.14),
			"dots":   "...",
		}

		actual, err := ExtractTypedNamedGroups(input, pattern, expected)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("extracts nested maps for named groups", func() {
		pattern := `^(?P<outer>(?P<inner>\d+)-(?P<tail>\w+))$`

		input := "123-abc"
		expected := map[string]any{
			"outer": map[string]any{
				"inner": uint8(123),
				"tail":  "abc",
			},
		}

		actual, err := ExtractTypedNamedGroups(input, pattern, expected)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("uses range-based integer types for named groups", func() {
		pattern := `^(?P<small>\d+)-(?P<medium>\d+)-(?P<large>\d+)-(?P<huge>\d+)$`

		input := "127-32767-2147483647-2147483648"
		expected := map[string]any{
			"small":  uint8(127),
			"medium": uint16(32767),
			"large":  uint32(2147483647),
			"huge":   uint32(2147483648),
		}

		actual, err := ExtractTypedNamedGroups(input, pattern, expected)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("uses range-based float types for named groups", func() {
		pattern := `^(?P<small>\d+\.\d+)-(?P<large>\d+\.\d+)$`

		input := "3.14-340282350000000000000000000000000000000.0"
		expected := map[string]any{
			"small": float32(3.14),
			"large": 340282350000000000000000000000000000000.0,
		}

		actual, err := ExtractTypedNamedGroups(input, pattern, expected)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("returns an error when expected values do not match", func() {
		pattern := `^(?P<word>\w+)-(?P<number>\d+)$`

		input := "hello-42"
		expected := map[string]any{
			"word":   "hello",
			"number": uint8(41),
		}

		_, err := ExtractTypedNamedGroups(input, pattern, expected)
		tAssert.ErrorIs(err, ErrUnexpectedGroupValue)
	})

	It("returns an error when expected values are a struct", func() {
		pattern := `^(?P<word>\w+)$`

		input := "hello"
		expected := struct {
			Word string
		}{
			Word: "hello",
		}

		_, err := ExtractTypedNamedGroups(input, pattern, expected)
		tAssert.ErrorIs(err, ErrExpectedStruct)
	})
})

var DescribeExtractTypedUnnamedGroups = Describe("ExtractTypedUnnamedGroups", func() {
	BeforeEach(func() {
		tAssert = testifyassert.New(GinkgoT())
	})

	It("extracts typed values for unnamed groups", func() {
		pattern := `^(\w+)-(\d+)-(\d+\.\d+)-(\.+)$`

		input := "hello-42-3.14-..."
		expected := []any{"hello", uint8(42), float32(3.14), "..."}

		actual, err := ExtractTypedUnnamedGroups(input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("extracts nested slices for unnamed groups", func() {
		pattern := `^((\d+)-(\w+))$`

		input := "123-abc"
		expected := []any{
			[]any{uint8(123), "abc"},
		}

		actual, err := ExtractTypedUnnamedGroups(input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("uses range-based integer types for unnamed groups", func() {
		pattern := `^(\d+)-(\d+)-(\d+)-(\d+)$`

		input := "127-32767-2147483647-2147483648"
		expected := []any{
			uint8(127),
			uint16(32767),
			uint32(2147483647),
			uint32(2147483648),
		}

		actual, err := ExtractTypedUnnamedGroups(input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("uses range-based float types for unnamed groups", func() {
		pattern := `^(\d+\.\d+)-(\d+\.\d+)$`

		input := "3.14-340282350000000000000000000000000000000.0"
		expected := []any{
			float32(3.14),
			340282350000000000000000000000000000000.0,
		}

		actual, err := ExtractTypedUnnamedGroups(input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})
})

var DescribeExtractToStruct = Describe("ExtractToStruct", func() {
	BeforeEach(func() {
		tAssert = testifyassert.New(GinkgoT())
	})

	type Sample struct {
		Word   string
		Number uint8
		Float  float32
		Dots   string
	}

	type Nested struct {
		Inner uint8
		Tail  string
	}

	type Container struct {
		Outer Nested
	}

	type DeepInner struct {
		Value uint8
	}

	type DeepOuter struct {
		Inner DeepInner
	}

	type DeepContainer struct {
		Outer DeepOuter
	}

	type SiblingInner struct {
		Tail string
	}

	type SiblingContainer struct {
		First  uint8
		Second SiblingInner
	}

	It("extracts typed values into struct fields", func() {
		pattern := `^(?P<Word>\w+)-(?P<Number>\d+)-(?P<Float>\d+\.\d+)-(?P<Dots>\.+)$`

		input := "hello-42-3.14-..."
		expected := Sample{
			Word:   "hello",
			Number: uint8(42),
			Float:  float32(3.14),
			Dots:   "...",
		}

		actual, err := ExtractToStruct[Sample](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("extracts nested groups into nested structs", func() {
		pattern := `^(?P<Outer>(?P<Inner>\d+)-(?P<Tail>\w+))$`

		input := "123-abc"
		expected := Container{
			Outer: Nested{
				Inner: uint8(123),
				Tail:  "abc",
			},
		}

		actual, err := ExtractToStruct[Container](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("extracts nested groups wrapped in unnamed captures", func() {
		pattern := `^((?P<Outer>(?P<Inner>\d+)-(?P<Tail>\w+)))$`

		input := "123-abc"
		expected := Container{
			Outer: Nested{
				Inner: uint8(123),
				Tail:  "abc",
			},
		}

		actual, err := ExtractToStruct[Container](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("extracts deeply nested groups into nested structs", func() {
		pattern := `^(?P<Outer>(?P<Inner>(?P<Value>\d+)))$`

		input := "42"
		expected := DeepContainer{
			Outer: DeepOuter{
				Inner: DeepInner{
					Value: uint8(42),
				},
			},
		}

		actual, err := ExtractToStruct[DeepContainer](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("extracts nested groups from mixed root captures", func() {
		pattern := `^(?P<First>\d+)-((?P<Second>(?P<Tail>\w+)))$`

		input := "7-abc"
		expected := SiblingContainer{
			First: uint8(7),
			Second: SiblingInner{
				Tail: "abc",
			},
		}

		actual, err := ExtractToStruct[SiblingContainer](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("returns an error for invalid struct field names", func() {
		pattern := `^(?P<word>\w+)$`

		input := "hello"
		_, err := ExtractToStruct[Sample](input, pattern)
		tAssert.Error(err)
	})

	It("returns an error for missing struct fields", func() {
		pattern := `^(?P<Missing>\w+)$`

		input := "hello"
		_, err := ExtractToStruct[Sample](input, pattern)
		tAssert.Error(err)
	})
})

var DescribeMapValues = Describe("MapValues", func() {
	BeforeEach(func() {
		tAssert = testifyassert.New(GinkgoT())
	})

	It("maps values while preserving keys", func() {
		input := map[string]int{"a": 1, "b": 2}
		expected := map[string]string{"a": "1", "b": "2"}

		actual := MapValues(input, strconv.Itoa)
		tAssert.Equal(expected, actual)
	})

	It("returns nil when input is nil", func() {
		actual := MapValues[string, int, int](nil, func(value int) int {
			return value
		})
		tAssert.Nil(actual)
	})

	It("returns an empty map when input is empty", func() {
		actual := MapValues(map[string]int{}, func(value int) int {
			return value * 2
		})
		tAssert.Empty(actual)
	})
})

var DescribeInvalidRegex = Describe("InvalidRegex", func() {
	BeforeEach(func() {
		tAssert = testifyassert.New(GinkgoT())
	})

	It("returns an error for ExtractGroups", func() {
		_, err := ExtractGroups("value", "(")
		tAssert.ErrorIs(err, ErrInvalidRegex)
	})

	It("returns an error for ExtractSlice", func() {
		_, err := ExtractSlice[int]("value", "(")
		tAssert.ErrorIs(err, ErrInvalidRegex)
	})

	It("returns an error for ExtractTypedNamedGroups", func() {
		_, err := ExtractTypedNamedGroups("value", "(", map[string]any{"value": "value"})
		tAssert.ErrorIs(err, ErrInvalidRegex)
	})

	It("returns an error for ExtractToStruct", func() {
		type Sample struct {
			Word string
		}

		_, err := ExtractToStruct[Sample]("value", "(")
		tAssert.ErrorIs(err, ErrInvalidRegex)
	})

	It("returns an error for ExtractTypedUnnamedGroups", func() {
		_, err := ExtractTypedUnnamedGroups("value", "(")
		tAssert.ErrorIs(err, ErrInvalidRegex)
	})
})
