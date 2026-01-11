package gotoon

import "strings"

// ArrayUnflattener handles unflattening of flat table data back into nested objects.
type ArrayUnflattener struct{}

// NewArrayUnflattener creates a new ArrayUnflattener.
func NewArrayUnflattener() *ArrayUnflattener {
	return &ArrayUnflattener{}
}

// Unflatten converts flat rows back into nested objects.
func (u *ArrayUnflattener) Unflatten(rows [][]any, columns []string) []map[string]any {
	result := make([]map[string]any, len(rows))

	for i, row := range rows {
		result[i] = u.unflattenRow(row, columns)
	}

	return result
}

// unflattenRow converts a single flat row into a nested object.
func (u *ArrayUnflattener) unflattenRow(row []any, columns []string) map[string]any {
	item := make(map[string]any)

	for i, column := range columns {
		var value any
		if i < len(row) {
			value = row[i]
		}
		u.setByPath(item, column, value)
	}

	return item
}

// setByPath sets a value in nested maps using a dot-separated path.
func (u *ArrayUnflattener) setByPath(data map[string]any, path string, value any) {
	segments := strings.Split(path, ".")

	current := data
	for i, segment := range segments {
		if i == len(segments)-1 {
			current[segment] = value
		} else {
			if _, exists := current[segment]; !exists {
				current[segment] = make(map[string]any)
			}

			nextMap, ok := current[segment].(map[string]any)
			if !ok {
				nextMap = make(map[string]any)
				current[segment] = nextMap
			}

			current = nextMap
		}
	}
}
