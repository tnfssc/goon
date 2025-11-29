package toon

import (
	"fmt"
	"strings"
)

// decode parses TOON content into a JsonValue (internal function)
func decode(source string, options DecodeOptions) (JsonValue, error) {
	// Ensure IndentSize has a default value to prevent divide-by-zero
	if options.IndentSize == 0 {
		options.IndentSize = 2
	}

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

	// Check for root array with TOON v2 format: [N]: values
	if strings.HasPrefix(strings.TrimSpace(first.Content), "[") {
		// Try TOON v2 inline array format
		headerInfo := ParseArrayHeaderLineTOONv2(first.Content)
		if headerInfo != nil {
			cursor.Advance()

			// Get inline values after colon
			colonIdx := strings.Index(first.Content, ":")
			if colonIdx != -1 {
				inlineValues := strings.TrimSpace(first.Content[colonIdx+1:])
				if inlineValues != "" {
					values := ParseDelimitedValues(inlineValues, headerInfo.Delimiter)
					arr := make(JsonArray, len(values))
					for i, v := range values {
						arr[i] = ParsePrimitiveToken(v)
					}
					return arr, nil
				}
			}

			// No inline values - check for list format
			if len(headerInfo.Fields) > 0 {
				return decodeTabularArray(headerInfo, cursor, 0, options)
			}
			return decodeListArray(headerInfo, cursor, 0, options)
		}

		// Try old format with [N|]
		headerInfo = ParseArrayHeaderLine(first.Content, DelimiterComma)
		if headerInfo != nil {
			cursor.Advance()
			return decodeArrayFromHeader(headerInfo, "", cursor, 0, options)
		}

		// Try inline array like [item1, item2]
		if arr, err := ParseInlineArray(first.Content); err == nil {
			cursor.Advance()
			return arr, nil
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
	// Check if content contains an array header pattern like "key[N]:" or "key[N]{fields}:"
	// TOON v2 format: key[N]: values or key[N]{fields}: for tabular

	// Look for bracket pattern to extract key and array info
	bracketStart := strings.Index(content, "[")
	colonIdx := strings.Index(content, ":")

	if bracketStart != -1 && colonIdx != -1 && bracketStart < colonIdx {
		// This might be an array header like "key[N]:" or "key[N]{fields}:"
		key := strings.TrimSpace(content[:bracketStart])
		afterKey := content[bracketStart:]

		// Parse the array header
		headerInfo := ParseArrayHeaderLineTOONv2(afterKey)
		if headerInfo != nil {
			headerInfo.Key = key

			// Get inline values after the colon if any
			colonInAfterKey := strings.Index(afterKey, ":")
			if colonInAfterKey != -1 {
				inlineValues := strings.TrimSpace(afterKey[colonInAfterKey+1:])
				if inlineValues != "" {
					// Parse inline values using delimiter
					values := ParseDelimitedValues(inlineValues, headerInfo.Delimiter)
					arr := make(JsonArray, len(values))
					for i, v := range values {
						arr[i] = ParsePrimitiveToken(v)
					}
					return key, arr, nil
				}
			}

			// No inline values - check for list or tabular format
			if len(headerInfo.Fields) > 0 {
				arr, err := decodeTabularArray(headerInfo, cursor, baseDepth, options)
				return key, arr, err
			}
			arr, err := decodeListArray(headerInfo, cursor, baseDepth, options)
			return key, arr, err
		}
	}

	// Simple key parsing (split by first colon)
	parts := strings.SplitN(content, ":", 2)
	key := strings.TrimSpace(parts[0])

	if len(parts) < 2 {
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

	// Check for array header with old format (for backward compatibility)
	if IsArrayHeaderAfterHyphen(rest) {
		headerInfo := ParseArrayHeaderLine(rest, DelimiterComma)
		if headerInfo != nil {
			val, err := decodeArrayFromHeader(headerInfo, "", cursor, baseDepth, options)
			return key, val, err
		} else {
			// Try inline array with brackets like [item1, item2]
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

	// Sibling fields are at depth baseDepth + 1 (one level deeper than the list item line)
	// because they align with the content after "- "
	siblingDepth := baseDepth + 1

	for !cursor.AtEnd() {
		line := cursor.Peek()
		if line == nil || line.Depth < baseDepth {
			break
		}

		// If we see a line at list item depth that is a list item, we're done with this object
		if line.Depth == baseDepth && (strings.HasPrefix(line.Content, ListItemPrefix) || line.Content == "-") {
			break
		}

		// Sibling fields should be at siblingDepth
		if line.Depth == siblingDepth && !strings.HasPrefix(line.Content, ListItemPrefix) && line.Content != "-" {
			cursor.Advance()
			k, v, err := decodeKeyValue(line.Content, cursor, siblingDepth, options)
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
