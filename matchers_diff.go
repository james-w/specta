package specta

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// BuildMatcherStructDiff creates a structured diff output for generated matchers.
// It shows matched (✓), failed (✗), and unchecked (~) fields with proper indentation.
func BuildMatcherStructDiff(typeName string, fieldValues map[string]any, fieldResults map[string]*MatchResult) string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("%s {\n", typeName))

	// Sort field names for consistent output
	fieldNames := make([]string, 0, len(fieldValues))
	for name := range fieldValues {
		fieldNames = append(fieldNames, name)
	}
	sort.Strings(fieldNames)

	for _, fieldName := range fieldNames {
		fieldValue := fieldValues[fieldName]
		result := fieldResults[fieldName]

		if result == nil {
			// Unchecked field
			buf.WriteString(formatUncheckedField(fieldName, fieldValue, 1))
		} else if result.Matched {
			// Matched field
			buf.WriteString(formatMatchedField(fieldName, fieldValue, 1))
		} else {
			// Failed field
			buf.WriteString(formatFailedField(fieldName, fieldValue, result, 1))
		}
	}

	buf.WriteString("}")
	return buf.String()
}

// formatMatchedField formats a field that passed its matcher.
func formatMatchedField(name string, value any, indent int) string {
	symbol := colorize("✓", colorGreen)
	indentStr := strings.Repeat("  ", indent)
	valueStr := formatValue(value, indent)
	return fmt.Sprintf("%s%s %s: %s\n", indentStr, symbol, name, valueStr)
}

// formatFailedField formats a field that failed its matcher.
func formatFailedField(name string, value any, result *MatchResult, indent int) string {
	symbol := colorize("✗", colorRed)
	indentStr := strings.Repeat("  ", indent)

	// Determine if this is an equality check or constraint check
	if result.Expected != nil && result.Actual != nil {
		// Equality-style: show expected and actual
		expectedStr := formatValue(result.Expected, indent)
		actualStr := formatValue(result.Actual, indent)
		return fmt.Sprintf("%s%s %s: expected %s but got %s\n", indentStr, symbol, name, expectedStr, actualStr)
	} else {
		// Constraint-style: show the message
		return fmt.Sprintf("%s%s %s: %s\n", indentStr, symbol, name, result.Message)
	}
}

// formatUncheckedField formats a field that had no matcher.
func formatUncheckedField(name string, value any, indent int) string {
	symbol := colorize("~", colorGrey)
	indentStr := strings.Repeat("  ", indent)
	valueStr := formatValue(value, indent)
	return fmt.Sprintf("%s%s %s: %s\n", indentStr, symbol, name, valueStr)
}

// formatValue formats a value for display with smart truncation for collections.
func formatValue(value any, indent int) string {
	if value == nil {
		return "nil"
	}

	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		return formatSlice(v, indent)
	case reflect.Map:
		return formatMap(v, indent)
	case reflect.Struct:
		return formatStruct(v, indent)
	case reflect.String:
		return fmt.Sprintf("%q", value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// formatSlice formats a slice/array with smart truncation.
func formatSlice(v reflect.Value, indent int) string {
	length := v.Len()
	if length == 0 {
		return "[]"
	}

	const maxItems = 5
	if length <= maxItems {
		// Show all items
		items := make([]string, length)
		for i := 0; i < length; i++ {
			items[i] = fmt.Sprintf("%v", v.Index(i).Interface())
		}
		return fmt.Sprintf("[%s]", strings.Join(items, ", "))
	}

	// Show first few items and summarize rest
	items := make([]string, maxItems)
	for i := 0; i < maxItems; i++ {
		items[i] = fmt.Sprintf("%v", v.Index(i).Interface())
	}
	remaining := length - maxItems
	return fmt.Sprintf("[%s, ... and %d more]", strings.Join(items, ", "), remaining)
}

// formatMap formats a map with smart truncation.
func formatMap(v reflect.Value, indent int) string {
	length := v.Len()
	if length == 0 {
		return "map[]"
	}

	const maxItems = 5
	keys := v.MapKeys()
	if length <= maxItems {
		// Show all entries
		pairs := make([]string, length)
		for i, key := range keys {
			val := v.MapIndex(key)
			pairs[i] = fmt.Sprintf("%v:%v", key.Interface(), val.Interface())
		}
		return fmt.Sprintf("map[%s]", strings.Join(pairs, " "))
	}

	// Show first few entries and summarize rest
	pairs := make([]string, maxItems)
	for i := 0; i < maxItems; i++ {
		key := keys[i]
		val := v.MapIndex(key)
		pairs[i] = fmt.Sprintf("%v:%v", key.Interface(), val.Interface())
	}
	remaining := length - maxItems
	return fmt.Sprintf("map[%s ... and %d more]", strings.Join(pairs, " "), remaining)
}

// formatStruct formats a struct value.
func formatStruct(v reflect.Value, indent int) string {
	t := v.Type()
	return fmt.Sprintf("%s{...}", t.Name())
}

// buildReflectionStructDiff creates a structured diff using reflection for DeepEqual.
func buildReflectionStructDiff(expected, actual any) string {
	expectedVal := reflect.ValueOf(expected)
	actualVal := reflect.ValueOf(actual)

	if expectedVal.Kind() != reflect.Struct || actualVal.Kind() != reflect.Struct {
		// Not structs, fall back to simple format
		return fmt.Sprintf("expected %v but got %v", expected, actual)
	}

	typeName := expectedVal.Type().Name()
	if typeName == "" {
		typeName = "struct"
	}

	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("%s {\n", typeName))

	// Iterate through struct fields
	numFields := expectedVal.NumField()
	for i := 0; i < numFields; i++ {
		field := expectedVal.Type().Field(i)
		if !field.IsExported() {
			continue // Skip unexported fields
		}

		fieldName := field.Name
		expectedFieldVal := expectedVal.Field(i).Interface()
		actualFieldVal := actualVal.Field(i).Interface()

		if reflect.DeepEqual(expectedFieldVal, actualFieldVal) {
			// Matched
			buf.WriteString(formatMatchedField(fieldName, actualFieldVal, 1))
		} else {
			// Failed - create a result to pass to formatFailedField
			result := &MatchResult{
				Matched:  false,
				Expected: expectedFieldVal,
				Actual:   actualFieldVal,
			}
			buf.WriteString(formatFailedField(fieldName, actualFieldVal, result, 1))
		}
	}

	buf.WriteString("}")
	return buf.String()
}
