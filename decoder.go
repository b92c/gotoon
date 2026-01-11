package gotoon

import (
	"regexp"
	"strconv"
	"strings"
)

// Decoder handles decoding TOON format strings to Go data structures.
type Decoder struct {
	config      *Config
	unflattener *ArrayUnflattener
}

// NewDecoder creates a new Decoder with the given configuration.
func NewDecoder(config *Config) *Decoder {
	if config == nil {
		config = DefaultConfig()
	}
	return &Decoder{
		config:      config,
		unflattener: NewArrayUnflattener(),
	}
}

// Decode converts a TOON format string to Go data structures.
func (d *Decoder) Decode(toon string) (map[string]any, error) {
	lines := strings.Split(toon, "\n")
	result := make(map[string]any)

	type stackItem struct {
		data   map[string]any
		indent int
		key    string
	}

	stack := []stackItem{{data: result, indent: -1, key: ""}}

	i := 0
	for i < len(lines) {
		line := lines[i]

		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		content := strings.TrimSpace(line)

		for len(stack) > 1 && indent <= stack[len(stack)-1].indent {
			stack = stack[:len(stack)-1]
		}

		current := stack[len(stack)-1].data

		if match := regexp.MustCompile(`^items\[(\d+)\]\{([^\}]*)\}:$`).FindStringSubmatch(content); match != nil {
			rowCount, _ := strconv.Atoi(match[1])
			columnsStr := match[2]

			var columns []string
			if columnsStr != "" {
				columns = strings.Split(columnsStr, ",")
				for j := range columns {
					columns[j] = strings.TrimSpace(columns[j])
				}
			}

			rows := [][]any{}
			for j := 0; j < rowCount && (i+1+j) < len(lines); j++ {
				rowLine := lines[i+1+j]
				rowContent := strings.TrimSpace(rowLine)

				if rowContent == "" {
					continue
				}

				cells := d.parseRow(rowContent, len(columns))
				rows = append(rows, cells)
			}

			i += rowCount

			var items []any
			if hasNestedColumns(columns) {
				objects := d.unflattener.Unflatten(rows, columns)
				items = make([]any, len(objects))
				for idx, obj := range objects {
					items[idx] = obj
				}
			} else {
				items = d.rowsToObjects(rows, columns)
			}

			if len(stack) > 1 {
				parentKey := stack[len(stack)-1].key
				parentData := stack[len(stack)-2].data
				parentData[parentKey] = items
			} else {
				current["_items"] = items
			}
		} else if strings.HasSuffix(content, ":") && !strings.Contains(content, ": ") {
			key := strings.TrimSuffix(content, ":")
			current[key] = make(map[string]any)
			stack = append(stack, stackItem{
				data:   current[key].(map[string]any),
				indent: indent,
				key:    key,
			})
		} else if strings.Contains(content, ": ") {
			parts := strings.SplitN(content, ": ", 2)
			key := parts[0]
			value := d.parseValue(parts[1])
			current[key] = value
		}

		i++
	}

	if items, exists := result["_items"]; exists {
		return map[string]any{"items": items}, nil
	}

	return result, nil
}

// parseRow parses a CSV-like row with escape handling.
func (d *Decoder) parseRow(row string, expectedCount int) []any {
	cells := []any{}
	current := ""
	escaped := false

	for i := 0; i < len(row); i++ {
		char := row[i]

		if escaped {
			current += string(char)
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == ',' {
			cells = append(cells, d.parseValue(current))
			current = ""
			continue
		}

		current += string(char)
	}

	cells = append(cells, d.parseValue(current))

	for len(cells) < expectedCount {
		cells = append(cells, nil)
	}

	return cells
}

// parseValue converts a string value to its appropriate type.
func (d *Decoder) parseValue(value string) any {
	value = strings.TrimSpace(value)

	if value == "" || value == "null" {
		return nil
	}

	if value == "true" {
		return true
	}

	if value == "false" {
		return false
	}

	if strings.Contains(value, ".") {
		if f, err := strconv.ParseFloat(value, 64); err == nil {
			return f
		}
	} else {
		if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int(i)
		}
	}

	value = strings.ReplaceAll(value, "\\n", "\n")
	value = strings.ReplaceAll(value, "\\,", ",")
	value = strings.ReplaceAll(value, "\\:", ":")
	value = strings.ReplaceAll(value, "\\\\", "\\")

	return value
}

// rowsToObjects converts rows to objects (non-nested case).
func (d *Decoder) rowsToObjects(rows [][]any, columns []string) []any {
	objects := make([]any, len(rows))

	for i, row := range rows {
		obj := make(map[string]any)
		for j, col := range columns {
			if j < len(row) {
				obj[col] = row[j]
			} else {
				obj[col] = nil
			}
		}
		objects[i] = obj
	}

	return objects
}

// hasNestedColumns checks if any column contains a dot (nested path).
func hasNestedColumns(columns []string) bool {
	for _, col := range columns {
		if strings.Contains(col, ".") {
			return true
		}
	}
	return false
}
