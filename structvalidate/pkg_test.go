package structvalidate

import (
	"regexp"
	"testing"
)

type sample struct {
	Name     string
	Age      int
	Email    string
	Version  string
	UserID   string
	Nickname string
}

func TestObjectSchema_StringRegexFailure(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Email": String().Regex(regexp.MustCompile(`^[^@]+@[^@]+\.[^@]+$`)),
	})

	input := sample{Name: "Jane", Age: 30, Email: "invalid-email"}
	if err := schema.Validate(input); err == nil {
		t.Fatalf("expected regex mismatch to produce an error")
	}
}

func TestObjectSchema_ParseAndMustParse(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Age": Int().Min(21),
	})

	type user struct {
		Age int
	}

	if err := schema.Parse(user{Age: 25}); err != nil {
		t.Fatalf("expected parse to succeed, got %v", err)
	}

	parseErr := schema.Parse(user{Age: 18})
	if parseErr == nil {
		t.Fatalf("expected parse to return a ParseError")
	}

	if _, ok := any(parseErr).(*ParseError); !ok {
		t.Fatalf("expected error to be a ParseError, got %T", parseErr)
	}

	deferred := func() (recovered any) {
		defer func() { recovered = recover() }()
		schema.MustParse(user{Age: 18})
		return nil
	}()

	if deferred == nil {
		t.Fatalf("expected MustParse to panic on validation error")
	}
}

func TestObjectSchema_PrimitiveValidators(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Name":   String().NonEmpty(),
		"Age":    Int().Min(21),
		"Score":  Float().Max(100.0),
		"Active": Bool(),
	})

	type extended struct {
		sample
		Score  float64
		Active bool
	}

	input := extended{
		sample: sample{Name: "Jane", Age: 25, Email: "jane@example.com"},
		Score:  98.5,
		Active: true,
	}

	if err := schema.Validate(input); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}

func TestObjectSchema_MissingField(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Phone": String().NonEmpty(),
	})

	input := sample{Name: "Jane", Age: 25, Email: "jane@example.com"}
	if err := schema.Validate(input); err == nil {
		t.Fatalf("expected missing field to produce an error")
	}
}

func TestObjectSchema_OptionalFieldMissing(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Phone": String().NonEmpty().Optional(),
	})

	input := sample{Name: "Jane", Age: 25, Email: "jane@example.com"}
	if err := schema.Validate(input); err != nil {
		t.Fatalf("optional field should be allowed to be missing: %v", err)
	}
}

func TestObjectSchema_OptionalFieldValidatedWhenPresent(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Phone": String().Regex(regexp.MustCompile(`^\\d{10}$`)).Optional(),
	})

	input := struct {
		sample
		Phone string
	}{sample: sample{Name: "Jane", Age: 25, Email: "jane@example.com"}, Phone: "invalid"}

	if err := schema.Validate(input); err == nil {
		t.Fatalf("expected optional field to be validated when present")
	}
}

func TestStringLengthOmitSpacesByDefault(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Name": String().MinLength(4),
	})

	input := sample{Name: "ab cd"}
	if err := schema.Validate(input); err != nil {
		t.Fatalf("expected validation to ignore spaces in length: %v", err)
	}
}

func TestStringLengthCountsSpacesWhenConfigured(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Name": String().MaxLength(3).CountSpaces(),
	})

	input := sample{Name: "a b c"}
	if err := schema.Validate(input); err == nil {
		t.Fatalf("expected length check to count spaces and fail")
	}
}

func TestStringEmailValidation(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Email": String().Email(),
	})

	input := sample{Email: "not-an-email"}
	if err := schema.Validate(input); err == nil {
		t.Fatalf("expected email validation to fail")
	}
}

func TestStringUUIDValidation(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"UserID": String().UUID(),
	})

	input := sample{UserID: "1234"}
	if err := schema.Validate(input); err == nil {
		t.Fatalf("expected UUID validation to fail")
	}
}

func TestStringSemverValidation(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Version": String().Semver(),
	})

	input := sample{Version: "v1.2.3-beta+build"}
	if err := schema.Validate(input); err != nil {
		t.Fatalf("expected semantic version to pass validation: %v", err)
	}
}

func TestStringLengthOmitSpacesFailure(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Nickname": String().MinLength(3),
	})

	input := sample{Nickname: "a b"}
	if err := schema.Validate(input); err == nil {
		t.Fatalf("expected validation to fail when non-space length is too short")
	}
}

func TestObjectSchema_UnsignedIntegersDoNotPanic(t *testing.T) {
	schema := Object(map[string]fieldValidator{
		"Count": Int().Max(5),
	})

	type unsignedSample struct {
		Count uint
	}

	recovered, err := func() (any, error) {
		var rec any
		defer func() { rec = recover() }()
		err := schema.Validate(unsignedSample{Count: 10})
		return rec, err
	}()

	if recovered != nil {
		t.Fatalf("expected validation to return an error instead of panicking: %v", recovered)
	}

	if err == nil {
		t.Fatalf("expected validation to fail for value exceeding max")
	}
}
