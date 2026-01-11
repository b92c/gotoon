package gotoon

import (
	"strings"
	"testing"
)

func TestEncodeSimpleKeyValue(t *testing.T) {
	data := map[string]any{
		"name": "John",
		"age":  30,
	}

	toon, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "name: John") {
		t.Errorf("Expected 'name: John' in output, got: %s", toon)
	}
	if !strings.Contains(toon, "age: 30") {
		t.Errorf("Expected 'age: 30' in output, got: %s", toon)
	}
}

func TestEncodeNestedObjects(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"name":  "John",
			"email": "john@example.com",
		},
	}

	toon, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "user:") {
		t.Errorf("Expected 'user:' in output, got: %s", toon)
	}
	if !strings.Contains(toon, "name: John") {
		t.Errorf("Expected 'name: John' in output, got: %s", toon)
	}
	if !strings.Contains(toon, "email: john@example.com") {
		t.Errorf("Expected 'email: john@example.com' in output, got: %s", toon)
	}
}

func TestEncodeUniformArraysAsTables(t *testing.T) {
	data := []any{
		map[string]any{"id": 1, "name": "Alice"},
		map[string]any{"id": 2, "name": "Bob"},
	}

	toon, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "items[2]") {
		t.Errorf("Expected 'items[2]' in output, got: %s", toon)
	}
	if !strings.Contains(toon, "1,Alice") || !strings.Contains(toon, "Alice") {
		t.Errorf("Expected '1,Alice' or 'Alice' in output, got: %s", toon)
	}
	if !strings.Contains(toon, "2,Bob") || !strings.Contains(toon, "Bob") {
		t.Errorf("Expected '2,Bob' or 'Bob' in output, got: %s", toon)
	}
}

func TestEncodeBooleans(t *testing.T) {
	data := map[string]any{
		"active":  true,
		"deleted": false,
	}

	toon, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "active: true") {
		t.Errorf("Expected 'active: true' in output, got: %s", toon)
	}
	if !strings.Contains(toon, "deleted: false") {
		t.Errorf("Expected 'deleted: false' in output, got: %s", toon)
	}
}

func TestEncodeNull(t *testing.T) {
	data := []any{
		map[string]any{"id": 1, "name": nil},
		map[string]any{"id": 2, "name": "Bob"},
	}

	toon, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "items[2]") {
		t.Errorf("Expected items[2] in output, got: %s", toon)
	}
	if !strings.Contains(toon, "Bob") {
		t.Errorf("Expected 'Bob' in output, got: %s", toon)
	}
	lines := strings.Split(toon, "\n")
	foundEmptyCell := false
	for _, line := range lines {
		if strings.Contains(line, ",1") || strings.Contains(line, "1,") {
			foundEmptyCell = true
			break
		}
	}
	if !foundEmptyCell {
		t.Errorf("Expected empty cell for null value, got: %s", toon)
	}
}

func TestEscapeSpecialCharacters(t *testing.T) {
	data := map[string]any{
		"message": "Hello, World: Test",
	}

	toon, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "Hello\\, World\\: Test") {
		t.Errorf("Expected escaped output, got: %s", toon)
	}
}

func TestReducesTokenCount(t *testing.T) {
	data := []any{
		map[string]any{"id": 1, "name": "Alice", "email": "alice@example.com", "active": true},
		map[string]any{"id": 2, "name": "Bob", "email": "bob@example.com", "active": false},
		map[string]any{"id": 3, "name": "Charlie", "email": "charlie@example.com", "active": true},
	}

	diff := Diff(data)

	jsonChars := diff["json_chars"].(int)
	toonChars := diff["toon_chars"].(int)

	if toonChars >= jsonChars {
		t.Errorf("TOON should be smaller than JSON. JSON: %d, TOON: %d", jsonChars, toonChars)
	}
}

func TestOmitNullValues(t *testing.T) {
	config := &Config{
		MinRowsForTable: 2,
		MaxFlattenDepth: 3,
		EscapeStyle:     "backslash",
		Omit:            []string{"null"},
		OmitKeys:        []string{},
		KeyAliases:      make(map[string]string),
	}

	encoder := NewEncoder(config)

	data := map[string]any{
		"name":  "Alice",
		"email": nil,
		"phone": nil,
		"city":  "Amsterdam",
	}

	toon, err := encoder.Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "name: Alice") {
		t.Errorf("Expected 'name: Alice' in output, got: %s", toon)
	}
	if !strings.Contains(toon, "city: Amsterdam") {
		t.Errorf("Expected 'city: Amsterdam' in output, got: %s", toon)
	}
	if strings.Contains(toon, "email:") {
		t.Errorf("Should not contain 'email:', got: %s", toon)
	}
	if strings.Contains(toon, "phone:") {
		t.Errorf("Should not contain 'phone:', got: %s", toon)
	}
}

func TestOmitKeys(t *testing.T) {
	config := &Config{
		MinRowsForTable: 2,
		MaxFlattenDepth: 3,
		EscapeStyle:     "backslash",
		Omit:            []string{},
		OmitKeys:        []string{"created_at", "updated_at"},
		KeyAliases:      make(map[string]string),
	}

	encoder := NewEncoder(config)

	data := map[string]any{
		"name":       "Alice",
		"created_at": "2024-01-01",
		"updated_at": "2024-01-02",
		"city":       "Amsterdam",
	}

	toon, err := encoder.Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "name: Alice") {
		t.Errorf("Expected 'name: Alice' in output")
	}
	if !strings.Contains(toon, "city: Amsterdam") {
		t.Errorf("Expected 'city: Amsterdam' in output")
	}
	if strings.Contains(toon, "created_at:") {
		t.Errorf("Should not contain 'created_at:'")
	}
	if strings.Contains(toon, "updated_at:") {
		t.Errorf("Should not contain 'updated_at:'")
	}
}

func TestKeyAliases(t *testing.T) {
	config := &Config{
		MinRowsForTable: 2,
		MaxFlattenDepth: 3,
		EscapeStyle:     "backslash",
		Omit:            []string{},
		OmitKeys:        []string{},
		KeyAliases: map[string]string{
			"created_at": "c@",
			"updated_at": "u@",
		},
	}

	encoder := NewEncoder(config)

	data := map[string]any{
		"name":       "Alice",
		"created_at": "2024-01-01",
		"updated_at": "2024-01-02",
	}

	toon, err := encoder.Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "name: Alice") {
		t.Errorf("Expected 'name: Alice' in output")
	}
	if !strings.Contains(toon, "c@: 2024-01-01") {
		t.Errorf("Expected 'c@: 2024-01-01' in output, got: %s", toon)
	}
	if !strings.Contains(toon, "u@: 2024-01-02") {
		t.Errorf("Expected 'u@: 2024-01-02' in output, got: %s", toon)
	}
}

func TestNumberPrecision(t *testing.T) {
	config := &Config{
		MinRowsForTable: 2,
		MaxFlattenDepth: 3,
		EscapeStyle:     "backslash",
		NumberPrecision: 2,
	}

	encoder := NewEncoder(config)

	data := map[string]any{
		"pi":    3.14159265359,
		"price": 99.999,
	}

	toon, err := encoder.Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "3.14") {
		t.Errorf("Expected '3.14' in output, got: %s", toon)
	}
	if !strings.Contains(toon, "100.00") {
		t.Errorf("Expected '100.00' in output, got: %s", toon)
	}
}

func TestTruncateStrings(t *testing.T) {
	config := &Config{
		MinRowsForTable: 2,
		MaxFlattenDepth: 3,
		EscapeStyle:     "backslash",
		TruncateStrings: 20,
	}

	encoder := NewEncoder(config)

	data := map[string]any{
		"name": "Alice",
		"bio":  "This is a very long biography that should be truncated.",
	}

	toon, err := encoder.Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "name: Alice") {
		t.Errorf("Expected 'name: Alice' in output")
	}
	if !strings.Contains(toon, "...") {
		t.Errorf("Expected '...' in truncated bio, got: %s", toon)
	}
	if strings.Contains(toon, "truncated") {
		t.Errorf("Should not contain full 'truncated' word, got: %s", toon)
	}
}

func TestDiff(t *testing.T) {
	data := []any{
		map[string]any{"id": 1, "name": "Alice"},
		map[string]any{"id": 2, "name": "Bob"},
	}

	diff := Diff(data)

	if diff["json_chars"].(int) <= 0 {
		t.Errorf("json_chars should be positive")
	}
	if diff["toon_chars"].(int) <= 0 {
		t.Errorf("toon_chars should be positive")
	}
	if diff["savings_percent"].(float64) <= 0 {
		t.Errorf("savings_percent should be positive")
	}
}

func TestOnly(t *testing.T) {
	data := []any{
		map[string]any{"id": 1, "name": "Alice", "email": "alice@example.com"},
		map[string]any{"id": 2, "name": "Bob", "email": "bob@example.com"},
	}

	toon, err := Only(data, []string{"id", "name"})
	if err != nil {
		t.Fatalf("Only failed: %v", err)
	}

	if !strings.Contains(toon, "Alice") {
		t.Errorf("Expected 'Alice' in output")
	}
	if !strings.Contains(toon, "Bob") {
		t.Errorf("Expected 'Bob' in output")
	}
	if strings.Contains(toon, "email") {
		t.Errorf("Should not contain 'email', got: %s", toon)
	}
}
