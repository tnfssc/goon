package toon

import (
	"reflect"
	"testing"
)

type TestPerson struct {
	Name string `toon:"name"`
	Age  int    `toon:"age"`
}

type TestComplex struct {
	Title    string            `toon:"title"`
	Count    int               `toon:"count"`
	Active   bool              `toon:"active"`
	Tags     []string          `toon:"tags"`
	Metadata map[string]string `toon:"metadata"`
}

type TestNested struct {
	ID     string     `toon:"id"`
	Person TestPerson `toon:"person"`
}

type TestWithIgnore struct {
	Public  string `toon:"public"`
	Ignored string `toon:"-"`
	private string
}

func TestMarshalSimpleStruct(t *testing.T) {
	person := TestPerson{
		Name: "Alice",
		Age:  30,
	}

	data, err := Marshal(person, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	expected := "age: 30\nname: Alice"
	if string(data) != expected {
		t.Errorf("Expected:\n%s\n\nGot:\n%s", expected, string(data))
	}
}

func TestUnmarshalSimpleStruct(t *testing.T) {
	toonData := []byte("name: Bob\nage: 25")

	var person TestPerson
	err := Unmarshal(toonData, &person, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if person.Name != "Bob" {
		t.Errorf("Expected Name 'Bob', got '%s'", person.Name)
	}
	if person.Age != 25 {
		t.Errorf("Expected Age 25, got %d", person.Age)
	}
}

func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	original := TestComplex{
		Title:  "Test Document",
		Count:  42,
		Active: true,
		Tags:   []string{"golang", "toon", "marshal"},
		Metadata: map[string]string{
			"author":  "John",
			"version": "1.5",
		},
	}

	// Marshal
	data, err := Marshal(original, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Unmarshal
	var decoded TestComplex
	err = Unmarshal(data, &decoded, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Compare
	if decoded.Title != original.Title {
		t.Errorf("Title mismatch: expected '%s', got '%s'", original.Title, decoded.Title)
	}
	if decoded.Count != original.Count {
		t.Errorf("Count mismatch: expected %d, got %d", original.Count, decoded.Count)
	}
	if decoded.Active != original.Active {
		t.Errorf("Active mismatch: expected %v, got %v", original.Active, decoded.Active)
	}
	if !reflect.DeepEqual(decoded.Tags, original.Tags) {
		t.Errorf("Tags mismatch: expected %v, got %v", original.Tags, decoded.Tags)
	}
	if !reflect.DeepEqual(decoded.Metadata, original.Metadata) {
		t.Errorf("Metadata mismatch: expected %v, got %v", original.Metadata, decoded.Metadata)
	}
}

func TestMarshalNestedStruct(t *testing.T) {
	nested := TestNested{
		ID: "123",
		Person: TestPerson{
			Name: "Charlie",
			Age:  35,
		},
	}

	data, err := Marshal(nested, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Should contain nested structure
	// Note: TOON v2 requires quoting numeric-like strings, so "123" becomes \"123\"
	dataStr := string(data)
	if !contains(dataStr, "id: \"123\"") {
		t.Errorf("Missing 'id: \"123\"' in output:\n%s", dataStr)
	}
	if !contains(dataStr, "person:") {
		t.Errorf("Missing 'person:' in output:\n%s", dataStr)
	}
}

func TestUnmarshalNestedStruct(t *testing.T) {
	toonData := []byte(`id: 456
person: 
  name: David
  age: 40`)

	var nested TestNested
	err := Unmarshal(toonData, &nested, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if nested.ID != "456" {
		t.Errorf("Expected ID '456', got '%s'", nested.ID)
	}
	if nested.Person.Name != "David" {
		t.Errorf("Expected Person.Name 'David', got '%s'", nested.Person.Name)
	}
	if nested.Person.Age != 40 {
		t.Errorf("Expected Person.Age 40, got %d", nested.Person.Age)
	}
}

func TestMarshalWithIgnoredFields(t *testing.T) {
	obj := TestWithIgnore{
		Public:  "visible",
		Ignored: "should not appear",
		private: "also should not appear",
	}

	data, err := Marshal(obj, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	dataStr := string(data)
	if !contains(dataStr, "public: visible") {
		t.Errorf("Missing 'public: visible' in output:\n%s", dataStr)
	}
	if contains(dataStr, "Ignored") || contains(dataStr, "should not appear") {
		t.Errorf("Ignored field should not be in output:\n%s", dataStr)
	}
	if contains(dataStr, "private") {
		t.Errorf("Private field should not be in output:\n%s", dataStr)
	}
}

func TestMarshalMap(t *testing.T) {
	data := map[string]interface{}{
		"name":  "Test",
		"value": 123,
		"flag":  true,
	}

	bytes, err := Marshal(data, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	str := string(bytes)
	if !contains(str, "name: Test") {
		t.Errorf("Missing 'name: Test' in output:\n%s", str)
	}
	if !contains(str, "value: 123") {
		t.Errorf("Missing 'value: 123' in output:\n%s", str)
	}
	if !contains(str, "flag: true") {
		t.Errorf("Missing 'flag: true' in output:\n%s", str)
	}
}

func TestUnmarshalToMap(t *testing.T) {
	toonData := []byte("name: TestMap\ncount: 5")

	var result map[string]interface{}
	err := Unmarshal(toonData, &result, DecodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result["name"] != "TestMap" {
		t.Errorf("Expected name 'TestMap', got %v", result["name"])
	}
	if result["count"] != float64(5) {
		t.Errorf("Expected count 5, got %v", result["count"])
	}
}

func TestMarshalArray(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5}

	data, err := Marshal(arr, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	// Should produce inline format with TOON v2 header [5]:
	dataStr := string(data)
	if !contains(dataStr, "[5]:") {
		t.Errorf("Missing array header '[5]:' in output:\n%s", dataStr)
	}
}

func TestUnmarshalArray(t *testing.T) {
	// Test TOON v2 inline format
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

func TestUnmarshalArrayLegacyList(t *testing.T) {
	// Test legacy list format for backward compatibility
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

func TestMarshalNil(t *testing.T) {
	data, err := Marshal(nil, EncodeOptions{IndentSize: 2})
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	if string(data) != "null" {
		t.Errorf("Expected 'null', got '%s'", string(data))
	}
}

func TestUnmarshalInvalidTarget(t *testing.T) {
	toonData := []byte("name: Test")

	// Try to unmarshal to non-pointer
	var person TestPerson
	err := Unmarshal(toonData, person, DecodeOptions{IndentSize: 2})
	if err == nil {
		t.Error("Expected error when unmarshaling to non-pointer, got nil")
	}
}

func TestMarshalPrimitiveTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Marshal(tt.input, EncodeOptions{IndentSize: 2})
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}
			if string(data) != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, string(data))
			}
		})
	}
}

func TestUnmarshalPrimitiveTypes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		target   interface{}
		expected interface{}
	}{
		{"string", "hello", new(string), "hello"},
		{"int", "42", new(int), 42},
		{"float", "3.14", new(float64), 3.14},
		{"bool true", "true", new(bool), true},
		{"bool false", "false", new(bool), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Unmarshal([]byte(tt.input), tt.target, DecodeOptions{IndentSize: 2})
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			// Dereference pointer to get actual value
			actualVal := reflect.ValueOf(tt.target).Elem().Interface()
			if !reflect.DeepEqual(actualVal, tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, actualVal)
			}
		})
	}
}

func TestEscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple newline", `"hello\nworld"`, "hello\nworld"},
		{"simple tab", `"hello\tworld"`, "hello\tworld"},
		{"simple quote", `"hello\"world"`, "hello\"world"},
		{"simple backslash", `"hello\\world"`, "hello\\world"},
		{"backslash then n", `"hello\\nworld"`, "hello\\nworld"},
		{"carriage return", `"hello\rworld"`, "hello\rworld"},
		{"multiple escapes", `"a\\b\"c\nd"`, "a\\b\"c\nd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			err := Unmarshal([]byte(tt.input), &result, DecodeOptions{IndentSize: 2})
			if err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
