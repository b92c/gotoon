# GoToon Quick Start Guide

## Installation

```bash
go get github.com/b92c/gotoon
```

## Basic Usage

### Simple Encoding

```go
package main

import (
    "fmt"
    "github.com/b92c/gotoon"
)

func main() {
    data := map[string]any{
        "name": "Alice",
        "age":  30,
    }
    
    toon, _ := gotoon.Encode(data)
    fmt.Println(toon)
    // Output:
    // name: Alice
    // age: 30
}
```

### Nested Objects (Key Feature!)

```go
users := []any{
    map[string]any{
        "id":   1,
        "name": "Alice",
        "role": map[string]any{"id": "admin", "level": 10},
    },
    map[string]any{
        "id":   2,
        "name": "Bob",
        "role": map[string]any{"id": "user", "level": 1},
    },
}

toon, _ := gotoon.Encode(users)
fmt.Println(toon)
// Output:
// items[2]{id,name,role.id,role.level}:
//   1,Alice,admin,10
//   2,Bob,user,1
```

### Decoding

```go
toonStr := `items[2]{id,name}:
  1,Alice
  2,Bob`

decoded, _ := gotoon.Decode(toonStr)
fmt.Printf("%+v\n", decoded)
// Output: map[items:[map[id:1 name:Alice] map[id:2 name:Bob]]]
```

### Measure Token Savings

```go
data := []any{
    map[string]any{"id": 1, "name": "Alice", "email": "alice@example.com"},
    map[string]any{"id": 2, "name": "Bob", "email": "bob@example.com"},
}

diff := gotoon.Diff(data)
fmt.Printf("JSON: %d chars, TOON: %d chars, Saved: %.1f%%\n",
    diff["json_chars"],
    diff["toon_chars"],
    diff["savings_percent"])
```

## Custom Configuration

```go
config := &gotoon.Config{
    MinRowsForTable: 2,
    MaxFlattenDepth: 3,
    Omit:            []string{"null", "empty"},
    OmitKeys:        []string{"created_at", "updated_at"},
    KeyAliases: map[string]string{
        "description": "desc",
    },
    NumberPrecision: 2,
    TruncateStrings: 100,
}

encoder := gotoon.NewEncoder(config)
toon, _ := encoder.Encode(data)
```

## Use Cases

### MCP Servers

```go
func HandleMCPTool() string {
    users := db.GetUsers(100)
    toon, _ := gotoon.Encode(map[string]any{
        "count": len(users),
        "users": users,
    })
    return toon
}
```

### LLM Context Optimization

```go
context, _ := gotoon.Encode(map[string]any{
    "conversation": messages,
    "user_data":    userProfile,
    "history":      recentActions,
})

response := llm.Chat([]Message{
    {Role: "system", Content: "Context:\n" + context},
    {Role: "user", Content: userQuestion},
})
```

## Why TOON?

- **40-60% smaller** than JSON
- **Full round-trip fidelity** - no data loss
- **Intelligent nested object handling** - automatic dot notation flattening
- **Type preservation** - int, float, bool, nil all preserved
- **Configurable** - omit nulls, truncate strings, alias keys

Perfect for LLM applications where every token counts!
