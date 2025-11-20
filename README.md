# goon

> **Go** implementation of the **TOON** data format.

[![Go Reference](https://pkg.go.dev/badge/github.com/tnfssc/goon.svg)](https://pkg.go.dev/github.com/tnfssc/goon)
[![Go Report Card](https://goreportcard.com/badge/github.com/tnfssc/goon)](https://goreportcard.com/report/github.com/tnfssc/goon)
[![License](https://img.shields.io/github/license/tnfssc/goon)](LICENSE.md)

**goon** is a high-performance, strictly compliant Go library and CLI for parsing and generating [TOON](https://github.com/toon-format/spec) (The Object-Oriented Notation) data.

## Features

- 🚀 **Fast & Efficient**: Built with performance in mind using a custom scanner and recursive descent parser.
- 🔄 **Marshal/Unmarshal**: Direct conversion between Go structs and TOON format with struct tag support (like `encoding/json`).
- 🔒 **Strict Mode**: Optional strict validation to ensure your TOON files are perfectly formatted (no tabs, correct indentation).
- 🛠️ **CLI Tools**: Includes `json2toon` and `toon2json` for easy integration into existing workflows.
- 📦 **Zero Dependencies**: The core library has no external dependencies.

## Installation

### Quick Install (Script)

**Linux & macOS**

```bash
curl -fsSL https://raw.githubusercontent.com/tnfssc/goon/develop/scripts/install.sh | sh
```

**Windows (PowerShell)**

```bash
irm https://raw.githubusercontent.com/tnfssc/goon/develop/scripts/install.ps1 | iex
```

### Manual Installation

#### Library

To use Goon in your Go project:

```bash
go get github.com/tnfssc/goon
```

#### CLI Tools

To install the command-line tool:

```bash
go install github.com/tnfssc/goon/cmd/goon@latest
```

## Usage

### Library

```go
package main

import (
	"fmt"
	"log"

	"github.com/tnfssc/goon/pkg/toon"
)

type Config struct {
	Name     string   `toon:"name"`
	Version  string   `toon:"version"`
	Features []string `toon:"features"`
}

func main() {
	// Marshal Go struct to TOON (like json.Marshal)
	config := Config{
		Name:     "MyApp",
		Version:  "1.0.0",
		Features: []string{"fast", "reliable", "simple"},
	}

	toonData, err := toon.Marshal(config, toon.EncodeOptions{IndentSize: 2})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(toonData))

	// Unmarshal TOON to Go struct (like json.Unmarshal)
	var decoded Config
	err = toon.Unmarshal(toonData, &decoded, toon.DecodeOptions{IndentSize: 2})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Decoded: %+v\n", decoded)
}
```

### CLI

Convert JSON to TOON:

```bash
goon encode input.json > output.toon
```

Convert TOON back to JSON:

```bash
goon decode input.toon > output.json
```

**Pipe from stdin:**

```bash
# JSON to TOON
cat data.json | goon encode > data.toon

# TOON to JSON
cat config.toon | goon decode > config.json

# Chain with other tools
curl https://api.example.com/data.json | goon encode > api-data.toon
```

## TOON Syntax Support

Goon supports the full TOON specification, including:

- **Key-Value Pairs**: Simple `key: value` syntax.
- **Nested Objects**: Indentation-based hierarchy.
- **Arrays**:
  - List format: `key: [length|]` followed by `- item` lines.
  - _Coming Soon_: Inline arrays and tabular arrays.
- **Primitives**: Strings, numbers, booleans, and null.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
