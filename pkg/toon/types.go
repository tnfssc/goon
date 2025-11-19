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

// DecodeOptions configuration for decoding
type DecodeOptions struct {
	IndentSize int
	Strict     bool
}

// EncodeOptions configuration for encoding
type EncodeOptions struct {
	IndentSize int
}
