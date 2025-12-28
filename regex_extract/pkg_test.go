package regex_extract

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractGroups(t *testing.T) {
	assert := assert.New(t)

	t.Run("Extracts named groups into a map", func(t *testing.T) {
		regex := regexp.MustCompile(`^(?P<number>\d+)(?P<word>\w+)`)

		input := "100w"
		expected := map[string]string{
			"number": "100",
			"word":   "w",
		}
		actual, err := ExtractGroups(input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})

	t.Run("Preserves mixed-case group names", func(t *testing.T) {
		regex := regexp.MustCompile(`^(?P<number>\d+)(?P<Word>\w+)`)

		input := "100w"
		expected := map[string]string{
			"number": "100",
			"Word":   "w",
		}
		actual, err := ExtractGroups(input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})

	t.Run("Returns an error when there is no match", func(t *testing.T) {
		regex := regexp.MustCompile(`^(?P<NUMBER>\d+)(?P<WORD>\w+)`)

		input := "nope"
		_, err := ExtractGroups(input, regex)
		assert.ErrorIs(err, ErrNoMatch)
	})

}

func TestExtractTypedNamedGroups(t *testing.T) {
	assert := assert.New(t)

	t.Run("Extracts typed values for named groups", func(t *testing.T) {
		regex := regexp.MustCompile(`^(?P<word>\w+)-(?P<number>\d+)-(?P<float>\d+\.\d+)-(?P<dots>\.+)$`)

		input := "hello-42-3.14-..."
		expected := map[string]any{
			"word":   "hello",
			"number": uint8(42),
			"float":  float32(3.14),
			"dots":   "...",
		}

		actual, err := ExtractTypedNamedGroups(input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})

	t.Run("Extracts nested maps for named groups", func(t *testing.T) {
		regex := regexp.MustCompile(`^(?P<outer>(?P<inner>\d+)-(?P<tail>\w+))$`)

		input := "123-abc"
		expected := map[string]any{
			"outer": map[string]any{
				"inner": uint8(123),
				"tail":  "abc",
			},
		}

		actual, err := ExtractTypedNamedGroups(input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})

	t.Run("Uses range-based integer types for named groups", func(t *testing.T) {
		regex := regexp.MustCompile(`^(?P<small>\d+)-(?P<medium>\d+)-(?P<large>\d+)-(?P<huge>\d+)$`)

		input := "127-32767-2147483647-2147483648"
		expected := map[string]any{
			"small":  uint8(127),
			"medium": uint16(32767),
			"large":  uint32(2147483647),
			"huge":   uint32(2147483648),
		}

		actual, err := ExtractTypedNamedGroups(input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})

	t.Run("Uses range-based float types for named groups", func(t *testing.T) {
		regex := regexp.MustCompile(`^(?P<small>\d+\.\d+)-(?P<large>\d+\.\d+)$`)

		input := "3.14-340282350000000000000000000000000000000.0"
		expected := map[string]any{
			"small": float32(3.14),
			"large": 340282350000000000000000000000000000000.0,
		}

		actual, err := ExtractTypedNamedGroups(input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})
}

func TestExtractTypedUnnamedGroups(t *testing.T) {
	assert := assert.New(t)

	t.Run("Extracts typed values for unnamed groups", func(t *testing.T) {
		regex := regexp.MustCompile(`^(\w+)-(\d+)-(\d+\.\d+)-(\.+)$`)

		input := "hello-42-3.14-..."
		expected := []any{"hello", uint8(42), float32(3.14), "..."}

		actual, err := ExtractTypedUnnamedGroups(input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})

	t.Run("Extracts nested slices for unnamed groups", func(t *testing.T) {
		regex := regexp.MustCompile(`^((\d+)-(\w+))$`)

		input := "123-abc"
		expected := []any{
			[]any{uint8(123), "abc"},
		}

		actual, err := ExtractTypedUnnamedGroups(input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})

	t.Run("Uses range-based integer types for unnamed groups", func(t *testing.T) {
		regex := regexp.MustCompile(`^(\d+)-(\d+)-(\d+)-(\d+)$`)

		input := "127-32767-2147483647-2147483648"
		expected := []any{
			uint8(127),
			uint16(32767),
			uint32(2147483647),
			uint32(2147483648),
		}

		actual, err := ExtractTypedUnnamedGroups(input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})

	t.Run("Uses range-based float types for unnamed groups", func(t *testing.T) {
		regex := regexp.MustCompile(`^(\d+\.\d+)-(\d+\.\d+)$`)

		input := "3.14-340282350000000000000000000000000000000.0"
		expected := []any{
			float32(3.14),
			340282350000000000000000000000000000000.0,
		}

		actual, err := ExtractTypedUnnamedGroups(input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})
}

func TestExtractToStruct(t *testing.T) {
	assert := assert.New(t)

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

	t.Run("Extracts typed values into struct fields", func(t *testing.T) {
		regex := regexp.MustCompile(`^(?P<Word>\w+)-(?P<Number>\d+)-(?P<Float>\d+\.\d+)-(?P<Dots>\.+)$`)

		input := "hello-42-3.14-..."
		expected := Sample{
			Word:   "hello",
			Number: uint8(42),
			Float:  float32(3.14),
			Dots:   "...",
		}

		actual, err := ExtractToStruct[Sample](input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})

	t.Run("Extracts nested groups into nested structs", func(t *testing.T) {
		regex := regexp.MustCompile(`^(?P<Outer>(?P<Inner>\d+)-(?P<Tail>\w+))$`)

		input := "123-abc"
		expected := Container{
			Outer: Nested{
				Inner: uint8(123),
				Tail:  "abc",
			},
		}

		actual, err := ExtractToStruct[Container](input, regex)
		assert.NoError(err)
		assert.Equal(expected, actual)
	})

	t.Run("Returns an error for invalid struct field names", func(t *testing.T) {
		regex := regexp.MustCompile(`^(?P<word>\w+)$`)

		input := "hello"
		_, err := ExtractToStruct[Sample](input, regex)
		assert.Error(err)
	})

	t.Run("Returns an error for missing struct fields", func(t *testing.T) {
		regex := regexp.MustCompile(`^(?P<Missing>\w+)$`)

		input := "hello"
		_, err := ExtractToStruct[Sample](input, regex)
		assert.Error(err)
	})
}
