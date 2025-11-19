package main

import (
	"fmt"
	"os"

	"github.com/tnfssc/goon/pkg/toon"
)

func main() {
	fmt.Println("=== Testing Divide-by-Zero Bug Fix ===")

	// Create a map with nested structure
	data := map[string]interface{}{
		"manifest-version": "1",
		"name":             "test-skill",
		"version":          "1.0.0",
		"skill": map[string]interface{}{
			"main": "SKILL.md",
		},
	}

	fmt.Println("Step 1: Encode with IndentSize=2...")
	toonStr, err := toon.Encode(data, toon.EncodeOptions{IndentSize: 2})
	if err != nil {
		fmt.Printf("ERROR encoding: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("TOON output:")
	fmt.Println("---")
	fmt.Println(toonStr)
	fmt.Println("---")

	fmt.Println("Step 2: Decode with empty DecodeOptions (should default to IndentSize=2)...")
	decoded, err := toon.Decode(toonStr, toon.DecodeOptions{})
	if err != nil {
		fmt.Printf("ERROR decoding: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ Successfully decoded: %+v\n\n", decoded)

	fmt.Println("Step 3: Encode with empty EncodeOptions (should default to IndentSize=2)...")
	toonStr2, err := toon.Encode(data, toon.EncodeOptions{})
	if err != nil {
		fmt.Printf("ERROR encoding: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("TOON output:")
	fmt.Println("---")
	fmt.Println(toonStr2)
	fmt.Println("---")

	fmt.Println("✓ All tests passed! Bug is fixed.")
}
