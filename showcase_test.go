package gotoon

import (
	"strings"
	"testing"
)

func TestFullFeatureShowcase(t *testing.T) {
	data := map[string]any{
		"api_version": "v2",
		"total_count": 3,
		"timestamp":   "2024-01-15T14:30:00Z",
		"results": []any{
			map[string]any{
				"id":     "booking_1",
				"status": "confirmed",
				"amount": 2500.50,
				"active": true,
				"notes":  nil,
				"client": map[string]any{
					"id":    "client_a",
					"name":  "Acme Corp",
					"email": "contact@acme.com",
					"address": map[string]any{
						"street": "123 Main St",
						"city":   "New York",
						"zip":    "10001",
					},
				},
				"services": map[string]any{
					"type":     "premium",
					"features": "enhanced, priority, 24/7",
				},
			},
			map[string]any{
				"id":     "booking_2",
				"status": "pending",
				"amount": 1750.00,
				"active": false,
				"notes":  "",
				"client": map[string]any{
					"id":    "client_b",
					"name":  "TechStart Inc",
					"email": "hello@techstart.io",
					"address": map[string]any{
						"street": "456 Tech Blvd",
						"city":   "San Francisco",
						"zip":    "94102",
					},
				},
				"services": map[string]any{
					"type":     "standard",
					"features": "basic, email-only",
				},
			},
			map[string]any{
				"id":     "booking_3",
				"status": "completed",
				"amount": 3200.75,
				"active": true,
				"notes":  "Rush order: completed ahead of schedule!",
				"client": map[string]any{
					"id":    "client_c",
					"name":  "Global Dynamics",
					"email": "info@globaldyn.com",
					"address": map[string]any{
						"street": "789 Business Park",
						"city":   "Chicago",
						"zip":    "60601",
					},
				},
				"services": map[string]any{
					"type":     "enterprise",
					"features": "all-inclusive, dedicated-support",
				},
			},
		},
	}

	toon, err := Encode(data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(toon, "api_version: v2") {
		t.Error("Missing simple key-value")
	}
	if !strings.Contains(toon, "total_count: 3") {
		t.Error("Missing integer value")
	}
	if !strings.Contains(toon, "results:") {
		t.Error("Missing nested array")
	}
	if !strings.Contains(toon, "client.id") {
		t.Error("Missing nested object flattening (level 1)")
	}
	if !strings.Contains(toon, "client.address.city") {
		t.Error("Missing multi-level nesting (level 2)")
	}
	if !strings.Contains(toon, "services.type") {
		t.Error("Missing nested object flattening (services)")
	}

	diff := Diff(data)
	jsonChars := diff["json_chars"].(int)
	toonChars := diff["toon_chars"].(int)
	savingsPercent := diff["savings_percent"].(float64)

	t.Logf("JSON: %d chars", jsonChars)
	t.Logf("TOON: %d chars", toonChars)
	t.Logf("Savings: %.1f%%", savingsPercent)

	if toonChars >= jsonChars {
		t.Error("TOON should be smaller than JSON")
	}
	if savingsPercent < 20 {
		t.Errorf("Expected at least 20%% savings, got %.1f%%", savingsPercent)
	}

	decoded, err := Decode(toon)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded["api_version"] != "v2" {
		t.Error("Round-trip failed: api_version mismatch")
	}
	if decoded["total_count"] != 3 {
		t.Error("Round-trip failed: total_count mismatch")
	}

	results, ok := decoded["results"].([]any)
	if !ok || len(results) != 3 {
		t.Fatalf("Round-trip failed: results should be array of 3 items")
	}

	firstResult := results[0].(map[string]any)
	if firstResult["id"] != "booking_1" {
		t.Error("Round-trip failed: first result id mismatch")
	}
	if firstResult["amount"] != 2500.50 {
		t.Error("Round-trip failed: float precision lost")
	}
	if firstResult["active"] != true {
		t.Error("Round-trip failed: boolean lost")
	}

	client := firstResult["client"].(map[string]any)
	if client["name"] != "Acme Corp" {
		t.Error("Round-trip failed: nested object lost")
	}

	address := client["address"].(map[string]any)
	if address["city"] != "New York" {
		t.Error("Round-trip failed: multi-level nesting lost")
	}

	config := &Config{
		MinRowsForTable: 2,
		MaxFlattenDepth: 3,
		Omit:            []string{"null", "empty"},
		NumberPrecision: 2,
	}

	encoder := NewEncoder(config)
	customToon, err := encoder.Encode(data)
	if err != nil {
		t.Fatalf("Custom encoding failed: %v", err)
	}

	if !strings.Contains(customToon, "items[3]") {
		t.Error("Custom config should still create table for results")
	}

	t.Log("All features working correctly!")
}
