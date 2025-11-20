package toon

import (
	"fmt"
	"reflect"
	"strings"
)

// Marshal converts a Go value to TOON format (as bytes).
// It uses reflection to convert structs to JsonValue, then encodes to TOON.
// Struct fields can use `toon:"fieldname"` tags to specify field names in TOON output.
// Fields with `toon:"-"` are ignored. Unexported fields are ignored.
func Marshal(v interface{}, options EncodeOptions) ([]byte, error) {
	jsonValue, err := toJsonValue(v)
	if err != nil {
		return nil, err
	}

	toonStr, err := encode(jsonValue, options)
	if err != nil {
		return nil, err
	}

	return []byte(toonStr), nil
}

// Unmarshal parses TOON data and stores the result in the value pointed to by v.
// It decodes TOON to JsonValue, then uses reflection to populate the struct.
// Struct fields can use `toon:"fieldname"` tags to specify field names in TOON data.
// Fields with `toon:"-"` are ignored.
func Unmarshal(data []byte, v interface{}, options DecodeOptions) error {
	jsonValue, err := decode(string(data), options)
	if err != nil {
		return err
	}

	return fromJsonValue(jsonValue, v)
}

// toJsonValue converts a Go value to JsonValue using reflection
func toJsonValue(v interface{}) (JsonValue, error) {
	if v == nil {
		return nil, nil
	}

	val := reflect.ValueOf(v)
	return toJsonValueReflect(val)
}

func toJsonValueReflect(val reflect.Value) (JsonValue, error) {
	// Dereference pointers
	for val.Kind() == reflect.Ptr || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return nil, nil
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.Bool:
		return val.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(val.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(val.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return val.Float(), nil
	case reflect.String:
		return val.String(), nil
	case reflect.Slice, reflect.Array:
		return toJsonArray(val)
	case reflect.Map:
		return toJsonMap(val)
	case reflect.Struct:
		return toJsonStruct(val)
	default:
		return nil, fmt.Errorf("unsupported type: %v", val.Type())
	}
}

func toJsonArray(val reflect.Value) ([]interface{}, error) {
	length := val.Len()
	arr := make([]interface{}, length)

	for i := 0; i < length; i++ {
		item, err := toJsonValueReflect(val.Index(i))
		if err != nil {
			return nil, err
		}
		arr[i] = item
	}

	return arr, nil
}

func toJsonMap(val reflect.Value) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	iter := val.MapRange()

	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		// Key must be string
		if key.Kind() != reflect.String {
			return nil, fmt.Errorf("map key must be string, got %v", key.Type())
		}

		jsonVal, err := toJsonValueReflect(value)
		if err != nil {
			return nil, err
		}
		obj[key.String()] = jsonVal
	}

	return obj, nil
}

func toJsonStruct(val reflect.Value) (map[string]interface{}, error) {
	obj := make(map[string]interface{})
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name from tag or use field name
		fieldName := field.Name
		tag := field.Tag.Get("toon")
		if tag != "" {
			if tag == "-" {
				// Skip this field
				continue
			}
			// Use tag name (handle "name,omitempty" style tags)
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		jsonVal, err := toJsonValueReflect(fieldVal)
		if err != nil {
			return nil, err
		}

		obj[fieldName] = jsonVal
	}

	return obj, nil
}

// fromJsonValue populates v with data from JsonValue using reflection
func fromJsonValue(jsonVal JsonValue, v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("Unmarshal target must be a non-nil pointer")
	}

	return fromJsonValueReflect(jsonVal, val.Elem())
}

func fromJsonValueReflect(jsonVal JsonValue, val reflect.Value) error {
	if jsonVal == nil {
		// Set to zero value
		val.Set(reflect.Zero(val.Type()))
		return nil
	}

	// Handle pointer types
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			val.Set(reflect.New(val.Type().Elem()))
		}
		return fromJsonValueReflect(jsonVal, val.Elem())
	}

	switch val.Kind() {
	case reflect.Bool:
		if b, ok := jsonVal.(bool); ok {
			val.SetBool(b)
			return nil
		}
		return fmt.Errorf("cannot convert %T to bool", jsonVal)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if f, ok := jsonVal.(float64); ok {
			val.SetInt(int64(f))
			return nil
		}
		return fmt.Errorf("cannot convert %T to int", jsonVal)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if f, ok := jsonVal.(float64); ok {
			val.SetUint(uint64(f))
			return nil
		}
		return fmt.Errorf("cannot convert %T to uint", jsonVal)

	case reflect.Float32, reflect.Float64:
		if f, ok := jsonVal.(float64); ok {
			val.SetFloat(f)
			return nil
		}
		return fmt.Errorf("cannot convert %T to float", jsonVal)

	case reflect.String:
		switch v := jsonVal.(type) {
		case string:
			val.SetString(v)
			return nil
		case float64:
			// Handle numeric values that should be strings
			val.SetString(fmt.Sprintf("%g", v))
			return nil
		case int:
			val.SetString(fmt.Sprintf("%d", v))
			return nil
		case bool:
			val.SetString(fmt.Sprintf("%t", v))
			return nil
		default:
			return fmt.Errorf("cannot convert %T to string", jsonVal)
		}

	case reflect.Slice:
		return fromJsonArrayToSlice(jsonVal, val)

	case reflect.Array:
		return fromJsonArrayToArray(jsonVal, val)

	case reflect.Map:
		return fromJsonObjectToMap(jsonVal, val)

	case reflect.Struct:
		return fromJsonObjectToStruct(jsonVal, val)

	case reflect.Interface:
		// For interface{}, just set the value directly
		if val.Type().NumMethod() == 0 {
			val.Set(reflect.ValueOf(jsonVal))
			return nil
		}
		return fmt.Errorf("cannot unmarshal into non-empty interface")

	default:
		return fmt.Errorf("unsupported type: %v", val.Type())
	}
}

func fromJsonArrayToSlice(jsonVal JsonValue, val reflect.Value) error {
	var arr []interface{}
	switch v := jsonVal.(type) {
	case []interface{}:
		arr = v
	case JsonArray:
		// JsonArray is []JsonValue which is []interface{}, so iterate and copy
		arr = make([]interface{}, len(v))
		for i, item := range v {
			arr[i] = item
		}
	default:
		return fmt.Errorf("expected array, got %T", jsonVal)
	}

	slice := reflect.MakeSlice(val.Type(), len(arr), len(arr))
	for i, item := range arr {
		if err := fromJsonValueReflect(item, slice.Index(i)); err != nil {
			return err
		}
	}

	val.Set(slice)
	return nil
}

func fromJsonArrayToArray(jsonVal JsonValue, val reflect.Value) error {
	var arr []interface{}
	switch v := jsonVal.(type) {
	case []interface{}:
		arr = v
	case JsonArray:
		// JsonArray is []JsonValue which is []interface{}, so iterate and copy
		arr = make([]interface{}, len(v))
		for i, item := range v {
			arr[i] = item
		}
	default:
		return fmt.Errorf("expected array, got %T", jsonVal)
	}

	if len(arr) != val.Len() {
		return fmt.Errorf("array length mismatch: expected %d, got %d", val.Len(), len(arr))
	}

	for i, item := range arr {
		if err := fromJsonValueReflect(item, val.Index(i)); err != nil {
			return err
		}
	}

	return nil
}

func fromJsonObjectToMap(jsonVal JsonValue, val reflect.Value) error {
	var obj map[string]interface{}
	switch v := jsonVal.(type) {
	case map[string]interface{}:
		obj = v
	case JsonObject:
		// JsonObject is map[string]JsonValue which is map[string]interface{}, so copy
		obj = make(map[string]interface{})
		for k, v := range v {
			obj[k] = v
		}
	default:
		return fmt.Errorf("expected object, got %T", jsonVal)
	}

	if val.IsNil() {
		val.Set(reflect.MakeMap(val.Type()))
	}

	keyType := val.Type().Key()
	if keyType.Kind() != reflect.String {
		return fmt.Errorf("map key must be string")
	}

	valueType := val.Type().Elem()
	for k, v := range obj {
		keyVal := reflect.ValueOf(k)
		valueVal := reflect.New(valueType).Elem()

		if err := fromJsonValueReflect(v, valueVal); err != nil {
			return err
		}

		val.SetMapIndex(keyVal, valueVal)
	}

	return nil
}

func fromJsonObjectToStruct(jsonVal JsonValue, val reflect.Value) error {
	var obj map[string]interface{}
	switch v := jsonVal.(type) {
	case map[string]interface{}:
		obj = v
	case JsonObject:
		// JsonObject is map[string]JsonValue which is map[string]interface{}, so copy
		obj = make(map[string]interface{})
		for k, v := range v {
			obj[k] = v
		}
	default:
		return fmt.Errorf("expected object, got %T", jsonVal)
	}

	typ := val.Type()

	// Build a map of toon field names to struct field indices
	fieldMap := make(map[string]int)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldName := field.Name
		tag := field.Tag.Get("toon")
		if tag != "" {
			if tag == "-" {
				continue
			}
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				fieldName = parts[0]
			}
		}

		fieldMap[fieldName] = i
	}

	// Populate struct fields from object
	for key, value := range obj {
		if fieldIdx, ok := fieldMap[key]; ok {
			fieldVal := val.Field(fieldIdx)
			if err := fromJsonValueReflect(value, fieldVal); err != nil {
				return fmt.Errorf("error setting field %s: %w", typ.Field(fieldIdx).Name, err)
			}
		}
		// If field not found, ignore (like json.Unmarshal does)
	}

	return nil
}
