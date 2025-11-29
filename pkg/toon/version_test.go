package toon

import (
	"reflect"
	"strings"
	"testing"
)

// =============================================================================
// V2 FORMAT TESTS (Default)
// =============================================================================

func TestV2PrimitiveArrayInline(t *testing.T) {
	// V2 encodes primitive arrays inline
	data := map[string]interface{}{
		"tags": []interface{}{"a", "b", "c"},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "tags[3]: a,b,c"
	if string(result) != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, string(result))
	}
}

func TestV2PrimitiveArrayInlineNumbers(t *testing.T) {
	data := map[string]interface{}{
		"nums": []interface{}{float64(1), float64(2), float64(3)},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "nums[3]: 1,2,3"
	if string(result) != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, string(result))
	}
}

func TestV2EmptyArray(t *testing.T) {
	data := map[string]interface{}{
		"empty": []interface{}{},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "empty[0]:"
	if string(result) != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, string(result))
	}
}

func TestV2MixedPrimitiveArray(t *testing.T) {
	data := map[string]interface{}{
		"mixed": []interface{}{"a", float64(1), true, false},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "mixed[4]: a,1,true,false"
	if string(result) != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, string(result))
	}
}

func TestV2ObjectArray(t *testing.T) {
	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"id": float64(1), "name": "Alice"},
			map[string]interface{}{"id": float64(2), "name": "Bob"},
		},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// V2 uses [N]: format
	if !strings.Contains(string(result), "users[2]:") {
		t.Errorf("Expected 'users[2]:' in output:\n%s", string(result))
	}
	if !strings.Contains(string(result), "- id:") {
		t.Errorf("Expected list items in output:\n%s", string(result))
	}
}

func TestV2RootPrimitiveArray(t *testing.T) {
	arr := []int{10, 20, 30}

	result, err := Marshal(arr, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "[3]: 10,20,30"
	if string(result) != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, string(result))
	}
}

func TestV2QuotedStrings(t *testing.T) {
	data := map[string]interface{}{
		"nums": []interface{}{"123", "true", "null"},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Numeric-like strings should be quoted
	if !strings.Contains(string(result), `"123"`) {
		t.Errorf("Expected quoted '123' in output:\n%s", string(result))
	}
	if !strings.Contains(string(result), `"true"`) {
		t.Errorf("Expected quoted 'true' in output:\n%s", string(result))
	}
}

func TestV2NestedArrays(t *testing.T) {
	data := map[string]interface{}{
		"matrix": []interface{}{
			[]interface{}{float64(1), float64(2)},
			[]interface{}{float64(3), float64(4)},
		},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// V2 should handle nested arrays
	if !strings.Contains(string(result), "matrix[2]:") {
		t.Errorf("Expected 'matrix[2]:' in output:\n%s", string(result))
	}
}

// =============================================================================
// V1 FORMAT TESTS
// =============================================================================

func TestV1PrimitiveArrayList(t *testing.T) {
	// V1 encodes all arrays as lists with [N|]
	data := map[string]interface{}{
		"tags": []interface{}{"a", "b", "c"},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2, Version: V1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// V1 format uses [N|] with list items
	if !strings.Contains(string(result), "[3|]") {
		t.Errorf("Expected '[3|]' in output:\n%s", string(result))
	}
	if !strings.Contains(string(result), "- a") {
		t.Errorf("Expected '- a' in output:\n%s", string(result))
	}
	if !strings.Contains(string(result), "- b") {
		t.Errorf("Expected '- b' in output:\n%s", string(result))
	}
	if !strings.Contains(string(result), "- c") {
		t.Errorf("Expected '- c' in output:\n%s", string(result))
	}
}

func TestV1PrimitiveArrayListNumbers(t *testing.T) {
	data := map[string]interface{}{
		"nums": []interface{}{float64(1), float64(2), float64(3)},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2, Version: V1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if !strings.Contains(string(result), "[3|]") {
		t.Errorf("Expected '[3|]' in output:\n%s", string(result))
	}
	if !strings.Contains(string(result), "- 1") {
		t.Errorf("Expected '- 1' in output:\n%s", string(result))
	}
}

func TestV1EmptyArray(t *testing.T) {
	data := map[string]interface{}{
		"empty": []interface{}{},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2, Version: V1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if !strings.Contains(string(result), "[0|]") {
		t.Errorf("Expected '[0|]' in output:\n%s", string(result))
	}
}

func TestV1ObjectArray(t *testing.T) {
	data := map[string]interface{}{
		"users": []interface{}{
			map[string]interface{}{"id": float64(1), "name": "Alice"},
			map[string]interface{}{"id": float64(2), "name": "Bob"},
		},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2, Version: V1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// V1 uses [N|] format
	if !strings.Contains(string(result), "[2|]") {
		t.Errorf("Expected '[2|]' in output:\n%s", string(result))
	}
	if !strings.Contains(string(result), "- id:") {
		t.Errorf("Expected list items in output:\n%s", string(result))
	}
}

func TestV1RootPrimitiveArray(t *testing.T) {
	arr := []int{10, 20, 30}

	result, err := Marshal(arr, EncodeOptions{IndentSize: 2, Version: V1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if !strings.Contains(string(result), "[3|]") {
		t.Errorf("Expected '[3|]' in output:\n%s", string(result))
	}
	if !strings.Contains(string(result), "- 10") {
		t.Errorf("Expected '- 10' in output:\n%s", string(result))
	}
}

func TestV1NestedArrays(t *testing.T) {
	data := map[string]interface{}{
		"matrix": []interface{}{
			[]interface{}{float64(1), float64(2)},
			[]interface{}{float64(3), float64(4)},
		},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2, Version: V1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// V1 should use [N|] for nested arrays too
	if !strings.Contains(string(result), "[2|]") {
		t.Errorf("Expected '[2|]' in output:\n%s", string(result))
	}
}

// =============================================================================
// DECODER TESTS - V1 FORMAT BACKWARD COMPATIBILITY
// =============================================================================

func TestDecodeV1PrimitiveArrayList(t *testing.T) {
	toonData := []byte(`tags: [3|]
  - a
  - b
  - c`)

	var result map[string]interface{}
	err := Unmarshal(toonData, &result, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Handle both []interface{} and JsonArray types
	tags := result["tags"]
	var tagSlice []interface{}
	switch v := tags.(type) {
	case []interface{}:
		tagSlice = v
	case JsonArray:
		tagSlice = make([]interface{}, len(v))
		for i, item := range v {
			tagSlice[i] = item
		}
	default:
		t.Fatalf("Expected array type, got %T", tags)
	}

	expected := []interface{}{"a", "b", "c"}
	if !reflect.DeepEqual(tagSlice, expected) {
		t.Errorf("Expected %v, got %v", expected, tagSlice)
	}
}

func TestDecodeV1NumberArrayList(t *testing.T) {
	toonData := []byte(`nums: [3|]
  - 1
  - 2
  - 3`)

	var result map[string]interface{}
	err := Unmarshal(toonData, &result, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	nums := result["nums"]
	var numSlice []interface{}
	switch v := nums.(type) {
	case []interface{}:
		numSlice = v
	case JsonArray:
		numSlice = make([]interface{}, len(v))
		for i, item := range v {
			numSlice[i] = item
		}
	default:
		t.Fatalf("Expected array type, got %T", nums)
	}

	expected := []interface{}{float64(1), float64(2), float64(3)}
	if !reflect.DeepEqual(numSlice, expected) {
		t.Errorf("Expected %v, got %v", expected, numSlice)
	}
}

func TestDecodeV1ObjectArray(t *testing.T) {
	toonData := []byte(`users: [2|]
  - id: 1
    name: Alice
  - id: 2
    name: Bob`)

	var result map[string]interface{}
	err := Unmarshal(toonData, &result, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	users := result["users"]
	var usersSlice []interface{}
	switch v := users.(type) {
	case []interface{}:
		usersSlice = v
	case JsonArray:
		usersSlice = make([]interface{}, len(v))
		for i, item := range v {
			usersSlice[i] = item
		}
	default:
		t.Fatalf("Expected array type, got %T", users)
	}

	if len(usersSlice) != 2 {
		t.Errorf("Expected 2 users, got %d", len(usersSlice))
	}
}

func TestDecodeV1RootArray(t *testing.T) {
	toonData := []byte(`[3|]
  - 10
  - 20
  - 30`)

	var result []int
	err := Unmarshal(toonData, &result, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	expected := []int{10, 20, 30}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// =============================================================================
// DECODER TESTS - V2 FORMAT
// =============================================================================

func TestDecodeV2PrimitiveArrayInline(t *testing.T) {
	toonData := []byte(`tags[3]: a,b,c`)

	var result map[string]interface{}
	err := Unmarshal(toonData, &result, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	tags := result["tags"]
	var tagSlice []interface{}
	switch v := tags.(type) {
	case []interface{}:
		tagSlice = v
	case JsonArray:
		tagSlice = make([]interface{}, len(v))
		for i, item := range v {
			tagSlice[i] = item
		}
	default:
		t.Fatalf("Expected array type, got %T", tags)
	}

	expected := []interface{}{"a", "b", "c"}
	if !reflect.DeepEqual(tagSlice, expected) {
		t.Errorf("Expected %v, got %v", expected, tagSlice)
	}
}

func TestDecodeV2NumberArrayInline(t *testing.T) {
	toonData := []byte(`nums[3]: 1,2,3`)

	var result map[string]interface{}
	err := Unmarshal(toonData, &result, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	nums := result["nums"]
	var numSlice []interface{}
	switch v := nums.(type) {
	case []interface{}:
		numSlice = v
	case JsonArray:
		numSlice = make([]interface{}, len(v))
		for i, item := range v {
			numSlice[i] = item
		}
	default:
		t.Fatalf("Expected array type, got %T", nums)
	}

	expected := []interface{}{float64(1), float64(2), float64(3)}
	if !reflect.DeepEqual(numSlice, expected) {
		t.Errorf("Expected %v, got %v", expected, numSlice)
	}
}

func TestDecodeV2EmptyArray(t *testing.T) {
	toonData := []byte(`empty[0]:`)

	var result map[string]interface{}
	err := Unmarshal(toonData, &result, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	empty := result["empty"]
	var emptySlice []interface{}
	switch v := empty.(type) {
	case []interface{}:
		emptySlice = v
	case JsonArray:
		emptySlice = make([]interface{}, len(v))
		for i, item := range v {
			emptySlice[i] = item
		}
	default:
		t.Fatalf("Expected array type, got %T", empty)
	}

	if len(emptySlice) != 0 {
		t.Errorf("Expected empty array, got %v", emptySlice)
	}
}

func TestDecodeV2ObjectArray(t *testing.T) {
	toonData := []byte(`users[2]:
  - id: 1
    name: Alice
  - id: 2
    name: Bob`)

	var result map[string]interface{}
	err := Unmarshal(toonData, &result, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	users := result["users"]
	var usersSlice []interface{}
	switch v := users.(type) {
	case []interface{}:
		usersSlice = v
	case JsonArray:
		usersSlice = make([]interface{}, len(v))
		for i, item := range v {
			usersSlice[i] = item
		}
	default:
		t.Fatalf("Expected array type, got %T", users)
	}

	if len(usersSlice) != 2 {
		t.Errorf("Expected 2 users, got %d", len(usersSlice))
	}
}

func TestDecodeV2RootArrayInline(t *testing.T) {
	toonData := []byte(`[3]: 10,20,30`)

	var result []int
	err := Unmarshal(toonData, &result, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	expected := []int{10, 20, 30}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Expected %v, got %v", expected, result)
	}
}

// =============================================================================
// ROUND-TRIP TESTS
// =============================================================================

func TestV2RoundTripPrimitiveArray(t *testing.T) {
	original := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}

	// Marshal with V2
	data, err := Marshal(original, EncodeOptions{IndentSize: 2, Version: V2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded map[string]interface{}
	err = Unmarshal(data, &decoded, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	items := decoded["items"]
	var itemsSlice []interface{}
	switch v := items.(type) {
	case []interface{}:
		itemsSlice = v
	case JsonArray:
		itemsSlice = make([]interface{}, len(v))
		for i, item := range v {
			itemsSlice[i] = item
		}
	default:
		t.Fatalf("Expected array type, got %T", items)
	}

	if !reflect.DeepEqual(itemsSlice, original["items"]) {
		t.Errorf("Round-trip failed: expected %v, got %v", original["items"], itemsSlice)
	}
}

func TestV1RoundTripPrimitiveArray(t *testing.T) {
	original := map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	}

	// Marshal with V1
	data, err := Marshal(original, EncodeOptions{IndentSize: 2, Version: V1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded map[string]interface{}
	err = Unmarshal(data, &decoded, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	items := decoded["items"]
	var itemsSlice []interface{}
	switch v := items.(type) {
	case []interface{}:
		itemsSlice = v
	case JsonArray:
		itemsSlice = make([]interface{}, len(v))
		for i, item := range v {
			itemsSlice[i] = item
		}
	default:
		t.Fatalf("Expected array type, got %T", items)
	}

	if !reflect.DeepEqual(itemsSlice, original["items"]) {
		t.Errorf("Round-trip failed: expected %v, got %v", original["items"], itemsSlice)
	}
}

func TestV2RoundTripComplexStructure(t *testing.T) {
	original := map[string]interface{}{
		"name": "test",
		"tags": []interface{}{"a", "b"},
		"data": map[string]interface{}{
			"nums":  []interface{}{float64(1), float64(2), float64(3)},
			"flag":  true,
			"value": float64(42),
		},
	}

	// Marshal with V2
	data, err := Marshal(original, EncodeOptions{IndentSize: 2, Version: V2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded map[string]interface{}
	err = Unmarshal(data, &decoded, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded["name"] != original["name"] {
		t.Errorf("name mismatch: expected %v, got %v", original["name"], decoded["name"])
	}

	tags := decoded["tags"]
	var tagsSlice []interface{}
	switch v := tags.(type) {
	case []interface{}:
		tagsSlice = v
	case JsonArray:
		tagsSlice = make([]interface{}, len(v))
		for i, item := range v {
			tagsSlice[i] = item
		}
	default:
		t.Fatalf("Expected array type for tags, got %T", tags)
	}
	if !reflect.DeepEqual(tagsSlice, original["tags"]) {
		t.Errorf("tags mismatch: expected %v, got %v", original["tags"], tagsSlice)
	}
}

func TestV1RoundTripComplexStructure(t *testing.T) {
	original := map[string]interface{}{
		"name": "test",
		"tags": []interface{}{"a", "b"},
		"data": map[string]interface{}{
			"nums":  []interface{}{float64(1), float64(2), float64(3)},
			"flag":  true,
			"value": float64(42),
		},
	}

	// Marshal with V1
	data, err := Marshal(original, EncodeOptions{IndentSize: 2, Version: V1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded map[string]interface{}
	err = Unmarshal(data, &decoded, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded["name"] != original["name"] {
		t.Errorf("name mismatch: expected %v, got %v", original["name"], decoded["name"])
	}

	tags := decoded["tags"]
	var tagsSlice []interface{}
	switch v := tags.(type) {
	case []interface{}:
		tagsSlice = v
	case JsonArray:
		tagsSlice = make([]interface{}, len(v))
		for i, item := range v {
			tagsSlice[i] = item
		}
	default:
		t.Fatalf("Expected array type for tags, got %T", tags)
	}
	if !reflect.DeepEqual(tagsSlice, original["tags"]) {
		t.Errorf("tags mismatch: expected %v, got %v", original["tags"], tagsSlice)
	}
}

// =============================================================================
// VERSION CONSTANT TESTS
// =============================================================================

func TestVersionConstants(t *testing.T) {
	// V2 should be default (0)
	if V2 != 0 {
		t.Errorf("Expected V2 to be 0 (default), got %d", V2)
	}
	// V1 should be 1
	if V1 != 1 {
		t.Errorf("Expected V1 to be 1, got %d", V1)
	}
}

func TestDefaultVersionIsV2(t *testing.T) {
	// An empty EncodeOptions should default to V2
	opts := EncodeOptions{IndentSize: 2}
	if opts.Version != V2 {
		t.Errorf("Expected default Version to be V2, got %d", opts.Version)
	}
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

func TestV2ArrayWithSpecialCharacters(t *testing.T) {
	data := map[string]interface{}{
		"special": []interface{}{"hello, world", "a:b", "test\"quote"},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Strings with commas, colons, quotes should be quoted
	if !strings.Contains(string(result), `"hello, world"`) {
		t.Errorf("Expected quoted 'hello, world' in output:\n%s", string(result))
	}
}

func TestV1ArrayWithSpecialCharacters(t *testing.T) {
	data := map[string]interface{}{
		"special": []interface{}{"hello, world", "a:b"},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2, Version: V1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// V1 list format should still quote special characters
	if !strings.Contains(string(result), `"hello, world"`) {
		t.Errorf("Expected quoted 'hello, world' in output:\n%s", string(result))
	}
}

func TestV2SingleElementArray(t *testing.T) {
	data := map[string]interface{}{
		"single": []interface{}{"only"},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "single[1]: only"
	if string(result) != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, string(result))
	}
}

func TestV1SingleElementArray(t *testing.T) {
	data := map[string]interface{}{
		"single": []interface{}{"only"},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2, Version: V1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if !strings.Contains(string(result), "[1|]") {
		t.Errorf("Expected '[1|]' in output:\n%s", string(result))
	}
	if !strings.Contains(string(result), "- only") {
		t.Errorf("Expected '- only' in output:\n%s", string(result))
	}
}

func TestV2BooleanArray(t *testing.T) {
	data := map[string]interface{}{
		"flags": []interface{}{true, false, true},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "flags[3]: true,false,true"
	if string(result) != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, string(result))
	}
}

func TestV2NullArray(t *testing.T) {
	data := map[string]interface{}{
		"nulls": []interface{}{nil, nil},
	}

	result, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "nulls[2]: null,null"
	if string(result) != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, string(result))
	}
}

// =============================================================================
// STRUCT ROUND-TRIP TESTS
// =============================================================================

type TestData struct {
	Name  string   `toon:"name"`
	Items []string `toon:"items"`
	Count int      `toon:"count"`
}

func TestV2StructRoundTrip(t *testing.T) {
	original := TestData{
		Name:  "test",
		Items: []string{"a", "b", "c"},
		Count: 42,
	}

	// Marshal with V2
	data, err := Marshal(original, EncodeOptions{IndentSize: 2, Version: V2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Should use V2 format
	if !strings.Contains(string(data), "items[3]:") {
		t.Errorf("Expected V2 format 'items[3]:' in output:\n%s", string(data))
	}

	// Unmarshal
	var decoded TestData
	err = Unmarshal(data, &decoded, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: expected %v, got %v", original.Name, decoded.Name)
	}
	if !reflect.DeepEqual(decoded.Items, original.Items) {
		t.Errorf("Items mismatch: expected %v, got %v", original.Items, decoded.Items)
	}
	if decoded.Count != original.Count {
		t.Errorf("Count mismatch: expected %v, got %v", original.Count, decoded.Count)
	}
}

func TestV1StructRoundTrip(t *testing.T) {
	original := TestData{
		Name:  "test",
		Items: []string{"a", "b", "c"},
		Count: 42,
	}

	// Marshal with V1
	data, err := Marshal(original, EncodeOptions{IndentSize: 2, Version: V1})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Should use V1 format
	if !strings.Contains(string(data), "[3|]") {
		t.Errorf("Expected V1 format '[3|]' in output:\n%s", string(data))
	}

	// Unmarshal
	var decoded TestData
	err = Unmarshal(data, &decoded, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: expected %v, got %v", original.Name, decoded.Name)
	}
	if !reflect.DeepEqual(decoded.Items, original.Items) {
		t.Errorf("Items mismatch: expected %v, got %v", original.Items, decoded.Items)
	}
	if decoded.Count != original.Count {
		t.Errorf("Count mismatch: expected %v, got %v", original.Count, decoded.Count)
	}
}
