package toon

// Delimiters
const (
	DelimiterComma = ","
	DelimiterPipe  = "|"
	DelimiterTab   = "\t"
)

// Constants
const (
	IndentSize     = 2
	ListItemPrefix = "- "
)

// JsonValue represents any JSON-compatible value
type JsonValue interface{}

// JsonObject represents a JSON object
type JsonObject map[string]JsonValue

// JsonArray represents a JSON array
type JsonArray []JsonValue

// TOONVersion represents the TOON format version
type TOONVersion int

const (
	// V2 is the default TOON v2 format with inline primitive arrays
	V2 TOONVersion = iota
	// V1 is the legacy TOON v1 format with list-style arrays
	V1
)

// DecodeOptions configuration for decoding
type DecodeOptions struct {
	IndentSize int
	Strict     bool
}

// EncodeOptions configuration for encoding
type EncodeOptions struct {
	IndentSize int
	Version    TOONVersion // Default is V2, set to V1 for legacy format
}
