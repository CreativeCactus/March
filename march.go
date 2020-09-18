package march

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// FlagRemain denotes the field to which remaining JSON fields can be unmarshaled
const FlagRemain = "remains"

// FlagHoist denotes a type whose values are hoisted to the parent struct when marshaling to JSON
const FlagHoist = "hoist"

// March is the top level interface for Un/Marshaling
type March struct {
	// TODO construct and make .tag private to avoid confusion with defaults
	Tag                string                            // The tag key to look up on structs
	Suffix             string                            // An optional override for custom functions eg. MarshalSUFFIX. Defaults to Tag
	NoMarshalJSON      bool                              // Prevents the default MarshalAsJSON method from trying to use MarshalJSON
	NoUnmarshalJSON    bool                              // Prevents the default UnmarshalAsJSON method from trying to use UnmarshalJSON
	Verbose            bool                              // Used in some cases to show field un/marshaling errors
	Strict             bool                              // Determines whether a failure to un/marshal a field results in a failure overall
	DefaultMarshaler   func(interface{}) ([]byte, error) // Override the default marshaler for types with no custom marshal function
	DefaultUnmarshaler func([]byte, interface{}) error   // Override the default unmarshaler for types with no custom unmarshal function
	Extensions         map[string]Extension              // Functions which MAY be called by encoders if the flag string is present on the field
}

// RawUnmarshal is a wrapper around json.RawMessage which
// provides some helpers for unmarshaling at runtime.
type RawUnmarshal struct {
	Bytes json.RawMessage
	March March
}

// MarshalFrom allows updating the underlying data via the marshaler
func (ru RawUnmarshal) MarshalFrom(v interface{}) error {
	data, err := ru.March.Marshal(v)
	if err != nil {
		return err
	}
	ru.Bytes = json.RawMessage(data)
	return nil
}

// UnmarshalTo allows convenient unmarshaling to a given type
func (ru RawUnmarshal) UnmarshalTo(v interface{}) error {
	return ru.March.Unmarshal(ru.Bytes, v)
}

// MarshalJSON ensures consistent behavior with json.RawMessage
func (ru RawUnmarshal) MarshalJSON() ([]byte, error) {
	return ru.Bytes.MarshalJSON()
}

// UnmarshalJSON ensures consistent behavior with json.RawMessage
func (ru RawUnmarshal) UnmarshalJSON(data []byte) error {
	return ru.Bytes.UnmarshalJSON(data)
}

// JSONType represents a JSON value's type
type JSONType string

const (
	// JSONObject is a JSON object-like type
	JSONObject JSONType = "object"
	// JSONArray is a JSON array-like type
	JSONArray JSONType = "array"
	// JSONNumber is a JSON numeric type
	JSONNumber JSONType = "number"
	// JSONString is a JSON string type
	JSONString JSONType = "string"
	// JSONNull is a JSON null type
	JSONNull JSONType = "null"
	// JSONInvalid is not a valid JSON type
	JSONInvalid JSONType = "invalid"
)

// TypeOfJSON determines the JSON message type of the underlying value
// If the underlying value is malformed JSON, it will return JSONInvalid
func (ru RawUnmarshal) TypeOfJSON() (t JSONType) {
	var v interface{}
	if err := json.Unmarshal(ru.Bytes, &v); err != nil {
		return JSONInvalid
	}

	switch v.(type) {
	case map[string]interface{}:
		return JSONObject
	case []interface{}:
		return JSONArray
	case float64:
		return JSONNumber
	case string:
		return JSONString
	default:
		return JSONInvalid
	}
}

// Marshal provides convenient defaults for March{}.Marshal
// Given any type, returns the representative []byte according to
// marshaler precedence.
// Refer to type March for details.
func Marshal(v interface{}) (data []byte, err error) {
	return March{}.Marshal(v)
}

// Unmarshal provides convenient defaults for March{}.Unmarshal
// Given []byte and any type, updates the pointed type according to
// unmarshaler precedence.
// Refer to type March for details.
func Unmarshal(data []byte, v interface{}) (err error) {
	return March{}.Unmarshal(data, v)
}

// Marshal takes any type with tags at the given tag key (determined
// by the value of M.TagKey()) and returns a recursively marshaled []byte,
// by default in JSON, or by a custom marshal method if one exists on
// the given type.
func (M March) Marshal(v interface{}) (data []byte, err error) {
	{ // Sanity check
		if !IsValidTagName(M.TagKey()) {
			err = fmt.Errorf("Malformed tag")
			return
		}
	}

	{ // Check the target type
		V, isValue := v.(reflect.Value)
		var T reflect.Type
		if isValue {
			T = V.Type()
		} else {
			V = reflect.ValueOf(v)
			T = reflect.TypeOf(v)
		}

		// Check if there is a method to call instead
		var ok bool
		data, ok, err = tryMarshal(T, V, M.MarshalMethodName())
		if err != nil || ok {
			return
		}
	}

	return M.MarshalDefault(v)
}

// MarshalDefault represents the absence of a MarshalX method
// (where X is determined by the March instance). It is the "sane default"
// of marshalers, and is based on JSON.
func (M March) MarshalDefault(v interface{}) (data []byte, err error) {
	if M.DefaultMarshaler != nil {
		return M.DefaultMarshaler(v)
	}
	return M.MarshalAsJSON(v)
}

// Unmarshal takes any type with tags at the given tag key (determined
// by the value of M.TagKey()) and recursively unmarshals onto the value v.
// By default in JSON, or by a custom unmarshal method if one exists on
// the given type.
func (M March) Unmarshal(data []byte, v interface{}) (err error) {
	// Sanity check
	if !IsValidTagName(M.TagKey()) {
		return fmt.Errorf("Malformed tag")
	}

	{ // Check the type of v
		V, isValue := v.(reflect.Value)
		var T reflect.Type
		if isValue {
			T = V.Type()
		} else {
			V = reflect.ValueOf(v)
			T = reflect.TypeOf(v)
		}
		kind := V.Kind()
		if kind != reflect.Ptr && !isValue {
			return fmt.Errorf("Value is not a nonzero pointer, slice, or reflect.Value")
		}

		// Check if there is a method to call instead
		var ok bool
		ok, err = tryUnmarshal(T, V, data, M.UnmarshalMethodName())
		if err != nil || ok {
			return
		}
	}

	return M.UnmarshalDefault(data, v)
}

// UnmarshalDefault represents the absence of an UnmarshalX method
// (where X is determined by the March instance). It is the "sane default"
// of unmarshalers, and is based on a JSON implementation of ReadFields.
func (M March) UnmarshalDefault(data []byte, v interface{}) (err error) {
	if M.DefaultUnmarshaler != nil {
		return M.DefaultUnmarshaler(data, v)
	}
	return M.UnmarshalAsJSON(data, v)
}

// Extension is a method which MAY be called by an encoder,
// usually according to the presence of flags.
type Extension func(March, *Values, *reflect.Value, *FieldDescriptor, int) (bool, error)

// DefaultExtensions contains basic features supported by March out-of-the-box.
var DefaultExtensions = map[string]Extension{
	"hoist": func(M March, values *Values, vfield *reflect.Value, tfield *FieldDescriptor, i int) (skip bool, err error) {
		*values = append(*values, *vfield) // Hoist
		skip = true
		return
	},
}

// GetExtensions returns the extensions map on the struct or a default
func (M March) GetExtensions() map[string]Extension {
	if M.Extensions != nil {
		return M.Extensions
	}
	return DefaultExtensions
}
