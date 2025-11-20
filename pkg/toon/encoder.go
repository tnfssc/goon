package toon

import (
	"fmt"
	"sort"
	"strings"
)

// encode converts a JsonValue to TOON string (internal function)
func encode(value JsonValue, options EncodeOptions) (string, error) {
	// Ensure IndentSize has a default value
	if options.IndentSize == 0 {
		options.IndentSize = 2
	}

	var sb strings.Builder
	if err := encodeValue(&sb, value, 0, options); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func encodeValue(sb *strings.Builder, value JsonValue, depth int, options EncodeOptions) error {

	switch v := value.(type) {
	case nil:
		sb.WriteString("null")
	case bool:
		sb.WriteString(fmt.Sprintf("%t", v))
	case float64:
		// Use %g to avoid trailing zeros but keep precision
		sb.WriteString(fmt.Sprintf("%g", v))
	case int:
		sb.WriteString(fmt.Sprintf("%d", v))
	case string:
		// TODO: Handle quoting if needed
		sb.WriteString(v)
	case map[string]interface{}:
		return encodeObject(sb, v, depth, options)
	case []interface{}:
		return encodeArray(sb, v, depth, options)
	default:
		// Try to handle other numeric types or custom types if possible, or error
		return fmt.Errorf("unsupported type: %T", v)
	}
	return nil
}

func encodeObject(sb *strings.Builder, obj map[string]interface{}, depth int, options EncodeOptions) error {
	// Sort keys for deterministic output
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	indent := strings.Repeat(" ", depth*options.IndentSize)

	for i, key := range keys {
		val := obj[key]

		// If it's the first item and we are not at root (depth > 0), we might need newline if called from parent
		// But `encodeValue` is usually called for the value part.
		// Actually, `encodeObject` is called when the value IS an object.

		// If we are at root (depth 0), we just list keys.
		// If we are nested, we need to ensure we are on a new line?
		// The caller handles the key and colon.

		// Wait, `encodeValue` is called for the VALUE.
		// If the value is an object, we need to print its fields on NEW lines with indentation.
		// BUT, if the object is empty, we print `{}`? Or just nothing?
		// TOON doesn't use braces. Empty object is just nothing?

		if i > 0 {
			sb.WriteString("\n")
		}

		sb.WriteString(indent)
		sb.WriteString(key)
		sb.WriteString(": ")

		// Check if value is complex (object or array) or primitive
		switch val := val.(type) {
		case map[string]interface{}:
			sb.WriteString("\n")
			if err := encodeObject(sb, val, depth+1, options); err != nil {
				return err
			}
		case []interface{}:
			// Array handling
			// Print on same line: "key: 3 |"
			if err := encodeArray(sb, val, depth, options); err != nil {
				return err
			}
		default:
			if err := encodeValue(sb, val, depth, options); err != nil {
				return err
			}
		}
	}
	return nil
}

func encodeArray(sb *strings.Builder, arr []interface{}, depth int, options EncodeOptions) error {
	// Use list format for now
	// Header: length |
	// But simple list is just items with "- "

	// We should output the header first if we want to be explicit, or just list items.
	// TOON arrays usually start with a header line if they are not inline.
	// e.g. "items: 3" (if key is items)
	// But here we are encoding the VALUE. The key was already printed by parent.
	// So we just print the array content.

	// If we want to use the header syntax:
	// The parent printed "key: ".
	// We print "3 |" (length and delimiter)
	// Then items.

	// Use list format with bracket syntax: [length|]
	sb.WriteString(fmt.Sprintf("[%d|]", len(arr)))

	indent := strings.Repeat(" ", (depth+1)*options.IndentSize)

	for _, item := range arr {
		sb.WriteString("\n")
		sb.WriteString(indent)
		sb.WriteString("- ")

		// If item is object, we can inline the first field or nest it
		// For simplicity, let's just encode the value
		// If it's an object, `encodeValue` -> `encodeObject` might try to print fields.
		// We need to handle the "object in list item" case specially if we want compact syntax.
		// e.g. "- name: John"

		if obj, ok := item.(map[string]interface{}); ok {
			// Special handling for object in list
			// Pick first key? Or just new line?
			// Let's try to be compact: "- firstKey: firstVal"
			// Then other keys on next lines.

			keys := make([]string, 0, len(obj))
			for k := range obj {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			if len(keys) > 0 {
				firstKey := keys[0]
				sb.WriteString(firstKey)
				sb.WriteString(": ")

				firstVal := obj[firstKey]
				// Encode first value
				// If first value is primitive, it goes on same line.
				// If it's complex, it might need newline.

				// Simplified: just call encodeValue for first val
				// But we need to handle subsequent keys.

				// This is getting complex for a simple encoder.
				// Let's just delegate to encodeValue but we need to handle the indentation of subsequent keys.
				// Actually, `encodeObject` uses `depth`.
				// If we call `encodeObject` here, it will print fields.
				// But we already printed "- ".
				// If we pass `depth+1`, it will indent relative to list item?

				// Let's do this:
				// Print first key-value manually.
				// Then print rest of keys with indentation.

				// Check if first val is primitive
				switch firstVal.(type) {
				case map[string]interface{}, []interface{}:
					sb.WriteString("\n")
					encodeValue(sb, firstVal, depth+2, options)
				default:
					encodeValue(sb, firstVal, 0, options) // 0 depth because it's inline
				}

				// Remaining keys
				subIndent := strings.Repeat(" ", (depth+1)*options.IndentSize)
				for i := 1; i < len(keys); i++ {
					k := keys[i]
					v := obj[k]
					sb.WriteString("\n")
					sb.WriteString(subIndent)
					sb.WriteString(k)
					sb.WriteString(": ")
					// Encode value
					switch v.(type) {
					case map[string]interface{}, []interface{}:
						sb.WriteString("\n")
						encodeValue(sb, v, depth+2, options)
					default:
						encodeValue(sb, v, 0, options)
					}
				}
			}
		} else {
			// Primitive or array
			if err := encodeValue(sb, item, 0, options); err != nil {
				return err
			}
		}
	}
	return nil
}
