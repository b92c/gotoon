package gotoon

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Encoder handles encoding Go data structures to TOON format.
type Encoder struct {
	config    *Config
	flattener *ArrayFlattener
}

// NewEncoder creates a new Encoder with the given configuration.
func NewEncoder(config *Config) *Encoder {
	if config == nil {
		config = DefaultConfig()
	}
	return &Encoder{
		config:    config,
		flattener: NewArrayFlattener(config.MaxFlattenDepth),
	}
}

// Encode converts data to TOON format string.
func (e *Encoder) Encode(data any) (string, error) {
	if str, ok := data.(string); ok && looksLikeJSON(str) {
		var decoded any
		if err := json.Unmarshal([]byte(str), &decoded); err == nil {
			return e.valueToToon(decoded, 0), nil
		}
	}

	return e.valueToToon(data, 0), nil
}

// valueToToon converts a value to TOON format with indentation.
func (e *Encoder) valueToToon(value any, depth int) string {
	indent := strings.Repeat("  ", depth)

	if slice, ok := value.([]any); ok {
		if isSequentialArraySlice(slice) {
			if e.flattener.HasNestedObjects(slice) {
				flattened := e.flattener.Flatten(slice)
				return e.flattenedToToon(flattened, depth)
			}

			if e.isArrayOfUniformObjects(slice) {
				return e.arrayOfObjectsToToon(slice, depth)
			}

			return e.sequentialArrayToToon(slice, depth)
		}
	}

	if m, ok := value.(map[string]any); ok {
		return e.associativeArrayToToon(m, depth)
	}

	return indent + e.escapeScalar(value)
}

// flattenedToToon converts flattened data to TOON table format.
func (e *Encoder) flattenedToToon(flattened *FlattenedData, depth int) string {
	indent := strings.Repeat("  ", depth)

	formattedCols := make([]string, len(flattened.Columns))
	for i, col := range flattened.Columns {
		formattedCols[i] = e.config.formatKey(col)
	}

	header := fmt.Sprintf("%sitems[%d]{%s}:", indent, len(flattened.Rows), strings.Join(formattedCols, ","))

	rowLines := make([]string, len(flattened.Rows))
	for i, row := range flattened.Rows {
		cells := make([]string, len(row))
		for j, cell := range row {
			cells[j] = e.escapeScalar(cell)
		}
		rowLines[i] = indent + "  " + strings.Join(cells, ",")
	}

	return header + "\n" + strings.Join(rowLines, "\n")
}

// arrayOfObjectsToToon converts an array of uniform objects to TOON table format.
func (e *Encoder) arrayOfObjectsToToon(arr []any, depth int) string {
	if len(arr) == 0 {
		return strings.Repeat("  ", depth) + "items[0]{}:"
	}

	firstObj, ok := arr[0].(map[string]any)
	if !ok {
		return e.sequentialArrayToToon(arr, depth)
	}

	fields := make([]string, 0, len(firstObj))
	for key := range firstObj {
		fields = append(fields, key)
	}

	formattedFields := make([]string, len(fields))
	for i, f := range fields {
		formattedFields[i] = e.config.formatKey(f)
	}

	indent := strings.Repeat("  ", depth)
	header := fmt.Sprintf("%sitems[%d]{%s}:", indent, len(arr), strings.Join(formattedFields, ","))

	rows := make([]string, len(arr))
	for i, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}

		cells := make([]string, len(fields))
		for j, field := range fields {
			cells[j] = e.escapeScalar(obj[field])
		}
		rows[i] = indent + "  " + strings.Join(cells, ",")
	}

	return header + "\n" + strings.Join(rows, "\n")
}

// sequentialArrayToToon converts a sequential array to TOON format.
func (e *Encoder) sequentialArrayToToon(arr []any, depth int) string {
	indent := strings.Repeat("  ", depth)
	lines := make([]string, len(arr))

	for i, item := range arr {
		if isScalar(item) {
			lines[i] = indent + e.escapeScalar(item)
		} else {
			lines[i] = e.valueToToon(item, depth)
		}
	}

	return strings.Join(lines, "\n")
}

// associativeArrayToToon converts a map to TOON format.
func (e *Encoder) associativeArrayToToon(m map[string]any, depth int) string {
	indent := strings.Repeat("  ", depth)
	lines := []string{}

	for key, val := range m {
		if e.config.shouldOmitKey(key) {
			continue
		}

		if val == nil && e.config.shouldOmit("null") {
			continue
		}

		if str, ok := val.(string); ok && str == "" && e.config.shouldOmit("empty") {
			continue
		}

		if b, ok := val.(bool); ok && !b && e.config.shouldOmit("false") {
			continue
		}

		formattedKey := e.config.formatKey(key)

		if isScalar(val) {
			lines = append(lines, indent+formattedKey+": "+e.escapeScalar(val))
		} else {
			lines = append(lines, indent+formattedKey+":")
			lines = append(lines, e.valueToToon(val, depth+1))
		}
	}

	return strings.Join(lines, "\n")
}

// escapeScalar converts a scalar value to its string representation.
func (e *Encoder) escapeScalar(v any) string {
	if v == nil {
		return ""
	}

	switch val := v.(type) {
	case bool:
		if val {
			return "true"
		}
		return "false"

	case time.Time:
		if e.config.DateFormat != "" {
			return val.Format(e.config.DateFormat)
		}
		return val.Format(time.RFC3339)

	case float64:
		if e.config.NumberPrecision >= 0 {
			return strconv.FormatFloat(val, 'f', e.config.NumberPrecision, 64)
		}
		s := strconv.FormatFloat(val, 'f', -1, 64)
		return s

	case float32:
		if e.config.NumberPrecision >= 0 {
			return strconv.FormatFloat(float64(val), 'f', e.config.NumberPrecision, 32)
		}
		s := strconv.FormatFloat(float64(val), 'f', -1, 32)
		return s

	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", val)

	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", val)

	case string:
		s := val

		if e.config.DateFormat != "" && looksLikeISODate(s) {
			if t, err := time.Parse(time.RFC3339, s); err == nil {
				return t.Format(e.config.DateFormat)
			}
			if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
				return t.Format(e.config.DateFormat)
			}
			if t, err := time.Parse("2006-01-02", s); err == nil {
				return t.Format(e.config.DateFormat)
			}
		}

		s = strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(s, " "))

		if e.config.EscapeStyle == "backslash" {
			s = strings.ReplaceAll(s, "\\", "\\\\")
			s = strings.ReplaceAll(s, ",", "\\,")
			s = strings.ReplaceAll(s, ":", "\\:")
			s = strings.ReplaceAll(s, "\n", "\\n")
		}

		if e.config.TruncateStrings > 0 && len(s) > e.config.TruncateStrings {
			s = s[:e.config.TruncateStrings] + "..."
		}

		return s

	case []any, map[string]any:
		if bytes, err := json.Marshal(val); err == nil {
			return string(bytes)
		}
		return "[]"

	default:
		return fmt.Sprintf("%v", val)
	}
}

// isArrayOfUniformObjects checks if all items are maps with the same keys.
func (e *Encoder) isArrayOfUniformObjects(arr []any) bool {
	if len(arr) < e.config.MinRowsForTable {
		return false
	}

	var firstKeys []string
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			return false
		}

		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}

		if firstKeys == nil {
			firstKeys = keys
		} else {
			if !stringSlicesEqual(keys, firstKeys) {
				return false
			}
		}
	}

	return true
}

// Helper functions

func isScalar(v any) bool {
	if v == nil {
		return true
	}
	switch v.(type) {
	case bool, string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, time.Time:
		return true
	default:
		return false
	}
}

func isSequentialArraySlice(v []any) bool {
	// In Go, []any is always sequential
	return true
}

func looksLikeJSON(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && (strings.HasPrefix(s, "{") || strings.HasPrefix(s, "["))
}

func looksLikeISODate(s string) bool {
	matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}([T\s]\d{2}:\d{2}(:\d{2})?)?`, s)
	return matched
}

func safeKey(k string) string {
	re := regexp.MustCompile(`[^A-Za-z0-9_\-\.]`)
	return re.ReplaceAllString(k, "")
}

func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[string]int)
	for _, s := range a {
		counts[s]++
	}
	for _, s := range b {
		counts[s]--
		if counts[s] < 0 {
			return false
		}
	}
	return true
}
