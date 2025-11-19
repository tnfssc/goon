package toon

import (
	"fmt"
	"strings"
)

// Decode parses TOON content into a JsonValue
func Decode(source string, options DecodeOptions) (JsonValue, error) {
	scanResult, err := ScanLines(source, options.IndentSize, options.Strict)
	if err != nil {
		return nil, err
	}

	cursor := NewLineCursor(scanResult.Lines, scanResult.BlankLines)
	return decodeValueFromLines(cursor, options)
}

func decodeValueFromLines(cursor *LineCursor, options DecodeOptions) (JsonValue, error) {
	first := cursor.Peek()
	if first == nil {
		return nil, fmt.Errorf("no content to decode")
	}

	// Check for root array
	// TODO: Implement root array detection logic if needed, for now assume object or primitive
	// The TS implementation checks `isArrayHeaderAfterHyphen` but that seems specific to list items?
	// Actually it checks `isArrayHeaderAfterHyphen` on the first line content.

	// Check for root array
	if IsArrayHeaderAfterHyphen(first.Content) {
		headerInfo := ParseArrayHeaderLine(first.Content, DelimiterComma)
		if headerInfo != nil {
			cursor.Advance()
			return decodeArrayFromHeader(headerInfo, "", cursor, 0, options)
		} else {
			// Try inline array
			if arr, err := ParseInlineArray(first.Content); err == nil {
				return arr, nil
			}
		}
	}

	// Check for single primitive
	if cursor.Length() == 1 && !strings.Contains(first.Content, ":") {
		return ParsePrimitiveToken(first.Content), nil
	}

	return decodeObject(cursor, 0, options)
}

func decodeObject(cursor *LineCursor, baseDepth int, options DecodeOptions) (JsonObject, error) {
	obj := make(JsonObject)

	// Detect depth of first field
	var computedDepth int = -1

	for !cursor.AtEnd() {
		line := cursor.Peek()
		if line == nil || line.Depth < baseDepth {
			break
		}

		if computedDepth == -1 && line.Depth >= baseDepth {
			computedDepth = line.Depth
		}

		if line.Depth == computedDepth {
			cursor.Advance()
			key, value, err := decodeKeyValue(line.Content, cursor, computedDepth, options)
			if err != nil {
				return nil, err
			}
			obj[key] = value
		} else {
			break
		}
	}

	return obj, nil
}

func decodeKeyValue(content string, cursor *LineCursor, baseDepth int, options DecodeOptions) (string, JsonValue, error) {
	// Simple key parsing (split by first colon)
	parts := strings.SplitN(content, ":", 2)
	key := strings.TrimSpace(parts[0])

	if len(parts) < 2 {
		// Should not happen if called correctly, or maybe it's a key without value (empty object?)
		// If no colon, it might be an error or specific syntax.
		// For now assume key: value
		return key, nil, fmt.Errorf("invalid key-value pair: %s", content)
	}

	rest := strings.TrimSpace(parts[1])

	// No value after colon - nested object or empty
	if rest == "" {
		nextLine := cursor.Peek()
		if nextLine != nil && nextLine.Depth > baseDepth {
			nested, err := decodeObject(cursor, baseDepth+1, options)
			return key, nested, err
		}
		return key, make(JsonObject), nil
	}

	// Check for array header first (before parsing key)
	// Actually, decodeKeyValue is called with the line content.
	// If the line IS an array header, it's not a key-value pair.
	// But decodeKeyValue is called by decodeObject which expects key-value.
	// If we are here, we split by colon.

	// Check for array header
	// e.g. "key: 3 |" -> rest is "3 |"
	if IsArrayHeaderAfterHyphen(rest) {
		headerInfo := ParseArrayHeaderLine(rest, DelimiterComma)
		if headerInfo != nil {
			// It is an array header!
			val, err := decodeArrayFromHeader(headerInfo, "", cursor, baseDepth, options)
			return key, val, err
		} else {
			// Try inline array
			if arr, err := ParseInlineArray(rest); err == nil {
				return key, arr, nil
			}
		}
	}

	// Inline primitive
	return key, ParsePrimitiveToken(rest), nil
}

func decodeArrayFromHeader(header *ArrayHeaderInfo, inlineValues string, cursor *LineCursor, baseDepth int, options DecodeOptions) (JsonArray, error) {
	if inlineValues != "" {
		// Inline primitive array
		// TODO: Implement inline array decoding
		return nil, fmt.Errorf("inline arrays not yet implemented")
	}

	if len(header.Fields) > 0 {
		return decodeTabularArray(header, cursor, baseDepth, options)
	}

	return decodeListArray(header, cursor, baseDepth, options)
}

func decodeListArray(header *ArrayHeaderInfo, cursor *LineCursor, baseDepth int, options DecodeOptions) (JsonArray, error) {
	items := make(JsonArray, 0, header.Length)
	itemDepth := baseDepth + 1

	for !cursor.AtEnd() && len(items) < header.Length {
		line := cursor.Peek()
		if line == nil || line.Depth < itemDepth {
			break
		}

		isListItem := strings.HasPrefix(line.Content, ListItemPrefix) || line.Content == "-"

		if line.Depth == itemDepth && isListItem {
			item, err := decodeListItem(cursor, itemDepth, options)
			if err != nil {
				return nil, err
			}
			items = append(items, item)
		} else {
			break
		}
	}

	return items, nil
}

func decodeListItem(cursor *LineCursor, baseDepth int, options DecodeOptions) (JsonValue, error) {
	line := cursor.Next()
	if line == nil {
		return nil, fmt.Errorf("expected list item")
	}

	var afterHyphen string
	if line.Content == "-" {
		return make(JsonObject), nil
	} else if strings.HasPrefix(line.Content, ListItemPrefix) {
		afterHyphen = strings.TrimPrefix(line.Content, ListItemPrefix)
	} else {
		return nil, fmt.Errorf("expected list item to start with '- '")
	}

	if strings.TrimSpace(afterHyphen) == "" {
		return make(JsonObject), nil
	}

	// Check for object first field after hyphen
	// e.g. "- name: John"
	if strings.Contains(afterHyphen, ":") {
		// Treat as start of an object
		// We need to parse this line as a key-value pair, then continue parsing the object at the same depth
		// This is a bit tricky because decodeObject expects to iterate over lines.
		// We might need a helper `decodeObjectFromListItem`
		return decodeObjectFromListItem(line, cursor, baseDepth, options)
	}

	// Primitive
	return ParsePrimitiveToken(afterHyphen), nil
}

func decodeObjectFromListItem(firstLine *ParsedLine, cursor *LineCursor, baseDepth int, options DecodeOptions) (JsonObject, error) {
	afterHyphen := strings.TrimPrefix(firstLine.Content, ListItemPrefix)
	key, value, err := decodeKeyValue(afterHyphen, cursor, baseDepth, options)
	if err != nil {
		return nil, err
	}

	obj := JsonObject{key: value}

	// Read subsequent fields at the same depth
	for !cursor.AtEnd() {
		line := cursor.Peek()
		if line == nil || line.Depth < baseDepth {
			break
		}

		// Must be same depth and NOT a list item (which would be next item in array)
		if line.Depth == baseDepth && !strings.HasPrefix(line.Content, ListItemPrefix) && line.Content != "-" {
			cursor.Advance()
			k, v, err := decodeKeyValue(line.Content, cursor, baseDepth, options)
			if err != nil {
				return nil, err
			}
			obj[k] = v
		} else {
			break
		}
	}

	return obj, nil
}

func decodeTabularArray(header *ArrayHeaderInfo, cursor *LineCursor, baseDepth int, options DecodeOptions) (JsonArray, error) {
	objects := make(JsonArray, 0, header.Length)
	rowDepth := baseDepth + 1

	for !cursor.AtEnd() && len(objects) < header.Length {
		line := cursor.Peek()
		if line == nil || line.Depth < rowDepth {
			break
		}

		if line.Depth == rowDepth {
			cursor.Advance()
			values := ParseDelimitedValues(line.Content, header.Delimiter)

			if len(values) != len(header.Fields) {
				// In strict mode this should error, for now just warn or best effort
				// return nil, fmt.Errorf("row values count mismatch")
			}

			obj := make(JsonObject)
			for i, field := range header.Fields {
				if i < len(values) {
					obj[field] = ParsePrimitiveToken(values[i])
				}
			}
			objects = append(objects, obj)
		} else {
			break
		}
	}

	return objects, nil
}
