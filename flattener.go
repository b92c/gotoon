package gotoon

import (
	"strings"
)

// ArrayFlattener handles flattening of nested objects in arrays.
type ArrayFlattener struct {
	maxDepth int
}

// NewArrayFlattener creates a new ArrayFlattener with the specified max depth.
func NewArrayFlattener(maxDepth int) *ArrayFlattener {
	return &ArrayFlattener{maxDepth: maxDepth}
}

// FlattenedData represents flattened array data with columns and rows.
type FlattenedData struct {
	Columns []string
	Rows    [][]any
}

// Flatten converts an array of objects with nested structures into a flat table format.
func (f *ArrayFlattener) Flatten(items []any) *FlattenedData {
	if len(items) == 0 {
		return &FlattenedData{Columns: []string{}, Rows: [][]any{}}
	}

	columns := f.extractColumns(items)
	rows := make([][]any, len(items))

	for i, item := range items {
		rows[i] = f.flattenRow(item, columns)
	}

	return &FlattenedData{
		Columns: columns,
		Rows:    rows,
	}
}

// HasNestedObjects checks if any item in the array has nested objects.
func (f *ArrayFlattener) HasNestedObjects(items []any) bool {
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}

		for _, value := range m {
			if nestedMap, ok := value.(map[string]any); ok && !isSequentialArray(nestedMap) {
				return true
			}
		}
	}
	return false
}

// extractColumns walks through all items and extracts column paths.
func (f *ArrayFlattener) extractColumns(items []any) []string {
	columnSet := make(map[string]bool)
	var columns []string

	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		f.walkItem(m, "", columnSet, &columns, 0)
	}

	return columns
}

// walkItem recursively walks an item to extract column paths.
func (f *ArrayFlattener) walkItem(item map[string]any, prefix string, columnSet map[string]bool, columns *[]string, depth int) {
	for key, value := range item {
		path := key
		if prefix != "" {
			path = prefix + "." + key
		}

		nestedMap, isMap := value.(map[string]any)
		if isMap && !isSequentialArray(nestedMap) && depth < f.maxDepth {
			f.walkItem(nestedMap, path, columnSet, columns, depth+1)
		} else {
			if !columnSet[path] {
				columnSet[path] = true
				*columns = append(*columns, path)
			}
		}
	}
}

// flattenRow converts a single item into a flat row based on column paths.
func (f *ArrayFlattener) flattenRow(item any, columns []string) []any {
	row := make([]any, len(columns))

	for i, col := range columns {
		row[i] = f.getByPath(item, col)
	}

	return row
}

// getByPath retrieves a value from nested maps using a dot-separated path.
func (f *ArrayFlattener) getByPath(data any, path string) any {
	segments := strings.Split(path, ".")

	current := data
	for _, segment := range segments {
		m, ok := current.(map[string]any)
		if !ok {
			return nil
		}

		value, exists := m[segment]
		if !exists {
			return nil
		}

		current = value
	}

	return current
}

// isSequentialArray checks if a value represents a sequential array (not a map).
func isSequentialArray(v any) bool {
	slice, ok := v.([]any)
	if !ok {
		return false
	}
	_ = slice
	return false
}
