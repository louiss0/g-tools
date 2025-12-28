package regex_extract

import (
	"regexp"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	testifyassert "github.com/stretchr/testify/assert"
)

func TestRegexExtract(t *testing.T) {
	RunSpecs(t, "Regex Extract Suite")
}

var describeExtractGroups = Describe("ExtractGroups", func() {
	var assertions *testifyassert.Assertions

	BeforeEach(func() {
		assertions = testifyassert.New(GinkgoT())
	})

	It("extracts named groups into a map", func() {
		regex := regexp.MustCompile(`^(?P<number>\d+)(?P<word>\w+)`)

		input := "100w"
		expected := map[string]string{
			"number": "100",
			"word":   "w",
		}
		actual, err := ExtractGroups(input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})

	It("preserves mixed-case group names", func() {
		regex := regexp.MustCompile(`^(?P<number>\d+)(?P<Word>\w+)`)

		input := "100w"
		expected := map[string]string{
			"number": "100",
			"Word":   "w",
		}
		actual, err := ExtractGroups(input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})

	It("returns an error when there is no match", func() {
		regex := regexp.MustCompile(`^(?P<NUMBER>\d+)(?P<WORD>\w+)`)

		input := "nope"
		_, err := ExtractGroups(input, regex)
		assertions.ErrorIs(err, ErrNoMatch)
	})
})

var describeExtractTypedNamedGroups = Describe("ExtractTypedNamedGroups", func() {
	var assertions *testifyassert.Assertions

	BeforeEach(func() {
		assertions = testifyassert.New(GinkgoT())
	})

	It("extracts typed values for named groups", func() {
		regex := regexp.MustCompile(`^(?P<word>\w+)-(?P<number>\d+)-(?P<float>\d+\.\d+)-(?P<dots>\.+)$`)

		input := "hello-42-3.14-..."
		expected := map[string]any{
			"word":   "hello",
			"number": uint8(42),
			"float":  float32(3.14),
			"dots":   "...",
		}

		actual, err := ExtractTypedNamedGroups(input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})

	It("extracts nested maps for named groups", func() {
		regex := regexp.MustCompile(`^(?P<outer>(?P<inner>\d+)-(?P<tail>\w+))$`)

		input := "123-abc"
		expected := map[string]any{
			"outer": map[string]any{
				"inner": uint8(123),
				"tail":  "abc",
			},
		}

		actual, err := ExtractTypedNamedGroups(input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})

	It("uses range-based integer types for named groups", func() {
		regex := regexp.MustCompile(`^(?P<small>\d+)-(?P<medium>\d+)-(?P<large>\d+)-(?P<huge>\d+)$`)

		input := "127-32767-2147483647-2147483648"
		expected := map[string]any{
			"small":  uint8(127),
			"medium": uint16(32767),
			"large":  uint32(2147483647),
			"huge":   uint32(2147483648),
		}

		actual, err := ExtractTypedNamedGroups(input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})

	It("uses range-based float types for named groups", func() {
		regex := regexp.MustCompile(`^(?P<small>\d+\.\d+)-(?P<large>\d+\.\d+)$`)

		input := "3.14-340282350000000000000000000000000000000.0"
		expected := map[string]any{
			"small": float32(3.14),
			"large": 340282350000000000000000000000000000000.0,
		}

		actual, err := ExtractTypedNamedGroups(input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})
})

var describeExtractTypedUnnamedGroups = Describe("ExtractTypedUnnamedGroups", func() {
	var assertions *testifyassert.Assertions

	BeforeEach(func() {
		assertions = testifyassert.New(GinkgoT())
	})

	It("extracts typed values for unnamed groups", func() {
		regex := regexp.MustCompile(`^(\w+)-(\d+)-(\d+\.\d+)-(\.+)$`)

		input := "hello-42-3.14-..."
		expected := []any{"hello", uint8(42), float32(3.14), "..."}

		actual, err := ExtractTypedUnnamedGroups(input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})

	It("extracts nested slices for unnamed groups", func() {
		regex := regexp.MustCompile(`^((\d+)-(\w+))$`)

		input := "123-abc"
		expected := []any{
			[]any{uint8(123), "abc"},
		}

		actual, err := ExtractTypedUnnamedGroups(input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})

	It("uses range-based integer types for unnamed groups", func() {
		regex := regexp.MustCompile(`^(\d+)-(\d+)-(\d+)-(\d+)$`)

		input := "127-32767-2147483647-2147483648"
		expected := []any{
			uint8(127),
			uint16(32767),
			uint32(2147483647),
			uint32(2147483648),
		}

		actual, err := ExtractTypedUnnamedGroups(input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})

	It("uses range-based float types for unnamed groups", func() {
		regex := regexp.MustCompile(`^(\d+\.\d+)-(\d+\.\d+)$`)

		input := "3.14-340282350000000000000000000000000000000.0"
		expected := []any{
			float32(3.14),
			340282350000000000000000000000000000000.0,
		}

		actual, err := ExtractTypedUnnamedGroups(input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})
})

var describeExtractToStruct = Describe("ExtractToStruct", func() {
	var assertions *testifyassert.Assertions

	BeforeEach(func() {
		assertions = testifyassert.New(GinkgoT())
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

	It("extracts typed values into struct fields", func() {
		regex := regexp.MustCompile(`^(?P<Word>\w+)-(?P<Number>\d+)-(?P<Float>\d+\.\d+)-(?P<Dots>\.+)$`)

		input := "hello-42-3.14-..."
		expected := Sample{
			Word:   "hello",
			Number: uint8(42),
			Float:  float32(3.14),
			Dots:   "...",
		}

		actual, err := ExtractToStruct[Sample](input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})

	It("extracts nested groups into nested structs", func() {
		regex := regexp.MustCompile(`^(?P<Outer>(?P<Inner>\d+)-(?P<Tail>\w+))$`)

		input := "123-abc"
		expected := Container{
			Outer: Nested{
				Inner: uint8(123),
				Tail:  "abc",
			},
		}

		actual, err := ExtractToStruct[Container](input, regex)
		assertions.NoError(err)
		assertions.Equal(expected, actual)
	})

	It("returns an error for invalid struct field names", func() {
		regex := regexp.MustCompile(`^(?P<word>\w+)$`)

		input := "hello"
		_, err := ExtractToStruct[Sample](input, regex)
		assertions.Error(err)
	})

	It("returns an error for missing struct fields", func() {
		regex := regexp.MustCompile(`^(?P<Missing>\w+)$`)

		input := "hello"
		_, err := ExtractToStruct[Sample](input, regex)
		assertions.Error(err)
	})
})
