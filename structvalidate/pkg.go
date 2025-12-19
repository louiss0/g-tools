package structvalidate

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// ObjectSchema describes the expected shape of a struct in a Zod-inspired, fluent API.
// It validates the presence of required fields, their types, and any additional
// constraints such as regex checks for strings or numeric ranges for integers and floats.
type ObjectSchema struct {
	fields map[string]fieldValidator
}

// Object constructs a new ObjectSchema from the provided field validators.
// Example:
//
// schema := Object(map[string]Field{
// "Name":  String().NonEmpty(),
// "Email": String().Regex(regexp.MustCompile(`^[^@]+@[^@]+$`)),
// "Age":   Int().Min(18),
// })
func Object(fields map[string]fieldValidator) *ObjectSchema {
	return &ObjectSchema{fields: fields}
}

// Validate checks the provided struct (or pointer to struct) against the schema.
// It returns an error describing the first validation failure or nil if the input satisfies
// all field requirements.
func (s *ObjectSchema) Validate(input any) error {
	if s == nil {
		return errors.New("validation failed: schema is nil")
	}

	if input == nil {
		return errors.New("validation failed: nil input provided")
	}

	value := reflect.ValueOf(input)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return errors.New("validation failed: nil pointer provided")
		}
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return fmt.Errorf("validation failed: expected struct, got %s", value.Kind())
	}

	for fieldName, validator := range s.fields {
		field := value.FieldByName(fieldName)
		if !field.IsValid() {
			if validator.IsOptional() {
				continue
			}
			return fmt.Errorf("validation failed: missing field %q", fieldName)
		}

		if err := validator.Validate(fieldName, field); err != nil {
			return err
		}
	}

	return nil
}

type fieldValidator interface {
	Validate(fieldName string, value reflect.Value) error
	IsOptional() bool
}

// StringSchema validates string fields with optional requirements such as non-empty or regex constraints.
type StringSchema struct {
	requireNonEmpty bool
	minLength       *int
	maxLength       *int
	ignoreSpaces    bool
	patterns        []stringPattern
	optional        bool
}

// String constructs a string validator.
func String() *StringSchema {
	return &StringSchema{ignoreSpaces: true}
}

// NonEmpty requires the string to be non-empty.
func (s *StringSchema) NonEmpty() *StringSchema {
	s.requireNonEmpty = true
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *StringSchema) Optional() *StringSchema {
	s.optional = true
	return s
}

// Regex attaches a regular expression requirement to the string validator.
func (s *StringSchema) Regex(exp *regexp.Regexp) *StringSchema {
	s.addPattern("required pattern", exp)
	return s
}

// MinLength requires the string length (excluding spaces by default) to be at least the provided value.
func (s *StringSchema) MinLength(length int) *StringSchema {
	s.minLength = &length
	return s
}

// MaxLength requires the string length (excluding spaces by default) to be at most the provided value.
func (s *StringSchema) MaxLength(length int) *StringSchema {
	s.maxLength = &length
	return s
}

// CountSpaces includes spaces in length calculations instead of omitting them.
func (s *StringSchema) CountSpaces() *StringSchema {
	s.ignoreSpaces = false
	return s
}

// Email validates the string against a basic email pattern.
func (s *StringSchema) Email() *StringSchema {
	s.addPattern("email", regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`))
	return s
}

// UUID validates the string as a UUID value.
func (s *StringSchema) UUID() *StringSchema {
	s.addPattern("uuid", regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`))
	return s
}

// Semver validates the string as a semantic version (with optional leading v, pre-release, and build metadata).
func (s *StringSchema) Semver() *StringSchema {
	s.addPattern("semver", regexp.MustCompile(`^v?(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`))
	return s
}

func (s *StringSchema) addPattern(name string, exp *regexp.Regexp) {
	s.patterns = append(s.patterns, stringPattern{name: name, exp: exp})
}

// Validate implements fieldValidator for StringSchema.
func (s *StringSchema) Validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.String {
		return fmt.Errorf("validation failed: field %q expected string, got %s", fieldName, value.Kind())
	}

	str := value.String()
	if s.requireNonEmpty && str == "" {
		return fmt.Errorf("validation failed: field %q must be non-empty", fieldName)
	}

	evaluated := str
	if s.ignoreSpaces {
		evaluated = strings.ReplaceAll(str, " ", "")
	}

	if s.minLength != nil && len(evaluated) < *s.minLength {
		return fmt.Errorf("validation failed: field %q length must be >= %d", fieldName, *s.minLength)
	}
	if s.maxLength != nil && len(evaluated) > *s.maxLength {
		return fmt.Errorf("validation failed: field %q length must be <= %d", fieldName, *s.maxLength)
	}

	for _, pattern := range s.patterns {
		if !pattern.exp.MatchString(str) {
			return fmt.Errorf("validation failed: field %q does not match %s pattern", fieldName, pattern.name)
		}
	}

	return nil
}

// IsOptional reports whether the string field may be omitted.
func (s *StringSchema) IsOptional() bool { return s.optional }

type stringPattern struct {
	name string
	exp  *regexp.Regexp
}

// IntSchema validates integer fields with optional minimum and maximum bounds.
type IntSchema struct {
	min      *int64
	max      *int64
	optional bool
}

// Int constructs an integer validator.
func Int() *IntSchema {
	return &IntSchema{}
}

// Min sets a minimum allowed value (inclusive).
func (s *IntSchema) Min(min int64) *IntSchema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *IntSchema) Max(max int64) *IntSchema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *IntSchema) Optional() *IntSchema {
	s.optional = true
	return s
}

// Validate implements fieldValidator for IntSchema.
func (s *IntSchema) Validate(fieldName string, value reflect.Value) error {
	if !isIntKind(value.Kind()) {
		return fmt.Errorf("validation failed: field %q expected integer, got %s", fieldName, value.Kind())
	}

	intVal := value.Int()
	if s.min != nil && intVal < *s.min {
		return fmt.Errorf("validation failed: field %q must be >= %d", fieldName, *s.min)
	}
	if s.max != nil && intVal > *s.max {
		return fmt.Errorf("validation failed: field %q must be <= %d", fieldName, *s.max)
	}

	return nil
}

// IsOptional reports whether the integer field may be omitted.
func (s *IntSchema) IsOptional() bool { return s.optional }

// FloatSchema validates floating point fields with optional minimum and maximum bounds.
type FloatSchema struct {
	min      *float64
	max      *float64
	optional bool
}

// Float constructs a float validator.
func Float() *FloatSchema {
	return &FloatSchema{}
}

// Min sets a minimum allowed value (inclusive).
func (s *FloatSchema) Min(min float64) *FloatSchema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *FloatSchema) Max(max float64) *FloatSchema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *FloatSchema) Optional() *FloatSchema {
	s.optional = true
	return s
}

// Validate implements fieldValidator for FloatSchema.
func (s *FloatSchema) Validate(fieldName string, value reflect.Value) error {
	if !isFloatKind(value.Kind()) {
		return fmt.Errorf("validation failed: field %q expected float, got %s", fieldName, value.Kind())
	}

	floatVal := value.Float()
	if s.min != nil && floatVal < *s.min {
		return fmt.Errorf("validation failed: field %q must be >= %g", fieldName, *s.min)
	}
	if s.max != nil && floatVal > *s.max {
		return fmt.Errorf("validation failed: field %q must be <= %g", fieldName, *s.max)
	}

	return nil
}

// IsOptional reports whether the float field may be omitted.
func (s *FloatSchema) IsOptional() bool { return s.optional }

// BoolSchema validates boolean fields.
type BoolSchema struct {
	optional bool
}

// Bool constructs a boolean validator.
func Bool() *BoolSchema { return &BoolSchema{} }

// Optional allows the field to be missing from the struct without producing an error.
func (b *BoolSchema) Optional() *BoolSchema {
	b.optional = true
	return b
}

// Validate implements fieldValidator for BoolSchema.
func (b *BoolSchema) Validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Bool {
		return fmt.Errorf("validation failed: field %q expected bool, got %s", fieldName, value.Kind())
	}
	return nil
}

// IsOptional reports whether the boolean field may be omitted.
func (b *BoolSchema) IsOptional() bool { return b.optional }

func isIntKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func isFloatKind(kind reflect.Kind) bool {
	return kind == reflect.Float32 || kind == reflect.Float64
}
