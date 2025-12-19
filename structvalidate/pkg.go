package structvalidate

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var validationErrorName = "G-ToolValidationError"

// ParseError wraps validation failures for the Parse helper.
type ParseError struct {
	err error
}

// Error implements the error interface.
func (e *ParseError) Error() string { return e.err.Error() }

// Unwrap exposes the underlying validation error.
func (e *ParseError) Unwrap() error { return e.err }

func validationError(format string, args ...any) error {
	return fmt.Errorf("%s: %s", validationErrorName, fmt.Sprintf(format, args...))
}

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
// "Email": String().Regex(`^[^@]+@[^@]+$`),
// "Age":   Int().Min(18),
// })
func Object(fields map[string]fieldValidator) *ObjectSchema {
	return &ObjectSchema{fields: fields}
}

// validate checks the provided struct (or pointer to struct) against the schema.
// It returns an error describing the first validation failure or nil if the input satisfies
// all field requirements.
func (s *ObjectSchema) validate(input any) error {
	if s == nil {
		return validationError("validation failed: schema is nil")
	}

	if input == nil {
		return validationError("validation failed: nil input provided")
	}

	value := reflect.ValueOf(input)
	if value.Kind() == reflect.Ptr {
		if value.IsNil() {
			return validationError("validation failed: nil pointer provided")
		}
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return validationError("validation failed: expected struct, got %s", value.Kind())
	}

	for fieldName, validator := range s.fields {
		field := value.FieldByName(fieldName)
		if !field.IsValid() {
			if validator.isOptional() {
				continue
			}
			return validationError("validation failed: missing field %q", fieldName)
		}

		if err := validator.validate(fieldName, field); err != nil {
			return err
		}
	}

	return nil
}

// Parse validates the input and returns a ParseError when validation fails.
func (s *ObjectSchema) Parse(input any) *ParseError {
	if err := s.validate(input); err != nil {
		return &ParseError{err: errors.New(err.Error())}
	}
	return nil
}

// MustParse validates the input and panics if validation fails.
func (s *ObjectSchema) MustParse(input any) {
	if err := s.Parse(input); err != nil {
		panic(err)
	}
}

type fieldValidator interface {
	validate(fieldName string, value reflect.Value) error
	isOptional() bool
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
// The provided pattern will be compiled with regexp.MustCompile.
func (s *StringSchema) Regex(pattern string) *StringSchema {
	s.addPattern("required pattern", regexp.MustCompile(pattern))
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

// validate implements fieldValidator for StringSchema.
func (s *StringSchema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.String {
		return validationError("validation failed: field %q expected string, got %s", fieldName, value.Kind())
	}

	str := value.String()
	if s.requireNonEmpty && str == "" {
		return validationError("validation failed: field %q must be non-empty", fieldName)
	}

	evaluated := str
	if s.ignoreSpaces {
		evaluated = strings.ReplaceAll(str, " ", "")
	}

	if s.minLength != nil && len(evaluated) < *s.minLength {
		return validationError("validation failed: field %q length must be >= %d", fieldName, *s.minLength)
	}
	if s.maxLength != nil && len(evaluated) > *s.maxLength {
		return validationError("validation failed: field %q length must be <= %d", fieldName, *s.maxLength)
	}

	for _, pattern := range s.patterns {
		if !pattern.exp.MatchString(str) {
			return validationError("validation failed: field %q does not match %s pattern", fieldName, pattern.name)
		}
	}

	return nil
}

// isOptional reports whether the string field may be omitted.
func (s *StringSchema) isOptional() bool { return s.optional }

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

// validate implements fieldValidator for IntSchema.
func (s *IntSchema) validate(fieldName string, value reflect.Value) error {
	kind := value.Kind()
	if !isIntKind(kind) {
		return validationError("validation failed: field %q expected integer, got %s", fieldName, value.Kind())
	}

	if isSignedIntKind(kind) {
		intVal := value.Int()
		if s.min != nil && intVal < *s.min {
			return validationError("validation failed: field %q must be >= %d", fieldName, *s.min)
		}
		if s.max != nil && intVal > *s.max {
			return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
		}
		return nil
	}

	uintVal := value.Uint()
	if s.min != nil {
		if *s.min < 0 {
			// unsigned values are always >= negative minimums
		} else if uintVal < uint64(*s.min) {
			return validationError("validation failed: field %q must be >= %d", fieldName, *s.min)
		}
	}
	if s.max != nil {
		if *s.max < 0 {
			return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
		}
		if uintVal > uint64(*s.max) {
			return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
		}
	}

	return nil
}

// isOptional reports whether the integer field may be omitted.
func (s *IntSchema) isOptional() bool { return s.optional }

// Int8Schema validates int8 fields with optional minimum and maximum bounds.
type Int8Schema struct {
	min      *int8
	max      *int8
	optional bool
}

// Int8 constructs an int8 validator.
func Int8() *Int8Schema { return &Int8Schema{} }

// Min sets a minimum allowed value (inclusive).
func (s *Int8Schema) Min(min int8) *Int8Schema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *Int8Schema) Max(max int8) *Int8Schema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *Int8Schema) Optional() *Int8Schema {
	s.optional = true
	return s
}

// validate implements fieldValidator for Int8Schema.
func (s *Int8Schema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Int8 {
		return validationError("validation failed: field %q expected int8, got %s", fieldName, value.Kind())
	}

	val := value.Int()
	if s.min != nil && val < int64(*s.min) {
		return validationError("validation failed: field %q must be >= %d", fieldName, *s.min)
	}
	if s.max != nil && val > int64(*s.max) {
		return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the int8 field may be omitted.
func (s *Int8Schema) isOptional() bool { return s.optional }

// Int16Schema validates int16 fields with optional minimum and maximum bounds.
type Int16Schema struct {
	min      *int16
	max      *int16
	optional bool
}

// Int16 constructs an int16 validator.
func Int16() *Int16Schema { return &Int16Schema{} }

// Min sets a minimum allowed value (inclusive).
func (s *Int16Schema) Min(min int16) *Int16Schema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *Int16Schema) Max(max int16) *Int16Schema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *Int16Schema) Optional() *Int16Schema {
	s.optional = true
	return s
}

// validate implements fieldValidator for Int16Schema.
func (s *Int16Schema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Int16 {
		return validationError("validation failed: field %q expected int16, got %s", fieldName, value.Kind())
	}

	val := value.Int()
	if s.min != nil && val < int64(*s.min) {
		return validationError("validation failed: field %q must be >= %d", fieldName, *s.min)
	}
	if s.max != nil && val > int64(*s.max) {
		return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the int16 field may be omitted.
func (s *Int16Schema) isOptional() bool { return s.optional }

// Int32Schema validates int32 fields with optional minimum and maximum bounds.
type Int32Schema struct {
	min      *int32
	max      *int32
	optional bool
}

// Int32 constructs an int32 validator.
func Int32() *Int32Schema { return &Int32Schema{} }

// Min sets a minimum allowed value (inclusive).
func (s *Int32Schema) Min(min int32) *Int32Schema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *Int32Schema) Max(max int32) *Int32Schema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *Int32Schema) Optional() *Int32Schema {
	s.optional = true
	return s
}

// validate implements fieldValidator for Int32Schema.
func (s *Int32Schema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Int32 {
		return validationError("validation failed: field %q expected int32, got %s", fieldName, value.Kind())
	}

	val := value.Int()
	if s.min != nil && val < int64(*s.min) {
		return validationError("validation failed: field %q must be >= %d", fieldName, *s.min)
	}
	if s.max != nil && val > int64(*s.max) {
		return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the int32 field may be omitted.
func (s *Int32Schema) isOptional() bool { return s.optional }

// Int64Schema validates int64 fields with optional minimum and maximum bounds.
type Int64Schema struct {
	min      *int64
	max      *int64
	optional bool
}

// Int64 constructs an int64 validator.
func Int64() *Int64Schema { return &Int64Schema{} }

// Min sets a minimum allowed value (inclusive).
func (s *Int64Schema) Min(min int64) *Int64Schema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *Int64Schema) Max(max int64) *Int64Schema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *Int64Schema) Optional() *Int64Schema {
	s.optional = true
	return s
}

// validate implements fieldValidator for Int64Schema.
func (s *Int64Schema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Int64 {
		return validationError("validation failed: field %q expected int64, got %s", fieldName, value.Kind())
	}

	val := value.Int()
	if s.min != nil && val < *s.min {
		return validationError("validation failed: field %q must be >= %d", fieldName, *s.min)
	}
	if s.max != nil && val > *s.max {
		return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the int64 field may be omitted.
func (s *Int64Schema) isOptional() bool { return s.optional }

// UintSchema validates uint fields with optional minimum and maximum bounds.
type UintSchema struct {
	min      *uint64
	max      *uint64
	optional bool
}

// Uint constructs a uint validator.
func Uint() *UintSchema { return &UintSchema{} }

// Min sets a minimum allowed value (inclusive).
func (s *UintSchema) Min(min uint64) *UintSchema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *UintSchema) Max(max uint64) *UintSchema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *UintSchema) Optional() *UintSchema {
	s.optional = true
	return s
}

// validate implements fieldValidator for UintSchema.
func (s *UintSchema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Uint {
		return validationError("validation failed: field %q expected uint, got %s", fieldName, value.Kind())
	}

	val := value.Uint()
	if s.min != nil && val < *s.min {
		return validationError("validation failed: field %q must be >= %d", fieldName, *s.min)
	}
	if s.max != nil && val > *s.max {
		return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the uint field may be omitted.
func (s *UintSchema) isOptional() bool { return s.optional }

// Uint8Schema validates uint8 fields with optional minimum and maximum bounds.
type Uint8Schema struct {
	min      *uint8
	max      *uint8
	optional bool
}

// Uint8 constructs a uint8 validator.
func Uint8() *Uint8Schema { return &Uint8Schema{} }

// Min sets a minimum allowed value (inclusive).
func (s *Uint8Schema) Min(min uint8) *Uint8Schema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *Uint8Schema) Max(max uint8) *Uint8Schema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *Uint8Schema) Optional() *Uint8Schema {
	s.optional = true
	return s
}

// validate implements fieldValidator for Uint8Schema.
func (s *Uint8Schema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Uint8 {
		return validationError("validation failed: field %q expected uint8, got %s", fieldName, value.Kind())
	}

	val := value.Uint()
	if s.min != nil && val < uint64(*s.min) {
		return validationError("validation failed: field %q must be >= %d", fieldName, *s.min)
	}
	if s.max != nil && val > uint64(*s.max) {
		return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the uint8 field may be omitted.
func (s *Uint8Schema) isOptional() bool { return s.optional }

// Uint16Schema validates uint16 fields with optional minimum and maximum bounds.
type Uint16Schema struct {
	min      *uint16
	max      *uint16
	optional bool
}

// Uint16 constructs a uint16 validator.
func Uint16() *Uint16Schema { return &Uint16Schema{} }

// Min sets a minimum allowed value (inclusive).
func (s *Uint16Schema) Min(min uint16) *Uint16Schema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *Uint16Schema) Max(max uint16) *Uint16Schema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *Uint16Schema) Optional() *Uint16Schema {
	s.optional = true
	return s
}

// validate implements fieldValidator for Uint16Schema.
func (s *Uint16Schema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Uint16 {
		return validationError("validation failed: field %q expected uint16, got %s", fieldName, value.Kind())
	}

	val := value.Uint()
	if s.min != nil && val < uint64(*s.min) {
		return validationError("validation failed: field %q must be >= %d", fieldName, *s.min)
	}
	if s.max != nil && val > uint64(*s.max) {
		return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the uint16 field may be omitted.
func (s *Uint16Schema) isOptional() bool { return s.optional }

// Uint32Schema validates uint32 fields with optional minimum and maximum bounds.
type Uint32Schema struct {
	min      *uint32
	max      *uint32
	optional bool
}

// Uint32 constructs a uint32 validator.
func Uint32() *Uint32Schema { return &Uint32Schema{} }

// Min sets a minimum allowed value (inclusive).
func (s *Uint32Schema) Min(min uint32) *Uint32Schema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *Uint32Schema) Max(max uint32) *Uint32Schema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *Uint32Schema) Optional() *Uint32Schema {
	s.optional = true
	return s
}

// validate implements fieldValidator for Uint32Schema.
func (s *Uint32Schema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Uint32 {
		return validationError("validation failed: field %q expected uint32, got %s", fieldName, value.Kind())
	}

	val := value.Uint()
	if s.min != nil && val < uint64(*s.min) {
		return validationError("validation failed: field %q must be >= %d", fieldName, *s.min)
	}
	if s.max != nil && val > uint64(*s.max) {
		return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the uint32 field may be omitted.
func (s *Uint32Schema) isOptional() bool { return s.optional }

// Uint64Schema validates uint64 fields with optional minimum and maximum bounds.
type Uint64Schema struct {
	min      *uint64
	max      *uint64
	optional bool
}

// Uint64 constructs a uint64 validator.
func Uint64() *Uint64Schema { return &Uint64Schema{} }

// Min sets a minimum allowed value (inclusive).
func (s *Uint64Schema) Min(min uint64) *Uint64Schema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *Uint64Schema) Max(max uint64) *Uint64Schema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *Uint64Schema) Optional() *Uint64Schema {
	s.optional = true
	return s
}

// validate implements fieldValidator for Uint64Schema.
func (s *Uint64Schema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Uint64 {
		return validationError("validation failed: field %q expected uint64, got %s", fieldName, value.Kind())
	}

	val := value.Uint()
	if s.min != nil && val < *s.min {
		return validationError("validation failed: field %q must be >= %d", fieldName, *s.min)
	}
	if s.max != nil && val > *s.max {
		return validationError("validation failed: field %q must be <= %d", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the uint64 field may be omitted.
func (s *Uint64Schema) isOptional() bool { return s.optional }

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

// validate implements fieldValidator for FloatSchema.
func (s *FloatSchema) validate(fieldName string, value reflect.Value) error {
	if !isFloatKind(value.Kind()) {
		return validationError("validation failed: field %q expected float, got %s", fieldName, value.Kind())
	}

	floatVal := value.Float()
	if s.min != nil && floatVal < *s.min {
		return validationError("validation failed: field %q must be >= %g", fieldName, *s.min)
	}
	if s.max != nil && floatVal > *s.max {
		return validationError("validation failed: field %q must be <= %g", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the float field may be omitted.
func (s *FloatSchema) isOptional() bool { return s.optional }

// Float32Schema validates float32 fields with optional minimum and maximum bounds.
type Float32Schema struct {
	min      *float32
	max      *float32
	optional bool
}

// Float32 constructs a float32 validator.
func Float32() *Float32Schema { return &Float32Schema{} }

// Min sets a minimum allowed value (inclusive).
func (s *Float32Schema) Min(min float32) *Float32Schema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *Float32Schema) Max(max float32) *Float32Schema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *Float32Schema) Optional() *Float32Schema {
	s.optional = true
	return s
}

// validate implements fieldValidator for Float32Schema.
func (s *Float32Schema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Float32 {
		return validationError("validation failed: field %q expected float32, got %s", fieldName, value.Kind())
	}

	floatVal := value.Float()
	if s.min != nil && floatVal < float64(*s.min) {
		return validationError("validation failed: field %q must be >= %g", fieldName, *s.min)
	}
	if s.max != nil && floatVal > float64(*s.max) {
		return validationError("validation failed: field %q must be <= %g", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the float32 field may be omitted.
func (s *Float32Schema) isOptional() bool { return s.optional }

// Float64Schema validates float64 fields with optional minimum and maximum bounds.
type Float64Schema struct {
	min      *float64
	max      *float64
	optional bool
}

// Float64 constructs a float64 validator.
func Float64() *Float64Schema { return &Float64Schema{} }

// Min sets a minimum allowed value (inclusive).
func (s *Float64Schema) Min(min float64) *Float64Schema {
	s.min = &min
	return s
}

// Max sets a maximum allowed value (inclusive).
func (s *Float64Schema) Max(max float64) *Float64Schema {
	s.max = &max
	return s
}

// Optional allows the field to be missing from the struct without producing an error.
func (s *Float64Schema) Optional() *Float64Schema {
	s.optional = true
	return s
}

// validate implements fieldValidator for Float64Schema.
func (s *Float64Schema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Float64 {
		return validationError("validation failed: field %q expected float64, got %s", fieldName, value.Kind())
	}

	floatVal := value.Float()
	if s.min != nil && floatVal < *s.min {
		return validationError("validation failed: field %q must be >= %g", fieldName, *s.min)
	}
	if s.max != nil && floatVal > *s.max {
		return validationError("validation failed: field %q must be <= %g", fieldName, *s.max)
	}

	return nil
}

// isOptional reports whether the float64 field may be omitted.
func (s *Float64Schema) isOptional() bool { return s.optional }

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

// validate implements fieldValidator for BoolSchema.
func (b *BoolSchema) validate(fieldName string, value reflect.Value) error {
	if value.Kind() != reflect.Bool {
		return validationError("validation failed: field %q expected bool, got %s", fieldName, value.Kind())
	}
	return nil
}

// isOptional reports whether the boolean field may be omitted.
func (b *BoolSchema) isOptional() bool { return b.optional }

func isIntKind(kind reflect.Kind) bool {
	return isSignedIntKind(kind) || isUintKind(kind)
}

func isSignedIntKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	default:
		return false
	}
}

func isUintKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	default:
		return false
	}
}

func isFloatKind(kind reflect.Kind) bool {
	return kind == reflect.Float32 || kind == reflect.Float64
}
