package gotoon

// Config holds configuration options for TOON encoding and decoding.
type Config struct {
	// MinRowsForTable is the minimum number of items required to use table format.
	// Arrays with fewer items will be encoded as regular YAML-like objects.
	MinRowsForTable int

	// MaxFlattenDepth controls how many levels deep to flatten nested objects.
	// Objects nested deeper than this will be JSON-encoded as strings.
	MaxFlattenDepth int

	// EscapeStyle determines how to escape special characters in string values.
	// Currently only "backslash" is supported.
	EscapeStyle string

	// Omit specifies which value types to omit from output.
	// Supported values: "null", "empty", "false", "all"
	Omit []string

	// OmitKeys specifies keys that should always be omitted from output.
	OmitKeys []string

	// KeyAliases maps long key names to shorter aliases to save tokens.
	KeyAliases map[string]string

	// DateFormat specifies the format for time.Time objects and ISO date strings.
	// Uses Go's time format syntax. When empty, dates are passed through as-is.
	DateFormat string

	// TruncateStrings specifies the maximum length for string values.
	// Strings exceeding this length will be truncated with "...".
	// When 0, strings are not truncated.
	TruncateStrings int

	// NumberPrecision specifies the maximum decimal places for float values.
	// When -1, floats are passed through as-is.
	NumberPrecision int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		MinRowsForTable: 2,
		MaxFlattenDepth: 3,
		EscapeStyle:     "backslash",
		Omit:            []string{},
		OmitKeys:        []string{},
		KeyAliases:      make(map[string]string),
		DateFormat:      "",
		TruncateStrings: 0,
		NumberPrecision: -1,
	}
}

// shouldOmit checks if a value type should be omitted.
func (c *Config) shouldOmit(valueType string) bool {
	for _, t := range c.Omit {
		if t == "all" || t == valueType {
			return true
		}
	}
	return false
}

// shouldOmitKey checks if a key should be omitted.
func (c *Config) shouldOmitKey(key string) bool {
	for _, k := range c.OmitKeys {
		if k == key {
			return true
		}
	}
	return false
}

// formatKey applies key aliases or returns the original key.
func (c *Config) formatKey(key string) string {
	if alias, ok := c.KeyAliases[key]; ok {
		return alias
	}
	return safeKey(key)
}
