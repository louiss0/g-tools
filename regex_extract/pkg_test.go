package regex_extract

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

var tAssert *assert.Assertions

func TestRegexExtract(t *testing.T) {
	tAssert = assert.New(t)
	RunSpecs(t, "Regex Extract Suite")
}

var DescribeExtractGroups = Describe("ExtractGroups", func() {
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
	It("extracts typed values for named groups", func() {
		pattern := `^(?P<word>\w+)-(?P<number>\d+)-(?P<float>\d+\.\d+)-(?P<dots>\.+)$`

		input := "hello-42-3.14-..."
		expected := map[string]any{
			"word":   "hello",
			"number": uint8(42),
			"float":  float32(3.14),
			"dots":   "...",
		}

		actual, err := ExtractTypedNamedGroups[any](input, pattern)
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

		actual, err := ExtractTypedNamedGroups[any](input, pattern)
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

		actual, err := ExtractTypedNamedGroups[any](input, pattern)
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

		actual, err := ExtractTypedNamedGroups[any](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("returns an error when the type parameter is a struct", func() {
		pattern := `^(?P<Word>\w+)$`

		input := "hello"
		type Expected struct {
			Word string
		}

		_, err := ExtractTypedNamedGroups[Expected](input, pattern)
		tAssert.ErrorIs(err, ErrExpectedStruct)
	})

	It("returns an error when a value cannot convert to the target type", func() {
		pattern := `^(?P<word>\w+)$`

		input := "hello"
		_, err := ExtractTypedNamedGroups[uint8](input, pattern)
		tAssert.ErrorIs(err, ErrUnexpectedGroupValue)
	})

	It("supports string values", func() {
		pattern := `^(?P<word>\w+)$`

		input := "hello"
		expected := map[string]string{
			"word": "hello",
		}

		actual, err := ExtractTypedNamedGroups[string](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports int values", func() {
		pattern := `^(?P<count>\d+)$`

		input := "42"
		expected := map[string]int{
			"count": 42,
		}

		actual, err := ExtractTypedNamedGroups[int](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports int8 values", func() {
		pattern := `^(?P<count>\d+)$`

		input := "120"
		expected := map[string]int8{
			"count": int8(120),
		}

		actual, err := ExtractTypedNamedGroups[int8](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports int16 values", func() {
		pattern := `^(?P<count>\d+)$`

		input := "300"
		expected := map[string]int16{
			"count": int16(300),
		}

		actual, err := ExtractTypedNamedGroups[int16](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports int32 values", func() {
		pattern := `^(?P<count>\d+)$`

		input := "70000"
		expected := map[string]int32{
			"count": int32(70000),
		}

		actual, err := ExtractTypedNamedGroups[int32](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports int64 values", func() {
		pattern := `^(?P<count>\d+)$`

		input := "4294967296"
		expected := map[string]int64{
			"count": int64(4294967296),
		}

		actual, err := ExtractTypedNamedGroups[int64](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports uint values", func() {
		pattern := `^(?P<count>\d+)$`

		input := "42"
		expected := map[string]uint{
			"count": 42,
		}

		actual, err := ExtractTypedNamedGroups[uint](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports uint8 values", func() {
		pattern := `^(?P<count>\d+)$`

		input := "42"
		expected := map[string]uint8{
			"count": uint8(42),
		}

		actual, err := ExtractTypedNamedGroups[uint8](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports uint16 values", func() {
		pattern := `^(?P<count>\d+)$`

		input := "300"
		expected := map[string]uint16{
			"count": uint16(300),
		}

		actual, err := ExtractTypedNamedGroups[uint16](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports uint32 values", func() {
		pattern := `^(?P<count>\d+)$`

		input := "70000"
		expected := map[string]uint32{
			"count": uint32(70000),
		}

		actual, err := ExtractTypedNamedGroups[uint32](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports uint64 values", func() {
		pattern := `^(?P<count>\d+)$`

		input := "4294967296"
		expected := map[string]uint64{
			"count": uint64(4294967296),
		}

		actual, err := ExtractTypedNamedGroups[uint64](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports float32 values", func() {
		pattern := `^(?P<value>\d+\.\d+)$`

		input := "2.5"
		expected := map[string]float32{
			"value": float32(2.5),
		}

		actual, err := ExtractTypedNamedGroups[float32](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports float64 values", func() {
		pattern := `^(?P<value>\d+\.\d+)$`

		input := "340282350000000000000000000000000000000.0"
		expected := map[string]float64{
			"value": 340282350000000000000000000000000000000.0,
		}

		actual, err := ExtractTypedNamedGroups[float64](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})
})

var DescribeExtractTypedUnnamedGroups = Describe("ExtractTypedUnnamedGroups", func() {
	It("supports string values", func() {
		pattern := `^(\w+)$`

		input := "hello"
		expected := []string{"hello"}

		actual, err := ExtractTypedUnnamedGroups[string](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports int values", func() {
		pattern := `^(\d+)$`

		input := "42"
		expected := []int{42}

		actual, err := ExtractTypedUnnamedGroups[int](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports int8 values", func() {
		pattern := `^(\d+)$`

		input := "120"
		expected := []int8{int8(120)}

		actual, err := ExtractTypedUnnamedGroups[int8](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports int16 values", func() {
		pattern := `^(\d+)$`

		input := "300"
		expected := []int16{int16(300)}

		actual, err := ExtractTypedUnnamedGroups[int16](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports int32 values", func() {
		pattern := `^(\d+)$`

		input := "70000"
		expected := []int32{int32(70000)}

		actual, err := ExtractTypedUnnamedGroups[int32](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports int64 values", func() {
		pattern := `^(\d+)$`

		input := "4294967296"
		expected := []int64{int64(4294967296)}

		actual, err := ExtractTypedUnnamedGroups[int64](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports uint values", func() {
		pattern := `^(\d+)$`

		input := "42"
		expected := []uint{42}

		actual, err := ExtractTypedUnnamedGroups[uint](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports uint8 values", func() {
		pattern := `^(\d+)$`

		input := "42"
		expected := []uint8{uint8(42)}

		actual, err := ExtractTypedUnnamedGroups[uint8](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports uint16 values", func() {
		pattern := `^(\d+)$`

		input := "300"
		expected := []uint16{uint16(300)}

		actual, err := ExtractTypedUnnamedGroups[uint16](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports uint32 values", func() {
		pattern := `^(\d+)$`

		input := "70000"
		expected := []uint32{uint32(70000)}

		actual, err := ExtractTypedUnnamedGroups[uint32](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports uint64 values", func() {
		pattern := `^(\d+)$`

		input := "4294967296"
		expected := []uint64{uint64(4294967296)}

		actual, err := ExtractTypedUnnamedGroups[uint64](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports float32 values", func() {
		pattern := `^(\d+\.\d+)$`

		input := "2.5"
		expected := []float32{float32(2.5)}

		actual, err := ExtractTypedUnnamedGroups[float32](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports float64 values", func() {
		pattern := `^(\d+\.\d+)$`

		input := "340282350000000000000000000000000000000.0"
		expected := []float64{340282350000000000000000000000000000000.0}

		actual, err := ExtractTypedUnnamedGroups[float64](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports nested slices", func() {
		pattern := `^((\d+)-(\w+))$`

		input := "123-abc"
		expected := [][]any{
			{uint8(123), "abc"},
		}

		actual, err := ExtractTypedUnnamedGroups[[]any](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports deeper nested slices", func() {
		pattern := `^(((\d+)-(\w+)))$`

		input := "123-abc"
		expected := []([][]any){
			{
				{uint8(123), "abc"},
			},
		}

		actual, err := ExtractTypedUnnamedGroups[[][]any](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports struct values from nested captures", func() {
		pattern := `^((\d+)-(\w+))$`

		input := "123-abc"
		type Pair struct {
			Number uint8
			Word   string
		}
		expected := []Pair{
			{
				Number: uint8(123),
				Word:   "abc",
			},
		}

		actual, err := ExtractTypedUnnamedGroups[Pair](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})

	It("supports nested struct values from nested captures", func() {
		pattern := `^(((\d+)-(\w+)))$`

		input := "123-abc"
		type Pair struct {
			Number uint8
			Word   string
		}
		type Outer struct {
			Inner Pair
		}
		expected := []Outer{
			{
				Inner: Pair{
					Number: uint8(123),
					Word:   "abc",
				},
			},
		}

		actual, err := ExtractTypedUnnamedGroups[Outer](input, pattern)
		tAssert.NoError(err)
		tAssert.Equal(expected, actual)
	})
})

var DescribeExtractToStruct = Describe("ExtractToStruct", func() {
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
	It("returns an error for ExtractGroups", func() {
		_, err := ExtractGroups("value", "(")
		tAssert.ErrorIs(err, ErrInvalidRegex)
	})

	It("returns an error for ExtractSlice", func() {
		_, err := ExtractSlice[int]("value", "(")
		tAssert.ErrorIs(err, ErrInvalidRegex)
	})

	It("returns an error for ExtractTypedNamedGroups", func() {
		_, err := ExtractTypedNamedGroups[any]("value", "(")
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
		_, err := ExtractTypedUnnamedGroups[any]("value", "(")
		tAssert.ErrorIs(err, ErrInvalidRegex)
	})
})

func BenchmarkExtractGroups(b *testing.B) {
	pattern := `^(?P<number>\d+)(?P<word>\w+)`
	input := "100w"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractGroups(input, pattern)
	}
}

func BenchmarkExtractSlice(b *testing.B) {
	pattern := `^(\w+)-(\d+)$`
	input := "note-42"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractSlice[string](input, pattern)
	}
}

func BenchmarkExtractTypedNamedGroups(b *testing.B) {
	pattern := `^(?P<word>\w+)-(?P<number>\d+)-(?P<float>\d+\.\d+)$`
	input := "hello-42-3.14"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractTypedNamedGroups[any](input, pattern)
	}
}

func BenchmarkExtractTypedUnnamedGroups(b *testing.B) {
	pattern := `^(\d+)-(\w+)$`
	input := "123-abc"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractTypedUnnamedGroups[[]any](input, pattern)
	}
}

func BenchmarkExtractToStruct(b *testing.B) {
	pattern := `^(?P<Word>\w+)-(?P<Number>\d+)$`
	input := "hello-42"

	type Sample struct {
		Word   string
		Number uint8
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractToStruct[Sample](input, pattern)
	}
}

func BenchmarkMapValues(b *testing.B) {
	input := map[string]int{"a": 1, "b": 2, "c": 3}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = MapValues(input, strconv.Itoa)
	}
}

func BenchmarkCompileRegexValid(b *testing.B) {
	pattern := `^(?P<number>\d+)(?P<word>\w+)$`

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = compileRegex(pattern)
	}
}

func BenchmarkCompileRegexInvalid(b *testing.B) {
	pattern := "("

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = compileRegex(pattern)
	}
}

func BenchmarkParseCaptureTree(b *testing.B) {
	pattern := `^(?P<outer>(?P<inner>\d+)-((?P<tail>\w+)))$`

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = parseCaptureTree(pattern)
	}
}

func BenchmarkExtractGroupsVariants(b *testing.B) {
	pattern := `^(?P<number>\d+)(?P<word>\w+)$`
	input := "100w"

	b.Run("compile+extract", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = ExtractGroups(input, pattern)
		}
	})

	b.Run("precompiled", func(b *testing.B) {
		regex, err := compileRegex(pattern)
		if err != nil {
			b.Fatal(err)
		}

		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = extractGroupsWithRegex(input, regex)
		}
	})
}

func BenchmarkExtractGroupsNoMatch(b *testing.B) {
	pattern := `^(?P<number>\d+)(?P<word>\w+)$`
	input := "nope"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractGroups(input, pattern)
	}
}

func BenchmarkExtractGroupsInvalidName(b *testing.B) {
	pattern := `^(?P<bad-name>\d+)$`
	input := "10"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractGroups(input, pattern)
	}
}

func BenchmarkExtractTypedNamedGroupsStructType(b *testing.B) {
	pattern := `^(?P<word>\w+)$`
	input := "hello"
	type Expected struct {
		Word string
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractTypedNamedGroups[Expected](input, pattern)
	}
}

func extractGroupsWithRegex(input string, regex *regexp.Regexp) (map[string]string, error) {
	submatches := regex.FindStringSubmatch(input)
	if len(submatches) == 0 {
		return nil, ErrNoMatch
	}

	groupNames := regex.SubexpNames()
	extracted := make(map[string]string, len(groupNames))

	for index, name := range groupNames {
		if index == 0 || name == "" {
			continue
		}

		if strings.Contains(name, "-") {
			return nil, fmt.Errorf("%w: %s", ErrInvalidGroupName, name)
		}

		if _, exists := extracted[name]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateGroupName, name)
		}

		extracted[name] = submatches[index]
	}

	return extracted, nil
}
