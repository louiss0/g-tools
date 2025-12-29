package regex_extract

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	ErrDuplicateGroupName   = errors.New("regex_extract: duplicate group name")
	ErrInvalidGroupName     = errors.New("regex_extract: group name contains a dash")
	ErrInvalidRegex         = errors.New("regex_extract: invalid regex pattern")
	ErrNoMatch              = errors.New("regex_extract: no match found")
	ErrStructType           = errors.New("regex_extract: target must be a struct type")
	ErrInvalidSliceValue    = errors.New("regex_extract: invalid slice value")
	ErrExpectedStruct       = errors.New("regex_extract: use ExtractToStruct for struct values")
	ErrUnexpectedGroupValue = errors.New("regex_extract: unexpected group value")

	floatPattern = regexp.MustCompile(`^\d+\.\d+$`)
	digitPattern = regexp.MustCompile(`^\d+$`)
)

type SliceValue interface {
	string | int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64
}

// MapValues transforms each value in the input map and returns a new map with the same keys.
func MapValues[key comparable, input any, output any](values map[key]input, mapper func(input) output) map[key]output {
	if values == nil {
		return nil
	}

	mapped := make(map[key]output, len(values))
	for entryKey, entryValue := range values {
		mapped[entryKey] = mapper(entryValue)
	}

	return mapped
}

func ExtractGroups(input string, pattern string) (map[string]string, error) {
	compiled, err := CompilePattern(pattern)
	if err != nil {
		return nil, err
	}

	return compiled.ExtractGroups(input)
}

func ExtractSlice[T SliceValue](input string, pattern string) ([]T, error) {
	compiled, err := CompilePattern(pattern)
	if err != nil {
		return nil, err
	}

	return ExtractSliceWithCompiled[T](input, compiled)
}

func ExtractTypedNamedGroups[T any](input string, pattern string) (map[string]T, error) {
	compiled, err := CompilePattern(pattern)
	if err != nil {
		return nil, err
	}

	return ExtractTypedNamedGroupsWithCompiled[T](input, compiled)
}

func ExtractToStruct[T any](input string, pattern string) (T, error) {
	compiled, err := CompilePattern(pattern)
	if err != nil {
		var zero T
		return zero, err
	}

	return ExtractToStructWithCompiled[T](input, compiled)
}

func ExtractTypedUnnamedGroups[T any](input string, pattern string) ([]T, error) {
	compiled, err := CompilePattern(pattern)
	if err != nil {
		return nil, err
	}

	return ExtractTypedUnnamedGroupsWithCompiled[T](input, compiled)
}

type CompiledPattern struct {
	pattern          string
	regex            *regexp.Regexp
	captureTreeOnce  sync.Once
	captureTreeNodes []*captureNode
}

func CompilePattern(pattern string) (*CompiledPattern, error) {
	regex, err := compileRegex(pattern)
	if err != nil {
		return nil, err
	}

	return &CompiledPattern{
		pattern: pattern,
		regex:   regex,
	}, nil
}

func (compiled *CompiledPattern) ExtractGroups(input string) (map[string]string, error) {
	submatches := compiled.regex.FindStringSubmatch(input)
	if len(submatches) == 0 {
		return nil, ErrNoMatch
	}

	groupNames := compiled.regex.SubexpNames()
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

func ExtractSliceWithCompiled[T SliceValue](input string, compiled *CompiledPattern) ([]T, error) {
	submatches := compiled.regex.FindStringSubmatch(input)
	if len(submatches) == 0 {
		return nil, ErrNoMatch
	}

	extracted := make([]T, 0, len(submatches)-1)
	for _, submatch := range submatches[1:] {
		value, err := parseSliceValue[T](submatch)
		if err != nil {
			return nil, err
		}
		extracted = append(extracted, value)
	}

	return extracted, nil
}

func ExtractTypedNamedGroupsWithCompiled[T any](input string, compiled *CompiledPattern) (map[string]T, error) {
	submatches := compiled.regex.FindStringSubmatch(input)
	if len(submatches) == 0 {
		return nil, ErrNoMatch
	}

	if isStructType[T]() {
		return nil, ErrExpectedStruct
	}

	captureTree := compiled.captureTree()
	extracted := map[string]any{}

	for _, node := range captureTree {
		if node.name == "" {
			if err := mergeNamedChildren(extracted, node, submatches); err != nil {
				return nil, err
			}
			continue
		}

		if strings.Contains(node.name, "-") {
			return nil, fmt.Errorf("%w: %s", ErrInvalidGroupName, node.name)
		}

		if _, exists := extracted[node.name]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateGroupName, node.name)
		}

		value, err := buildNamedValue(node, submatches)
		if err != nil {
			return nil, err
		}

		extracted[node.name] = value
	}

	return convertNamedGroupValues[T](extracted)
}

func ExtractToStructWithCompiled[T any](input string, compiled *CompiledPattern) (T, error) {
	var result T
	submatches := compiled.regex.FindStringSubmatch(input)
	if len(submatches) == 0 {
		return result, ErrNoMatch
	}

	targetType := reflect.TypeOf(result)
	if targetType.Kind() != reflect.Struct {
		return result, ErrStructType
	}

	structValue := reflect.New(targetType).Elem()

	captureTree := compiled.captureTree()
	if err := applyStructNodes(structValue, captureTree, submatches); err != nil {
		return result, err
	}

	return structValue.Interface().(T), nil
}

func ExtractTypedUnnamedGroupsWithCompiled[T any](input string, compiled *CompiledPattern) ([]T, error) {
	submatches := compiled.regex.FindStringSubmatch(input)
	if len(submatches) == 0 {
		return nil, ErrNoMatch
	}

	captureTree := compiled.captureTree()
	extracted := make([]T, 0, len(captureTree))

	for _, node := range captureTree {
		value, include, err := buildUnnamedValue(node, submatches)
		if err != nil {
			return nil, err
		}
		if include {
			converted, err := convertUnnamedGroupValue[T](value)
			if err != nil {
				return nil, err
			}
			extracted = append(extracted, converted)
		}
	}

	return extracted, nil
}

func (compiled *CompiledPattern) captureTree() []*captureNode {
	compiled.captureTreeOnce.Do(func() {
		compiled.captureTreeNodes = parseCaptureTree(compiled.pattern)
	})

	return compiled.captureTreeNodes
}

type captureNode struct {
	name     string
	index    int
	children []*captureNode
}

func parseCaptureTree(pattern string) []*captureNode {
	var rootNodes []*captureNode
	var groupStack []bool
	var captureStack []*captureNode
	escaped := false
	inCharClass := false
	groupIndex := 0

	for i := 0; i < len(pattern); i++ {
		character := pattern[i]

		if escaped {
			escaped = false
			continue
		}

		if character == '\\' {
			escaped = true
			continue
		}

		if inCharClass {
			if character == ']' {
				inCharClass = false
			}
			continue
		}

		if character == '[' {
			inCharClass = true
			continue
		}

		if character == '(' {
			groupName, isCapturing, advance := parseGroupStart(pattern, i)
			groupStack = append(groupStack, isCapturing)

			if isCapturing {
				groupIndex++
				node := &captureNode{
					name:  groupName,
					index: groupIndex,
				}

				if len(captureStack) == 0 {
					rootNodes = append(rootNodes, node)
				} else {
					parent := captureStack[len(captureStack)-1]
					parent.children = append(parent.children, node)
				}

				captureStack = append(captureStack, node)
			}

			if advance > 0 {
				i += advance
			}
			continue
		}

		if character == ')' {
			if len(groupStack) == 0 {
				continue
			}

			isCapturing := groupStack[len(groupStack)-1]
			groupStack = groupStack[:len(groupStack)-1]

			if isCapturing && len(captureStack) > 0 {
				captureStack = captureStack[:len(captureStack)-1]
			}
		}
	}

	return rootNodes
}

func parseGroupStart(pattern string, index int) (string, bool, int) {
	if index+1 >= len(pattern) || pattern[index+1] != '?' {
		return "", true, 0
	}

	if strings.HasPrefix(pattern[index+1:], "?P<") {
		nameStart := index + 4
		nameEnd := strings.IndexByte(pattern[nameStart:], '>')
		if nameEnd == -1 {
			return "", false, 0
		}
		nameEnd += nameStart

		return pattern[nameStart:nameEnd], true, nameEnd - index
	}

	return "", false, 0
}

func buildNamedValue(node *captureNode, submatches []string) (any, error) {
	if hasNamedChildren(node) {
		return buildNamedChildren(node, submatches)
	}

	if node.index >= len(submatches) {
		return "", nil
	}

	return inferValue(submatches[node.index]), nil
}

func buildNamedChildren(node *captureNode, submatches []string) (map[string]any, error) {
	extracted := map[string]any{}

	for _, child := range node.children {
		if child.name == "" {
			if err := mergeNamedChildren(extracted, child, submatches); err != nil {
				return nil, err
			}
			continue
		}

		if strings.Contains(child.name, "-") {
			return nil, fmt.Errorf("%w: %s", ErrInvalidGroupName, child.name)
		}

		if _, exists := extracted[child.name]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateGroupName, child.name)
		}

		value, err := buildNamedValue(child, submatches)
		if err != nil {
			return nil, err
		}

		extracted[child.name] = value
	}

	return extracted, nil
}

func mergeNamedChildren(target map[string]any, node *captureNode, submatches []string) error {
	if !hasNamedChildren(node) {
		return nil
	}

	children, err := buildNamedChildren(node, submatches)
	if err != nil {
		return err
	}

	for key, value := range children {
		if _, exists := target[key]; exists {
			return fmt.Errorf("%w: %s", ErrDuplicateGroupName, key)
		}
		target[key] = value
	}

	return nil
}

func hasNamedChildren(node *captureNode) bool {
	for _, child := range node.children {
		if child.name != "" || hasNamedChildren(child) {
			return true
		}
	}

	return false
}

func buildUnnamedValue(node *captureNode, submatches []string) (any, bool, error) {
	var values []any
	for _, child := range node.children {
		value, include, err := buildUnnamedValue(child, submatches)
		if err != nil {
			return nil, false, err
		}
		if include {
			values = append(values, value)
		}
	}

	if len(values) > 0 {
		return values, true, nil
	}

	if node.name != "" {
		return nil, false, nil
	}

	if node.index >= len(submatches) {
		return "", true, nil
	}

	return inferValue(submatches[node.index]), true, nil
}

func inferValue(value string) any {
	if isFloat(value) {
		parsed, err := strconv.ParseFloat(value, 64)
		if err == nil {
			if math.Abs(parsed) <= math.MaxFloat32 {
				return float32(parsed)
			}
			return parsed
		}
	}

	if isDigit(value) {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			return narrowUint(parsed)
		}
	}

	return value
}

func isFloat(value string) bool {
	return floatPattern.MatchString(value)
}

func isDigit(value string) bool {
	return digitPattern.MatchString(value)
}

func compileRegex(pattern string) (*regexp.Regexp, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidRegex, err)
	}
	return regex, nil
}

func parseSliceValue[T SliceValue](value string) (T, error) {
	var zero T

	switch any(zero).(type) {
	case string:
		return any(value).(T), nil
	case int:
		parsed, err := strconv.ParseInt(value, 10, strconv.IntSize)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(int(parsed)).(T), nil
	case int8:
		parsed, err := strconv.ParseInt(value, 10, 8)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(int8(parsed)).(T), nil
	case int16:
		parsed, err := strconv.ParseInt(value, 10, 16)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(int16(parsed)).(T), nil
	case int32:
		parsed, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(int32(parsed)).(T), nil
	case int64:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(parsed).(T), nil
	case uint:
		parsed, err := strconv.ParseUint(value, 10, strconv.IntSize)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(uint(parsed)).(T), nil
	case uint8:
		parsed, err := strconv.ParseUint(value, 10, 8)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(uint8(parsed)).(T), nil
	case uint16:
		parsed, err := strconv.ParseUint(value, 10, 16)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(uint16(parsed)).(T), nil
	case uint32:
		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(uint32(parsed)).(T), nil
	case uint64:
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(parsed).(T), nil
	case float32:
		parsed, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(float32(parsed)).(T), nil
	case float64:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
		}
		return any(parsed).(T), nil
	default:
		return zero, fmt.Errorf("%w: %s", ErrInvalidSliceValue, value)
	}
}

func isStructType[T any]() bool {
	targetType := reflect.TypeOf((*T)(nil)).Elem()
	return targetType.Kind() == reflect.Struct
}

func convertNamedGroupValues[T any](values map[string]any) (map[string]T, error) {
	converted := make(map[string]T, len(values))
	for key, value := range values {
		convertedValue, err := convertNamedGroupValue[T](value)
		if err != nil {
			return nil, err
		}
		converted[key] = convertedValue
	}

	return converted, nil
}

func convertNamedGroupValue[T any](value any) (T, error) {
	var zero T
	if value == nil {
		return zero, nil
	}

	targetType := reflect.TypeOf((*T)(nil)).Elem()
	valueValue := reflect.ValueOf(value)

	if targetType.Kind() == reflect.String && valueValue.Kind() != reflect.String {
		return zero, fmt.Errorf("%w: %v", ErrUnexpectedGroupValue, value)
	}

	if targetType.Kind() == reflect.Interface {
		if valueValue.IsValid() && valueValue.Type().Implements(targetType) {
			return valueValue.Interface().(T), nil
		}
		return zero, fmt.Errorf("%w: %v", ErrUnexpectedGroupValue, value)
	}

	if valueValue.Type().AssignableTo(targetType) {
		return valueValue.Interface().(T), nil
	}
	if valueValue.Type().ConvertibleTo(targetType) {
		return valueValue.Convert(targetType).Interface().(T), nil
	}

	return zero, fmt.Errorf("%w: %v", ErrUnexpectedGroupValue, value)
}

func convertUnnamedGroupValue[T any](value any) (T, error) {
	var zero T
	targetType := reflect.TypeOf((*T)(nil)).Elem()
	converted, err := convertValueToType(value, targetType)
	if err != nil {
		return zero, err
	}

	return converted.Interface().(T), nil
}

func convertValueToType(value any, targetType reflect.Type) (reflect.Value, error) {
	if value == nil {
		return reflect.Zero(targetType), nil
	}

	valueValue := reflect.ValueOf(value)

	if targetType.Kind() == reflect.Interface {
		if valueValue.Type().Implements(targetType) {
			return valueValue.Convert(targetType), nil
		}
		return reflect.Zero(targetType), fmt.Errorf("%w: %v", ErrUnexpectedGroupValue, value)
	}

	if targetType.Kind() == reflect.Pointer {
		converted, err := convertValueToType(value, targetType.Elem())
		if err != nil {
			return reflect.Zero(targetType), err
		}
		pointer := reflect.New(targetType.Elem())
		pointer.Elem().Set(converted)
		return pointer, nil
	}

	if targetType.Kind() == reflect.Struct {
		if valueValue.Kind() != reflect.Slice && valueValue.Kind() != reflect.Array {
			return reflect.Zero(targetType), fmt.Errorf("%w: %v", ErrUnexpectedGroupValue, value)
		}
		return convertSliceToStruct(valueValue, targetType)
	}

	if targetType.Kind() == reflect.Slice {
		if valueValue.Kind() != reflect.Slice && valueValue.Kind() != reflect.Array {
			return reflect.Zero(targetType), fmt.Errorf("%w: %v", ErrUnexpectedGroupValue, value)
		}
		return convertSliceToSlice(valueValue, targetType)
	}

	if valueValue.Type().AssignableTo(targetType) {
		return valueValue.Convert(targetType), nil
	}
	if valueValue.Type().ConvertibleTo(targetType) {
		return valueValue.Convert(targetType), nil
	}

	return reflect.Zero(targetType), fmt.Errorf("%w: %v", ErrUnexpectedGroupValue, value)
}

func convertSliceToStruct(values reflect.Value, targetType reflect.Type) (reflect.Value, error) {
	result := reflect.New(targetType).Elem()
	fieldCount := targetType.NumField()
	if values.Len() != fieldCount {
		return reflect.Zero(targetType), fmt.Errorf("%w: %v", ErrUnexpectedGroupValue, values.Interface())
	}

	for index := 0; index < fieldCount; index++ {
		field := targetType.Field(index)
		fieldValue := result.Field(index)
		if !fieldValue.CanSet() || !field.IsExported() {
			return reflect.Zero(targetType), fmt.Errorf("%w: %s", ErrExpectedStruct, field.Name)
		}

		converted, err := convertValueToType(values.Index(index).Interface(), field.Type)
		if err != nil {
			return reflect.Zero(targetType), err
		}
		fieldValue.Set(converted)
	}

	return result, nil
}

func convertSliceToSlice(values reflect.Value, targetType reflect.Type) (reflect.Value, error) {
	elementType := targetType.Elem()
	converted := reflect.MakeSlice(targetType, values.Len(), values.Len())
	for index := 0; index < values.Len(); index++ {
		value := values.Index(index).Interface()
		convertedValue, err := convertValueToType(value, elementType)
		if err != nil {
			return reflect.Zero(targetType), err
		}
		converted.Index(index).Set(convertedValue)
	}

	return converted, nil
}

func narrowUint(value uint64) any {
	if value <= math.MaxUint8 {
		return uint8(value)
	}
	if value <= math.MaxUint16 {
		return uint16(value)
	}
	if value <= math.MaxUint32 {
		return uint32(value)
	}
	return value
}

func applyStructNodes(structValue reflect.Value, nodes []*captureNode, submatches []string) error {
	for _, node := range nodes {
		if node.name == "" {
			if err := applyStructNodes(structValue, node.children, submatches); err != nil {
				return err
			}
			continue
		}

		if !isValidStructFieldName(node.name) {
			return fmt.Errorf("regex_extract: invalid struct field name %s", node.name)
		}

		field := structValue.FieldByName(node.name)
		if !field.IsValid() {
			return fmt.Errorf("regex_extract: missing struct field for group %s", node.name)
		}

		if err := assignStructValue(field, node, submatches); err != nil {
			return err
		}
	}

	return nil
}

func assignStructValue(field reflect.Value, node *captureNode, submatches []string) error {
	if !field.CanSet() {
		return fmt.Errorf("regex_extract: field %s must be exported", node.name)
	}

	if hasNamedChildren(node) {
		if field.Kind() == reflect.Pointer {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}

		if field.Kind() != reflect.Struct {
			return fmt.Errorf("regex_extract: field %s must be a struct for nested groups", node.name)
		}

		return applyStructNodes(field, node.children, submatches)
	}

	if node.index >= len(submatches) {
		return nil
	}

	return setStructValue(field, inferValue(submatches[node.index]))
}

func setStructValue(field reflect.Value, value any) error {
	valueValue := reflect.ValueOf(value)
	if !valueValue.IsValid() {
		return nil
	}

	if field.Kind() == reflect.Pointer {
		elementType := field.Type().Elem()
		if !valueValue.Type().ConvertibleTo(elementType) {
			return fmt.Errorf("regex_extract: cannot assign %s to %s", valueValue.Type(), elementType)
		}
		converted := valueValue.Convert(elementType)
		pointerValue := reflect.New(elementType)
		pointerValue.Elem().Set(converted)
		field.Set(pointerValue)
		return nil
	}

	if !valueValue.Type().ConvertibleTo(field.Type()) {
		return fmt.Errorf("regex_extract: cannot assign %s to %s", valueValue.Type(), field.Type())
	}
	field.Set(valueValue.Convert(field.Type()))
	return nil
}

func isValidStructFieldName(name string) bool {
	if name == "" {
		return false
	}

	first := name[0]
	if !isUppercaseLetter(first) {
		return false
	}

	for index := 1; index < len(name); index++ {
		if !isGoIdentifierCharacter(name[index]) {
			return false
		}
	}

	return true
}

func isUppercaseLetter(character byte) bool {
	return character >= 'A' && character <= 'Z'
}

func isGoIdentifierCharacter(character byte) bool {
	if character >= 'a' && character <= 'z' {
		return true
	}
	if character >= 'A' && character <= 'Z' {
		return true
	}
	if character >= '0' && character <= '9' {
		return true
	}
	return character == '_'
}
