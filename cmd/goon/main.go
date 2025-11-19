package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/tnfssc/goon/pkg/toon"
)

// version is set via ldflags during build
var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "encode":
		runEncode()
	case "decode":
		runDecode()
	case "--version", "-v", "version":
		fmt.Printf("goon version %s\n", version)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: goon <command> [arguments]")
	fmt.Println("Commands:")
	fmt.Println("  encode   Convert JSON to TOON")
	fmt.Println("  decode   Convert TOON to JSON")
	fmt.Println("  version  Show version information")
}

func runEncode() {
	cmd := flag.NewFlagSet("encode", flag.ExitOnError)
	indent := cmd.Int("indent", 2, "Indentation size")
	cmd.Parse(os.Args[2:])

	input, err := readInput(cmd.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	var data interface{}
	if err := json.Unmarshal(input, &data); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	output, err := toon.Encode(data, toon.EncodeOptions{IndentSize: *indent})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding TOON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(output)
}

func runDecode() {
	cmd := flag.NewFlagSet("decode", flag.ExitOnError)
	indent := cmd.Int("indent", 2, "Indentation size for output JSON")
	strict := cmd.Bool("strict", true, "Enable strict mode parsing")
	cmd.Parse(os.Args[2:])

	input, err := readInput(cmd.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	data, err := toon.Decode(string(input), toon.DecodeOptions{IndentSize: 2, Strict: *strict})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding TOON: %v\n", err)
		os.Exit(1)
	}

	// Marshal with indentation
	indentStr := ""
	for i := 0; i < *indent; i++ {
		indentStr += " "
	}

	output, err := json.MarshalIndent(data, "", indentStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))
}

func readInput(args []string) ([]byte, error) {
	if len(args) > 0 {
		return os.ReadFile(args[0])
	}
	return io.ReadAll(os.Stdin)
}
