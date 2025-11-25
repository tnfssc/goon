package toon

import (
	"fmt"
	"strconv"
	"strings"
)

// ArrayHeaderInfo contains parsed array header information
type ArrayHeaderInfo struct {
	Key       string
	Length    int
	Delimiter string
	Fields    []string
}

// ParsePrimitiveToken parses a primitive value from a string
func ParsePrimitiveToken(token string) JsonValue {
	token = strings.TrimSpace(token)
	if token == "null" {
		return nil
	}
	if token == "true" {
		return true
	}
	if token == "false" {
		return false
	}

	// Try parsing as number
	if num, err := strconv.ParseFloat(token, 64); err == nil {
		return num
	}

	// Return as string (unquoted)
	// Handle quoted strings with escape sequences
	if strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"") && len(token) >= 2 {
		inner := token[1 : len(token)-1]
		// Unescape: process escape sequences character by character
		var result strings.Builder
		i := 0
		for i < len(inner) {
			if inner[i] == '\\' && i+1 < len(inner) {
				switch inner[i+1] {
				case '\\':
					result.WriteByte('\\')
					i += 2
				case '"':
					result.WriteByte('"')
					i += 2
				case 'n':
					result.WriteByte('\n')
					i += 2
				case 'r':
					result.WriteByte('\r')
					i += 2
				case 't':
					result.WriteByte('\t')
					i += 2
				default:
					// Unknown escape, keep as is
					result.WriteByte(inner[i])
					i++
				}
			} else {
				result.WriteByte(inner[i])
				i++
			}
		}
		return result.String()
	}

	return token
}

// ParseArrayHeaderLineTOONv2 parses a TOON v2 array header
// Format: [N]: or [N]{fields}: or [N|]: (pipe delimiter) or [N	]: (tab delimiter)
// Example: "[3]: 1,2,3" or "[2]{id,name}:" or "[3|]: a|b|c"
func ParseArrayHeaderLineTOONv2(line string) *ArrayHeaderInfo {
	startBracket := strings.Index(line, "[")
	if startBracket == -1 {
		return nil
	}
	endBracket := strings.Index(line, "]")
	if endBracket == -1 || endBracket <= startBracket {
		return nil
	}

	// Parse bracket content [N] or [N|] or [N	]
	bracketContent := line[startBracket+1 : endBracket]

	// Find end of length (digits)
	var i int
	for i = 0; i < len(bracketContent); i++ {
		if bracketContent[i] < '0' || bracketContent[i] > '9' {
			break
		}
	}

	if i == 0 {
		// No digits at start
		return nil
	}

	lengthStr := bracketContent[:i]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil
	}

	// Determine delimiter
	delimiter := "," // Default is comma
	rest := bracketContent[i:]

	if strings.HasPrefix(rest, "|") {
		delimiter = "|"
	} else if strings.HasPrefix(rest, "\t") {
		delimiter = "\t"
	}
	// If rest is empty or just whitespace, delimiter stays as comma

	// Check for fields segment after bracket: {field1,field2}
	afterBracket := strings.TrimSpace(line[endBracket+1:])
	var fields []string

	if strings.HasPrefix(afterBracket, "{") {
		closeBrace := strings.Index(afterBracket, "}")
		if closeBrace != -1 {
			fieldsContent := afterBracket[1:closeBrace]
			fields = ParseDelimitedValues(fieldsContent, delimiter)
			afterBracket = strings.TrimSpace(afterBracket[closeBrace+1:])
		}
	}

	// Check for colon
	if !strings.HasPrefix(afterBracket, ":") {
		// Not a valid array header (no colon)
		return nil
	}

	return &ArrayHeaderInfo{
		Key:       "", // Key will be set by caller
		Length:    length,
		Delimiter: delimiter,
		Fields:    fields,
	}
}

// ParseArrayHeaderLine parses a line to check if it's an array header
// Format: [key]: [length<delimiter>] [fields...]
// Example: "items: [3|] name age"
func ParseArrayHeaderLine(line string, defaultDelim string) *ArrayHeaderInfo {
	// Simplified parser:
	// 1. Check for brackets [...]
	// 2. Extract content inside brackets
	// 3. Parse length and delimiter from content

	startBracket := strings.Index(line, "[")
	if startBracket == -1 {
		return nil
	}
	endBracket := strings.Index(line, "]")
	if endBracket == -1 || endBracket < startBracket {
		return nil
	}

	// Extract key if present (before bracket)
	var key string
	beforeBracket := strings.TrimSpace(line[:startBracket])
	if strings.HasSuffix(beforeBracket, ":") {
		key = strings.TrimSuffix(beforeBracket, ":")
	}

	// Parse bracket content
	bracketContent := line[startBracket+1 : endBracket]

	// Find end of length (digits)
	var i int
	for i = 0; i < len(bracketContent); i++ {
		if bracketContent[i] < '0' || bracketContent[i] > '9' {
			break
		}
	}

	if i == 0 {
		// No digits at start, probably inline array [item1, item2]
		return nil
	}

	lengthStr := bracketContent[:i]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return nil
	}

	// Check for delimiter/separator
	// We REQUIRE a pipe '|' to distinguish header from inline array like [1, 2]
	// So [2] is inline array. [2|] is header. [2|name,role] is header.

	rest := bracketContent[i:]
	if len(rest) == 0 {
		// Just [2]. Treat as inline array.
		return nil
	}

	delimiter := defaultDelim
	var fields []string

	// Must start with |
	if strings.HasPrefix(rest, "|") {
		delimiter = "|"
		rest = rest[1:]

		// Parse fields inside bracket
		if len(rest) > 0 {
			// Fields are separated by the delimiter?
			// If delimiter is |, fields are name|role?
			// Or name, role?
			// Usually | sets the delimiter for the rows.
			// But fields in header?
			// Let's assume fields in header are comma separated if delimiter is |, or same delimiter?
			// If [2|name, role], delimiter is |.
			// ParseDelimitedValues(rest, "|") -> "name, role". 1 field.
			// ParseDelimitedValues(rest, ",") -> "name", "role". 2 fields.

			// Let's try to detect if we should use comma for fields?
			// Or just use the delimiter.
			// If I use [2|name|role], then delimiter | works.

			fields = ParseDelimitedValues(rest, delimiter)
		}
	} else {
		// Starts with something else (e.g. comma).
		// [2, 3]. Treat as inline array.
		return nil
	}

	info := &ArrayHeaderInfo{
		Key:       key,
		Length:    length,
		Delimiter: delimiter,
		Fields:    fields,
	}

	// Check for fields (after bracket) - append if found
	afterBracket := strings.TrimSpace(line[endBracket+1:])
	if afterBracket != "" {
		if strings.HasPrefix(afterBracket, "{") && strings.HasSuffix(afterBracket, "}") {
			fieldsContent := afterBracket[1 : len(afterBracket)-1]
			moreFields := ParseDelimitedValues(fieldsContent, delimiter)
			info.Fields = append(info.Fields, moreFields...)
		} else {
			// Assume space separated or delimited by delimiter?
			// Let's use ParseDelimitedValues with delimiter
			moreFields := ParseDelimitedValues(afterBracket, delimiter)
			info.Fields = append(info.Fields, moreFields...)
		}
	}

	return info
}

// ParseDelimitedValues splits a string by delimiter, respecting quotes
func ParseDelimitedValues(content string, delimiter string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false
	delimRunes := []rune(delimiter)
	delimLen := len(delimRunes)

	runes := []rune(content)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '"' {
			inQuote = !inQuote
			current.WriteRune(r)
			continue
		}

		// Check for delimiter
		isDelim := false
		if !inQuote {
			if len(delimiter) == 1 {
				if r == delimRunes[0] {
					isDelim = true
				}
			} else {
				// Multi-character delimiter support if needed, though usually 1 char
				if i+delimLen <= len(runes) && string(runes[i:i+delimLen]) == delimiter {
					isDelim = true
					i += delimLen - 1 // Skip rest of delimiter
				}
			}
		}

		if isDelim {
			parts = append(parts, current.String())
			current.Reset()
		} else {
			current.WriteRune(r)
		}
	}
	parts = append(parts, current.String())

	// Trim spaces and unquote if necessary (ParsePrimitiveToken handles unquoting)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// ParseInlineArray parses an inline array string e.g. "[item1, item2]"
func ParseInlineArray(content string) (JsonArray, error) {
	content = strings.TrimSpace(content)
	if !strings.HasPrefix(content, "[") || !strings.HasSuffix(content, "]") {
		return nil, fmt.Errorf("invalid inline array format")
	}

	inner := content[1 : len(content)-1]
	if strings.TrimSpace(inner) == "" {
		return make(JsonArray, 0), nil
	}

	// Default delimiter is comma
	// TODO: Support custom delimiters if specified in some header-like way?
	// For standard inline arrays, it's comma.
	parts := ParseDelimitedValues(inner, ",")

	arr := make(JsonArray, len(parts))
	for i, p := range parts {
		arr[i] = ParsePrimitiveToken(p)
	}
	return arr, nil
}

// IsArrayHeaderAfterHyphen checks if the content after a hyphen looks like an array header
func IsArrayHeaderAfterHyphen(content string) bool {
	// Must start with [
	return strings.HasPrefix(strings.TrimSpace(content), "[")
}
