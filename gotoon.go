// Package gotoon provides a Token-Optimized Object Notation (TOON) encoder and decoder.
// TOON is designed to reduce token usage when sending data to LLMs while maintaining
// full round-trip fidelity.
package gotoon

import "encoding/json"

var (
	defaultEncoder = NewEncoder(DefaultConfig())
	defaultDecoder = NewDecoder(DefaultConfig())
)

// Encode converts data to TOON format using the default encoder.
func Encode(data any) (string, error) {
	return defaultEncoder.Encode(data)
}

// Decode converts a TOON format string to Go data structures using the default decoder.
func Decode(toon string) (map[string]any, error) {
	return defaultDecoder.Decode(toon)
}

// Diff estimates token savings between JSON and TOON formats.
// Returns a map with json_chars, toon_chars, saved_chars, and savings_percent.
func Diff(data any) map[string]any {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return map[string]any{
			"json_chars":      0,
			"toon_chars":      0,
			"saved_chars":     0,
			"savings_percent": 0.0,
		}
	}

	toon, err := Encode(data)
	if err != nil {
		return map[string]any{
			"json_chars":      len(jsonBytes),
			"toon_chars":      0,
			"saved_chars":     0,
			"savings_percent": 0.0,
		}
	}

	jsonLen := len(jsonBytes)
	toonLen := len(toon)
	saved := jsonLen - toonLen
	savingsPercent := 0.0
	if jsonLen > 0 {
		savingsPercent = float64(saved) / float64(jsonLen) * 100
	}

	return map[string]any{
		"json_chars":      jsonLen,
		"toon_chars":      toonLen,
		"saved_chars":     saved,
		"savings_percent": savingsPercent,
	}
}

// Only encodes only specific keys from the data.
func Only(data any, keys []string) (string, error) {
	filtered := filterKeys(data, keys)
	return Encode(filtered)
}

// filterKeys recursively filters data to only include specified keys.
func filterKeys(data any, keys []string) any {
	if slice, ok := data.([]any); ok {
		result := make([]any, len(slice))
		for i, item := range slice {
			result[i] = filterKeys(item, keys)
		}
		return result
	}

	if m, ok := data.(map[string]any); ok {
		filtered := make(map[string]any)
		for _, key := range keys {
			if val, exists := m[key]; exists {
				filtered[key] = val
			}
		}
		return filtered
	}

	return data
}
