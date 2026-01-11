package gotoon

import (
	"strings"
	"testing"
)

func TestEncodeNestedObjectsWithDotNotation(t *testing.T) {
	data := map[string]any{
		"bookings": []any{
			map[string]any{
				"id":     "abc",
				"status": "confirmed",
				"artist": map[string]any{"id": "xyz", "name": "DJ Test"},
			},
			map[string]any{
				"id":     "def",
				"status": "pending",
				"artist": map[string]any{"id": "uvw", "name": "Band"},
			},
		},
	}

	toon, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "artist.id") || !strings.Contains(toon, "artist.name") {
		t.Errorf("Expected nested columns with dot notation, got: %s", toon)
	}
	if !strings.Contains(toon, "abc") && !strings.Contains(toon, "confirmed") {
		t.Errorf("Expected data values in output, got: %s", toon)
	}
}

func TestDecodeNestedObjectsFromDotNotation(t *testing.T) {
	toon := `items[2]{id,artist.id,artist.name}:
  abc,xyz,DJ Test
  def,uvw,Band`

	decoded, err := Decode(toon)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	items, exists := decoded["items"]
	if !exists {
		items, exists = decoded["_items"]
	}

	if !exists {
		t.Fatalf("Expected 'items' or '_items' in decoded result, got: %+v", decoded)
	}

	itemsSlice, ok := items.([]any)
	if !ok {
		t.Fatalf("Expected items to be a slice, got: %T", items)
	}

	if len(itemsSlice) != 2 {
		t.Fatalf("Expected 2 items, got: %d", len(itemsSlice))
	}

	first, ok := itemsSlice[0].(map[string]any)
	if !ok {
		t.Fatalf("Expected first item to be a map, got: %T", itemsSlice[0])
	}

	if first["id"] != "abc" {
		t.Errorf("Expected id='abc', got: %v", first["id"])
	}

	artist, ok := first["artist"].(map[string]any)
	if !ok {
		t.Fatalf("Expected artist to be a map, got: %T", first["artist"])
	}

	if artist["id"] != "xyz" {
		t.Errorf("Expected artist.id='xyz', got: %v", artist["id"])
	}
	if artist["name"] != "DJ Test" {
		t.Errorf("Expected artist.name='DJ Test', got: %v", artist["name"])
	}
}

func TestRoundTripNestedObjects(t *testing.T) {
	data := []any{
		map[string]any{
			"id":   1,
			"user": map[string]any{"name": "John", "email": "john@test.com"},
		},
		map[string]any{
			"id":   2,
			"user": map[string]any{"name": "Jane", "email": "jane@test.com"},
		},
	}

	encoded, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	var items []any
	if itemsVal, exists := decoded["items"]; exists {
		items = itemsVal.([]any)
	} else if itemsVal, exists := decoded["_items"]; exists {
		items = itemsVal.([]any)
	} else {
		t.Fatalf("Could not find items in decoded result: %+v", decoded)
	}

	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got: %d", len(items))
	}

	first := items[0].(map[string]any)
	if first["id"] != 1 {
		t.Errorf("Expected id=1, got: %v", first["id"])
	}

	user := first["user"].(map[string]any)
	if user["name"] != "John" {
		t.Errorf("Expected name='John', got: %v", user["name"])
	}
	if user["email"] != "john@test.com" {
		t.Errorf("Expected email='john@test.com', got: %v", user["email"])
	}
}

func TestMultiLevelNesting(t *testing.T) {
	data := []any{
		map[string]any{
			"id": 1,
			"event": map[string]any{
				"name":  "Festival",
				"venue": map[string]any{"name": "Club X", "city": "Amsterdam"},
			},
		},
		map[string]any{
			"id": 2,
			"event": map[string]any{
				"name":  "Concert",
				"venue": map[string]any{"name": "Arena Y", "city": "Rotterdam"},
			},
		},
	}

	toon, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "event.name") {
		t.Errorf("Expected 'event.name' column")
	}
	if !strings.Contains(toon, "event.venue.name") {
		t.Errorf("Expected 'event.venue.name' column")
	}
	if !strings.Contains(toon, "event.venue.city") {
		t.Errorf("Expected 'event.venue.city' column")
	}

	decoded, err := Decode(toon)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	var items []any
	if itemsVal, exists := decoded["items"]; exists {
		items = itemsVal.([]any)
	} else if itemsVal, exists := decoded["_items"]; exists {
		items = itemsVal.([]any)
	}

	if len(items) != 2 {
		t.Fatalf("Expected 2 items after decode, got: %d", len(items))
	}

	first := items[0].(map[string]any)
	event := first["event"].(map[string]any)
	venue := event["venue"].(map[string]any)

	if venue["city"] != "Amsterdam" {
		t.Errorf("Expected city='Amsterdam', got: %v", venue["city"])
	}
}

func TestMissingNestedProperties(t *testing.T) {
	data := []any{
		map[string]any{
			"id":     1,
			"artist": map[string]any{"name": "DJ A"},
		},
		map[string]any{
			"id":     2,
			"artist": map[string]any{"name": "DJ B", "genre": "Techno"},
		},
	}

	toon, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "artist.name") {
		t.Errorf("Expected 'artist.name' column")
	}
	if !strings.Contains(toon, "artist.genre") {
		t.Errorf("Expected 'artist.genre' column")
	}

	decoded, err := Decode(toon)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	var items []any
	if itemsVal, exists := decoded["items"]; exists {
		items = itemsVal.([]any)
	} else if itemsVal, exists := decoded["_items"]; exists {
		items = itemsVal.([]any)
	}

	first := items[0].(map[string]any)
	artist1 := first["artist"].(map[string]any)

	if artist1["genre"] != nil {
		t.Errorf("Expected genre=nil for first artist, got: %v", artist1["genre"])
	}

	second := items[1].(map[string]any)
	artist2 := second["artist"].(map[string]any)

	if artist2["genre"] != "Techno" {
		t.Errorf("Expected genre='Techno' for second artist, got: %v", artist2["genre"])
	}
}

func TestComplexBookingExample(t *testing.T) {
	data := map[string]any{
		"count":       2,
		"total_count": 2,
		"bookings": []any{
			map[string]any{
				"id":     "abc123",
				"status": "confirmed",
				"artist": map[string]any{"id": "art1", "name": "DJ Awesome"},
				"event":  map[string]any{"id": "evt1", "name": "Summer Festival"},
				"financial": map[string]any{
					"currency":   "EUR",
					"artist_fee": 2500,
				},
			},
			map[string]any{
				"id":     "def456",
				"status": "pending",
				"artist": map[string]any{"id": "art2", "name": "Band Cool"},
				"event":  map[string]any{"id": "evt1", "name": "Summer Festival"},
				"financial": map[string]any{
					"currency":   "EUR",
					"artist_fee": 1500,
				},
			},
		},
	}

	toon, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "count: 2") {
		t.Errorf("Expected 'count: 2' in output")
	}
	if !strings.Contains(toon, "artist.id") {
		t.Errorf("Expected 'artist.id' column")
	}
	if !strings.Contains(toon, "artist.name") {
		t.Errorf("Expected 'artist.name' column")
	}
	if !strings.Contains(toon, "event.id") {
		t.Errorf("Expected 'event.id' column")
	}
	if !strings.Contains(toon, "financial.currency") {
		t.Errorf("Expected 'financial.currency' column")
	}

	decoded, err := Decode(toon)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	count, ok := decoded["count"]
	if !ok {
		t.Fatalf("Expected 'count' in decoded result")
	}
	if count != 2 {
		t.Errorf("Expected count=2, got: %v", count)
	}

	bookingsVal, ok := decoded["bookings"]
	if !ok {
		t.Fatalf("Expected 'bookings' in decoded result, got: %+v", decoded)
	}

	bookings, ok := bookingsVal.([]any)
	if !ok {
		t.Fatalf("Expected bookings to be []any, got: %T = %+v", bookingsVal, bookingsVal)
	}

	if len(bookings) != 2 {
		t.Fatalf("Expected 2 bookings, got: %d", len(bookings))
	}

	first := bookings[0].(map[string]any)
	artist := first["artist"].(map[string]any)

	if artist["name"] != "DJ Awesome" {
		t.Errorf("Expected artist.name='DJ Awesome', got: %v", artist["name"])
	}

	financial := first["financial"].(map[string]any)
	if financial["artist_fee"] != 2500 {
		t.Errorf("Expected artist_fee=2500, got: %v", financial["artist_fee"])
	}
}
