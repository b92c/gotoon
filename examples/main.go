package main

import (
"fmt"

"github.com/b92c/gotoon"
)

func main() {
// Example 1: Simple encoding
fmt.Println("=== Example 1: Simple Key-Value ===")
data1 := map[string]any{
"name": "John Doe",
"age":  30,
"city": "New York",
}
toon1, _ := gotoon.Encode(data1)
fmt.Println(toon1)
fmt.Println()

// Example 2: Nested objects
fmt.Println("=== Example 2: Nested Objects ===")
data2 := map[string]any{
"user": map[string]any{
"id":    1,
"name":  "Alice",
"email": "alice@example.com",
"profile": map[string]any{
"bio":      "Software Engineer",
"location": "San Francisco",
},
},
}
toon2, _ := gotoon.Encode(data2)
fmt.Println(toon2)
fmt.Println()

// Example 3: Array of objects with nested structures
fmt.Println("=== Example 3: Nested Array (Key Feature) ===")
data3 := []any{
map[string]any{
"id":   1,
"name": "Alice",
"role": map[string]any{
"id":    "admin",
"level": 10,
},
},
map[string]any{
"id":   2,
"name": "Bob",
"role": map[string]any{
"id":    "user",
"level": 1,
},
},
}
toon3, _ := gotoon.Encode(data3)
fmt.Println(toon3)
fmt.Println()

// Example 4: Complex nested structure
fmt.Println("=== Example 4: Complex Multi-Level Nesting ===")
data4 := map[string]any{
"orders": []any{
map[string]any{
"id":     "ord_1",
"status": "shipped",
"customer": map[string]any{
"id":   "cust_1",
"name": "Alice",
},
"total": 99.99,
},
map[string]any{
"id":     "ord_2",
"status": "pending",
"customer": map[string]any{
"id":   "cust_2",
"name": "Bob",
},
"total": 149.50,
},
},
}
toon4, _ := gotoon.Encode(data4)
fmt.Println(toon4)
fmt.Println()

// Example 5: Measure savings
fmt.Println("=== Example 5: Token Savings ===")
diff := gotoon.Diff(data4)
fmt.Printf("JSON characters: %d\n", diff["json_chars"])
fmt.Printf("TOON characters: %d\n", diff["toon_chars"])
fmt.Printf("Saved characters: %d\n", diff["saved_chars"])
fmt.Printf("Savings: %.1f%%\n", diff["savings_percent"])
fmt.Println()

// Example 6: Custom configuration
fmt.Println("=== Example 6: Custom Config (Omit nulls, Key aliases) ===")
config := &gotoon.Config{
MinRowsForTable: 2,
MaxFlattenDepth: 3,
EscapeStyle:     "backslash",
Omit:            []string{"null", "empty"},
OmitKeys:        []string{"created_at", "updated_at"},
KeyAliases: map[string]string{
"description": "desc",
},
NumberPrecision: 2,
}

encoder := gotoon.NewEncoder(config)

data6 := map[string]any{
"name":        "Product A",
"description": "A wonderful product",
"price":       19.999,
"stock":       nil,
"notes":       "",
"created_at":  "2024-01-01",
}

toon6, _ := encoder.Encode(data6)
fmt.Println(toon6)
fmt.Println()

// Example 7: Encode only specific keys
fmt.Println("=== Example 7: Encode Only Specific Keys ===")
users := []any{
map[string]any{"id": 1, "name": "Alice", "email": "alice@example.com", "password": "secret123"},
map[string]any{"id": 2, "name": "Bob", "email": "bob@example.com", "password": "secret456"},
}

toon7, _ := gotoon.Only(users, []string{"id", "name"})
fmt.Println(toon7)
fmt.Println()

// Example 8: Round-trip (Encode + Decode)
fmt.Println("=== Example 8: Round-Trip ===")
original := map[string]any{
"users": []any{
map[string]any{
"id":   1,
"name": "Alice",
"meta": map[string]any{"role": "admin", "active": true},
},
map[string]any{
"id":   2,
"name": "Bob",
"meta": map[string]any{"role": "user", "active": false},
},
},
}

encoded, _ := gotoon.Encode(original)
fmt.Println("Encoded:")
fmt.Println(encoded)
fmt.Println()

decoded, _ := gotoon.Decode(encoded)
fmt.Println("Decoded:")
fmt.Printf("%+v\n", decoded)
}
