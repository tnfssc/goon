package toon

import (
	"fmt"
	"sort"
	"strconv"
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

// isPrimitive checks if a value is a primitive (not object or array)
func isPrimitive(v interface{}) bool {
	switch v.(type) {
	case map[string]interface{}, []interface{}, JsonObject, JsonArray:
		return false
	default:
		return true
	}
}

// isPrimitiveArray checks if all items in an array are primitives
func isPrimitiveArray(arr []interface{}) bool {
	for _, item := range arr {
		if !isPrimitive(item) {
			return false
		}
	}
	return true
}

// needsQuoting checks if a string value needs to be quoted
func needsQuoting(s string) bool {
	// Empty string needs quoting
	if s == "" {
		return true
	}
	// Leading or trailing whitespace
	if s != strings.TrimSpace(s) {
		return true
	}
	// Reserved literals
	if s == "true" || s == "false" || s == "null" {
		return true
	}
	// Numeric-like
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	// Leading zeros (like "05")
	if len(s) > 1 && s[0] == '0' && s[1] >= '0' && s[1] <= '9' {
		return true
	}
	// Contains colon, double quote, backslash
	if strings.ContainsAny(s, ":\"\\") {
		return true
	}
	// Contains brackets or braces
	if strings.ContainsAny(s, "[]{}") {
		return true
	}
	// Contains control characters
	if strings.ContainsAny(s, "\n\r\t") {
		return true
	}
	// Contains comma (delimiter)
	if strings.Contains(s, ",") {
		return true
	}
	// Starts with hyphen
	if len(s) > 0 && s[0] == '-' {
		return true
	}
	return false
}

// quoteString quotes and escapes a string value
func quoteString(s string) string {
	// Escape special characters
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return "\"" + s + "\""
}

// encodePrimitiveValue encodes a primitive value, optionally quoting strings
func encodePrimitiveValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case int:
		return fmt.Sprintf("%d", val)
	case string:
		if needsQuoting(val) {
			return quoteString(val)
		}
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
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
		if needsQuoting(v) {
			sb.WriteString(quoteString(v))
		} else {
			sb.WriteString(v)
		}
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

		if i > 0 {
			sb.WriteString("\n")
		}

		sb.WriteString(indent)
		sb.WriteString(key)

		// Check if value is complex (object or array) or primitive
		switch val := val.(type) {
		case map[string]interface{}:
			sb.WriteString(":")
			if len(val) > 0 {
				sb.WriteString("\n")
				if err := encodeObject(sb, val, depth+1, options); err != nil {
					return err
				}
			}
		case []interface{}:
			// Array handling - encode header inline with key
			if err := encodeArrayWithKey(sb, val, depth, options); err != nil {
				return err
			}
		default:
			sb.WriteString(": ")
			if err := encodeValue(sb, val, depth, options); err != nil {
				return err
			}
		}
	}
	return nil
}

// encodeArrayWithKey encodes an array as a value for a key (key already written, need to write header and content)
func encodeArrayWithKey(sb *strings.Builder, arr []interface{}, depth int, options EncodeOptions) error {
	// V2 format: key[N]: for comma delimiter (default)
	// V1 format: key: [N|] with list items

	if options.Version == V1 {
		return encodeArrayWithKeyV1(sb, arr, depth, options)
	}

	// V2 format (default)
	if isPrimitiveArray(arr) {
		// Inline format for primitive arrays
		sb.WriteString(fmt.Sprintf("[%d]:", len(arr)))
		if len(arr) > 0 {
			sb.WriteString(" ")
			for i, item := range arr {
				if i > 0 {
					sb.WriteString(",")
				}
				sb.WriteString(encodePrimitiveValue(item))
			}
		}
	} else {
		// List format for arrays with objects or nested arrays
		sb.WriteString(fmt.Sprintf("[%d]:", len(arr)))

		indent := strings.Repeat(" ", (depth+1)*options.IndentSize)

		for _, item := range arr {
			sb.WriteString("\n")
			sb.WriteString(indent)

			if obj, ok := item.(map[string]interface{}); ok {
				// Special handling for object in list
				if len(obj) == 0 {
					// Empty object - just the hyphen, per spec
					sb.WriteString("-")
				} else {
					sb.WriteString("- ")
					keys := make([]string, 0, len(obj))
					for k := range obj {
						keys = append(keys, k)
					}
					sort.Strings(keys)

					firstKey := keys[0]
					firstVal := obj[firstKey]

					// Write first key
					sb.WriteString(firstKey)

					// Check if first val is an array (needs special handling)
					if arrVal, ok := firstVal.([]interface{}); ok {
						if err := encodeArrayWithKey(sb, arrVal, depth+1, options); err != nil {
							return err
						}
					} else if objVal, ok := firstVal.(map[string]interface{}); ok {
						sb.WriteString(":")
						if len(objVal) > 0 {
							sb.WriteString("\n")
							if err := encodeObject(sb, objVal, depth+2, options); err != nil {
								return err
							}
						}
					} else {
						sb.WriteString(": ")
						if err := encodeValue(sb, firstVal, 0, options); err != nil {
							return err
						}
					}

					// Remaining keys at depth+1
					subIndent := strings.Repeat(" ", (depth+1)*options.IndentSize)
					for i := 1; i < len(keys); i++ {
						k := keys[i]
						v := obj[k]
						sb.WriteString("\n")
						sb.WriteString(subIndent)
						sb.WriteString(k)

						// Check if value is array or object
						if arrVal, ok := v.([]interface{}); ok {
							if err := encodeArrayWithKey(sb, arrVal, depth+1, options); err != nil {
								return err
							}
						} else if objVal, ok := v.(map[string]interface{}); ok {
							sb.WriteString(":")
							if len(objVal) > 0 {
								sb.WriteString("\n")
								if err := encodeObject(sb, objVal, depth+2, options); err != nil {
									return err
								}
							}
						} else {
							sb.WriteString(": ")
							if err := encodeValue(sb, v, 0, options); err != nil {
								return err
							}
						}
					}
				}
			} else if nestedArr, ok := item.([]interface{}); ok {
				// Nested array as list item
				sb.WriteString("- ")
				// Format: - [M]: v1,v2,... for primitive arrays
				// or - [M]: with list items for complex arrays
				if isPrimitiveArray(nestedArr) {
					sb.WriteString(fmt.Sprintf("[%d]:", len(nestedArr)))
					if len(nestedArr) > 0 {
						sb.WriteString(" ")
						for i, nestedItem := range nestedArr {
							if i > 0 {
								sb.WriteString(",")
							}
							sb.WriteString(encodePrimitiveValue(nestedItem))
						}
					}
				} else {
					// Complex nested array - use list format
					sb.WriteString(fmt.Sprintf("[%d]:", len(nestedArr)))
					nestedIndent := strings.Repeat(" ", (depth+2)*options.IndentSize)
					for _, nestedItem := range nestedArr {
						sb.WriteString("\n")
						sb.WriteString(nestedIndent)
						sb.WriteString("- ")
						if err := encodeValue(sb, nestedItem, depth+2, options); err != nil {
							return err
						}
					}
				}
			} else {
				// Primitive item
				sb.WriteString("- ")
				if err := encodeValue(sb, item, 0, options); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// encodeArrayWithKeyV1 encodes an array in TOON v1 format (list style with [N|])
func encodeArrayWithKeyV1(sb *strings.Builder, arr []interface{}, depth int, options EncodeOptions) error {
	// V1 format: key: [N|] with list items for all arrays
	sb.WriteString(fmt.Sprintf(": [%d|]", len(arr)))

	indent := strings.Repeat(" ", (depth+1)*options.IndentSize)

	for _, item := range arr {
		sb.WriteString("\n")
		sb.WriteString(indent)

		if obj, ok := item.(map[string]interface{}); ok {
			if len(obj) == 0 {
				// Empty object - just the hyphen, per spec
				sb.WriteString("-")
			} else {
				sb.WriteString("- ")
				keys := make([]string, 0, len(obj))
				for k := range obj {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				firstKey := keys[0]
				firstVal := obj[firstKey]

				sb.WriteString(firstKey)

				if arrVal, ok := firstVal.([]interface{}); ok {
					if err := encodeArrayWithKeyV1(sb, arrVal, depth+1, options); err != nil {
						return err
					}
				} else if objVal, ok := firstVal.(map[string]interface{}); ok {
					sb.WriteString(":")
					if len(objVal) > 0 {
						sb.WriteString("\n")
						if err := encodeObject(sb, objVal, depth+2, options); err != nil {
							return err
						}
					}
				} else {
					sb.WriteString(": ")
					if err := encodeValue(sb, firstVal, 0, options); err != nil {
						return err
					}
				}

				subIndent := strings.Repeat(" ", (depth+1)*options.IndentSize)
				for i := 1; i < len(keys); i++ {
					k := keys[i]
					v := obj[k]
					sb.WriteString("\n")
					sb.WriteString(subIndent)
					sb.WriteString(k)

					if arrVal, ok := v.([]interface{}); ok {
						if err := encodeArrayWithKeyV1(sb, arrVal, depth+1, options); err != nil {
							return err
						}
					} else if objVal, ok := v.(map[string]interface{}); ok {
						sb.WriteString(":")
						if len(objVal) > 0 {
							sb.WriteString("\n")
							if err := encodeObject(sb, objVal, depth+2, options); err != nil {
								return err
							}
						}
					} else {
						sb.WriteString(": ")
						if err := encodeValue(sb, v, 0, options); err != nil {
							return err
						}
					}
				}
			}
		} else if nestedArr, ok := item.([]interface{}); ok {
			// Nested array
			sb.WriteString("- ")
			sb.WriteString(fmt.Sprintf("[%d|]", len(nestedArr)))
			nestedIndent := strings.Repeat(" ", (depth+2)*options.IndentSize)
			for _, nestedItem := range nestedArr {
				sb.WriteString("\n")
				sb.WriteString(nestedIndent)
				sb.WriteString("- ")
				if err := encodeValue(sb, nestedItem, depth+2, options); err != nil {
					return err
				}
			}
		} else {
			// Primitive item
			sb.WriteString("- ")
			if err := encodeValue(sb, item, 0, options); err != nil {
				return err
			}
		}
	}
	return nil
}

func encodeArray(sb *strings.Builder, arr []interface{}, depth int, options EncodeOptions) error {
	// This is called when array is the root value
	if options.Version == V1 {
		return encodeArrayV1(sb, arr, depth, options)
	}

	// V2 format (default)
	if isPrimitiveArray(arr) {
		// Inline format for primitive arrays at root
		sb.WriteString(fmt.Sprintf("[%d]:", len(arr)))
		if len(arr) > 0 {
			sb.WriteString(" ")
			for i, item := range arr {
				if i > 0 {
					sb.WriteString(",")
				}
				sb.WriteString(encodePrimitiveValue(item))
			}
		}
	} else {
		// List format for arrays with objects
		sb.WriteString(fmt.Sprintf("[%d]:", len(arr)))

		indent := strings.Repeat(" ", (depth+1)*options.IndentSize)

		for _, item := range arr {
			sb.WriteString("\n")
			sb.WriteString(indent)

			if obj, ok := item.(map[string]interface{}); ok {
				if len(obj) == 0 {
					// Empty object - just the hyphen, per spec
					sb.WriteString("-")
				} else {
					sb.WriteString("- ")
					keys := make([]string, 0, len(obj))
					for k := range obj {
						keys = append(keys, k)
					}
					sort.Strings(keys)

					firstKey := keys[0]
					firstVal := obj[firstKey]

					sb.WriteString(firstKey)

					if arrVal, ok := firstVal.([]interface{}); ok {
						if err := encodeArrayWithKey(sb, arrVal, depth+1, options); err != nil {
							return err
						}
					} else if objVal, ok := firstVal.(map[string]interface{}); ok {
						sb.WriteString(":")
						if len(objVal) > 0 {
							sb.WriteString("\n")
							if err := encodeObject(sb, objVal, depth+2, options); err != nil {
								return err
							}
						}
					} else {
						sb.WriteString(": ")
						if err := encodeValue(sb, firstVal, 0, options); err != nil {
							return err
						}
					}

					subIndent := strings.Repeat(" ", (depth+1)*options.IndentSize)
					for i := 1; i < len(keys); i++ {
						k := keys[i]
						v := obj[k]
						sb.WriteString("\n")
						sb.WriteString(subIndent)
						sb.WriteString(k)

						if arrVal, ok := v.([]interface{}); ok {
							if err := encodeArrayWithKey(sb, arrVal, depth+1, options); err != nil {
								return err
							}
						} else if objVal, ok := v.(map[string]interface{}); ok {
							sb.WriteString(":")
							if len(objVal) > 0 {
								sb.WriteString("\n")
								if err := encodeObject(sb, objVal, depth+2, options); err != nil {
									return err
								}
							}
						} else {
							sb.WriteString(": ")
							if err := encodeValue(sb, v, 0, options); err != nil {
								return err
							}
						}
					}
				}
			} else if nestedArr, ok := item.([]interface{}); ok {
				sb.WriteString("- ")
				if isPrimitiveArray(nestedArr) {
					sb.WriteString(fmt.Sprintf("[%d]:", len(nestedArr)))
					if len(nestedArr) > 0 {
						sb.WriteString(" ")
						for i, nestedItem := range nestedArr {
							if i > 0 {
								sb.WriteString(",")
							}
							sb.WriteString(encodePrimitiveValue(nestedItem))
						}
					}
				} else {
					sb.WriteString(fmt.Sprintf("[%d]:", len(nestedArr)))
					nestedIndent := strings.Repeat(" ", (depth+2)*options.IndentSize)
					for _, nestedItem := range nestedArr {
						sb.WriteString("\n")
						sb.WriteString(nestedIndent)
						sb.WriteString("- ")
						if err := encodeValue(sb, nestedItem, depth+2, options); err != nil {
							return err
						}
					}
				}
			} else {
				sb.WriteString("- ")
				if err := encodeValue(sb, item, 0, options); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// encodeArrayV1 encodes an array in TOON v1 format (list style)
func encodeArrayV1(sb *strings.Builder, arr []interface{}, depth int, options EncodeOptions) error {
	// V1 format: [N|] with list items
	sb.WriteString(fmt.Sprintf("[%d|]", len(arr)))

	indent := strings.Repeat(" ", (depth+1)*options.IndentSize)

	for _, item := range arr {
		sb.WriteString("\n")
		sb.WriteString(indent)

		if obj, ok := item.(map[string]interface{}); ok {
			if len(obj) == 0 {
				// Empty object - just the hyphen, per spec
				sb.WriteString("-")
			} else {
				sb.WriteString("- ")
				keys := make([]string, 0, len(obj))
				for k := range obj {
					keys = append(keys, k)
				}
				sort.Strings(keys)

				firstKey := keys[0]
				firstVal := obj[firstKey]

				sb.WriteString(firstKey)

				if arrVal, ok := firstVal.([]interface{}); ok {
					if err := encodeArrayWithKeyV1(sb, arrVal, depth+1, options); err != nil {
						return err
					}
				} else if objVal, ok := firstVal.(map[string]interface{}); ok {
					sb.WriteString(":")
					if len(objVal) > 0 {
						sb.WriteString("\n")
						if err := encodeObject(sb, objVal, depth+2, options); err != nil {
							return err
						}
					}
				} else {
					sb.WriteString(": ")
					if err := encodeValue(sb, firstVal, 0, options); err != nil {
						return err
					}
				}

				subIndent := strings.Repeat(" ", (depth+1)*options.IndentSize)
				for i := 1; i < len(keys); i++ {
					k := keys[i]
					v := obj[k]
					sb.WriteString("\n")
					sb.WriteString(subIndent)
					sb.WriteString(k)

					if arrVal, ok := v.([]interface{}); ok {
						if err := encodeArrayWithKeyV1(sb, arrVal, depth+1, options); err != nil {
							return err
						}
					} else if objVal, ok := v.(map[string]interface{}); ok {
						sb.WriteString(":")
						if len(objVal) > 0 {
							sb.WriteString("\n")
							if err := encodeObject(sb, objVal, depth+2, options); err != nil {
								return err
							}
						}
					} else {
						sb.WriteString(": ")
						if err := encodeValue(sb, v, 0, options); err != nil {
							return err
						}
					}
				}
			}
		} else if nestedArr, ok := item.([]interface{}); ok {
			sb.WriteString("- ")
			sb.WriteString(fmt.Sprintf("[%d|]", len(nestedArr)))
			nestedIndent := strings.Repeat(" ", (depth+2)*options.IndentSize)
			for _, nestedItem := range nestedArr {
				sb.WriteString("\n")
				sb.WriteString(nestedIndent)
				sb.WriteString("- ")
				if err := encodeValue(sb, nestedItem, depth+2, options); err != nil {
					return err
				}
			}
		} else {
			sb.WriteString("- ")
			if err := encodeValue(sb, item, 0, options); err != nil {
				return err
			}
		}
	}
	return nil
}
