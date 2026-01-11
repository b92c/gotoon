# GoToon

[![Go Reference](https://pkg.go.dev/badge/github.com/b92c/gotoon.svg)](https://pkg.go.dev/github.com/b92c/gotoon)
[![Go Report Card](https://goreportcard.com/badge/github.com/b92c/gotoon)](https://goreportcard.com/report/github.com/b92c/gotoon)


Token-Optimized Object Notation encoder/decoder for Go with intelligent nested object handling.

TOON is a compact, YAML-like format designed to reduce token usage when sending data to LLMs. This package achieves **40-60% token reduction** compared to JSON while maintaining full round-trip fidelity.

## Installation

```bash
go get github.com/b92c/gotoon
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/b92c/gotoon"
)

func main() {
    data := map[string]any{
        "users": []map[string]any{
            {"id": 1, "name": "Alice", "role": map[string]any{"id": "admin", "level": 10}},
            {"id": 2, "name": "Bob", "role": map[string]any{"id": "user", "level": 1}},
        },
    }

    // Encode to TOON
    toon, _ := gotoon.Encode(data)
    fmt.Println(toon)

    // Decode back to map
    original, _ := gotoon.Decode(toon)
    fmt.Printf("%+v\n", original)
}
```

**Output:**
```
users:
  items[2]{id,name,role.id,role.level}:
    1,Alice,admin,10
    2,Bob,user,1
```

## Why TOON?

When building MCP servers or LLM-powered applications, every token counts. JSON's verbosity wastes context window space with repeated keys and structural characters.

**JSON (398 bytes):**
```json
{"orders":[{"id":"ord_1","status":"shipped","customer":{"id":"cust_1","name":"Alice"},"total":99.99},{"id":"ord_2","status":"pending","customer":{"id":"cust_2","name":"Bob"},"total":149.50}]}
```

**TOON (186 bytes) - 53% smaller:**
```
orders:
  items[2]{id,status,customer.id,customer.name,total}:
    ord_1,shipped,cust_1,Alice,99.99
    ord_2,pending,cust_2,Bob,149.5
```

## Features

### Nested Object Flattening

The key differentiator. Arrays containing objects with nested properties are automatically flattened using dot notation:

```go
data := []map[string]any{
    {"id": 1, "author": map[string]any{"name": "Jane", "email": "jane@example.com"}},
    {"id": 2, "author": map[string]any{"name": "John", "email": "john@example.com"}},
}

toon, _ := gotoon.Encode(data)
// items[2]{id,author.name,author.email}:
//   1,Jane,jane@example.com
//   2,John,john@example.com

decoded, _ := gotoon.Decode(toon)
// Returns original nested structure
```

### Multi-Level Nesting

Handles deeply nested structures:

```go
data := []map[string]any{
    {
        "id": 1,
        "product": map[string]any{
            "name": "Widget",
            "category": map[string]any{"id": "cat_1", "name": "Electronics"},
        },
    },
}

toon, _ := gotoon.Encode(data)
// items[1]{id,product.name,product.category.id,product.category.name}:
//   1,Widget,cat_1,Electronics
```

### Type Preservation

All scalar types are preserved through encode/decode:

```go
data := map[string]any{
    "count":   42,
    "price":   19.99,
    "active":  true,
    "deleted": false,
    "notes":   nil,
}

decoded, _ := gotoon.Decode(gotoon.Encode(data))
// Types are preserved: int, float, bool, nil
```

### Special Character Escaping

Commas, colons, and newlines in values are automatically escaped:

```go
data := map[string]any{"message": "Hello, World: How are you?"}
toon, _ := gotoon.Encode(data)
// message: Hello\, World\: How are you?
```

## Configuration

Create a custom encoder/decoder with options:

```go
config := &gotoon.Config{
    // Arrays with fewer items use regular object format instead of tables
    MinRowsForTable: 2,

    // How deep to flatten nested objects (deeper = JSON string)
    MaxFlattenDepth: 3,

    // Escape style for special characters
    EscapeStyle: "backslash",

    // Omit values to save tokens
    Omit: []string{"null", "empty"},

    // Always skip these keys
    OmitKeys: []string{"created_at", "updated_at"},

    // Shorten verbose keys
    KeyAliases: map[string]string{
        "description":     "desc",
        "organization_id": "org_id",
    },

    // Format dates
    DateFormat: "2006-01-02",

    // Truncate long strings (adds ... suffix)
    TruncateStrings: 100,

    // Limit decimal places for floats
    NumberPrecision: 2,
}

encoder := gotoon.NewEncoder(config)
decoder := gotoon.NewDecoder(config)

toon, _ := encoder.Encode(data)
decoded, _ := decoder.Decode(toon)
```

### Token-Saving Options

```go
config := &gotoon.Config{
    // Omit values to save tokens: 'null', 'empty', 'false', or 'all'
    Omit: []string{"null", "empty"},

    // Always skip these keys
    OmitKeys: []string{"created_at", "updated_at"},

    // Shorten verbose keys
    KeyAliases: map[string]string{
        "description":     "desc",
        "organization_id": "org_id",
    },
}
```

### Value Transformation

```go
config := &gotoon.Config{
    // Format dates (time.Time objects and ISO strings)
    DateFormat: "2006-01-02",

    // Truncate long strings (adds ... suffix)
    TruncateStrings: 100,

    // Limit decimal places for floats
    NumberPrecision: 2,
}
```

## Utility Functions

### Measure Savings

```go
data := getUsersWithRoles()

diff := gotoon.Diff(data)
// map[string]any{
//     "json_chars":      12500,
//     "toon_chars":      5200,
//     "saved_chars":     7300,
//     "savings_percent": 58.4,
// }
```

### Encode Specific Keys Only

```go
users := getAllUsers()

// Only include id and name, exclude email, password, etc.
toon, _ := gotoon.Only(users, []string{"id", "name"})
```

## Use Cases

### MCP Servers

Reduce token usage when returning data from MCP tool calls:

```go
func HandleListUsers() (string, error) {
    users := db.GetUsersWithRoles(100)
    
    return gotoon.Encode(map[string]any{
        "count": len(users),
        "users": users,
    })
}
```

### LLM Context

Pack more data into your context window:

```go
context, _ := gotoon.Encode(map[string]any{
    "conversation":   messages,
    "user_profile":   user,
    "recent_orders":  orders,
})

response := llm.Chat([]Message{
    {Role: "system", Content: "Context:\n" + context},
    {Role: "user", Content: question},
})
```

### API Responses

Optional TOON responses for token-conscious clients:

```go
func HandleGetProducts(w http.ResponseWriter, r *http.Request) {
    data := db.GetProducts()
    
    if r.Header.Get("Accept") == "application/toon" {
        toon, _ := gotoon.Encode(data)
        w.Header().Set("Content-Type", "application/toon")
        w.Write([]byte(toon))
        return
    }
    
    json.NewEncoder(w).Encode(data)
}
```

## Benchmarks

Real-world benchmarks from production applications with 17,000+ records:

| Data Type | JSON | TOON | Savings |
|-----------|------|------|---------|
| 50 records (nested objects) | 13,055 bytes | 5,080 bytes | **61%** |
| 100 records (nested objects) | 26,156 bytes | 10,185 bytes | **61%** |
| 500 records (nested objects) | 129,662 bytes | 49,561 bytes | **62%** |
| 1,000 records (nested objects) | 258,965 bytes | 98,629 bytes | **62%** |
| 100 records (mixed nesting) | 43,842 bytes | 26,267 bytes | **40%** |
| Single object | 169 bytes | 124 bytes | **27%** |

### Token Impact

For a typical paginated API response (50 records):
- **JSON**: ~3,274 tokens
- **TOON**: ~1,279 tokens
- **Saved**: ~2,000 tokens per request

## Testing

```bash
go test ./...
go test -bench=. -benchmem
```

## Requirements

- Go 1.19+

## Credits

Based on [Laravel TOON](https://github.com/mischasigtermans/laravel-toon) by Mischa Sigtermans

## License

MIT
