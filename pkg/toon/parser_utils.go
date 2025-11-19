package toon

import (
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
	// Note: In a full implementation, we might want to handle quoted strings explicitly
	// to support escape sequences, but for now we'll assume simple strings.
	if strings.HasPrefix(token, "\"") && strings.HasSuffix(token, "\"") {
		return token[1 : len(token)-1]
	}

	return token
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
	delimiter := defaultDelim

	if strings.HasSuffix(bracketContent, "|") {
		delimiter = "|"
		bracketContent = bracketContent[:len(bracketContent)-1]
	} else if strings.HasSuffix(bracketContent, ",") {
		delimiter = ","
		bracketContent = bracketContent[:len(bracketContent)-1]
	} else if strings.HasSuffix(bracketContent, "\t") {
		delimiter = "\t"
		bracketContent = bracketContent[:len(bracketContent)-1]
	}

	length, err := strconv.Atoi(strings.TrimSpace(bracketContent))
	if err != nil {
		return nil
	}

	info := &ArrayHeaderInfo{
		Key:       key,
		Length:    length,
		Delimiter: delimiter,
	}

	// Check for fields (after bracket)
	afterBracket := strings.TrimSpace(line[endBracket+1:])
	if afterBracket != "" {
		// Check for braces { fields } if strictly following spec, or just space separated?
		// Reference implementation checks for braces `{...}` for fields.
		// Let's assume fields are just space separated for now or in braces.
		// If starts with {, find }
		if strings.HasPrefix(afterBracket, "{") && strings.HasSuffix(afterBracket, "}") {
			fieldsContent := afterBracket[1 : len(afterBracket)-1]
			info.Fields = ParseDelimitedValues(fieldsContent, delimiter)
		} else {
			// Fallback or simple space separated?
			// Reference implementation seems to require braces for fields.
			// "Check for fields segment (braces come after bracket)"
			// So we should look for braces.
			// If no braces, maybe no fields.
		}
	}

	return info
}

// ParseDelimitedValues splits a string by delimiter
func ParseDelimitedValues(content string, delimiter string) []string {
	// This is a naive split. A real implementation should respect quotes.
	// For now, we'll use simple split and trim.
	parts := strings.Split(content, delimiter)
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// IsArrayHeaderAfterHyphen checks if the content after a hyphen looks like an array header
func IsArrayHeaderAfterHyphen(content string) bool {
	// Must start with [
	return strings.HasPrefix(strings.TrimSpace(content), "[")
}
