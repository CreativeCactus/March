package march

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
)

var debug = len(os.Getenv("DEBUG")) > 0

// FlagRemain denotes the field to which remaining JSON fields can be unmarshalled
const FlagRemain = "remains"

// FlagHoist denotes a type whose values are hoisted to the parent struct when marshalling to JSON
const FlagHoist = "hoist"

// March is the top level interface for Un/Marshalling
type March struct {
	// TODO construct and make .tag private to avoid confusion with defaults
	Tag                 string                            // The tag key to look up on structs
	Suffix              string                            // An optional override for custom functions eg. MarshalSUFFIX. Defaults to Tag
	Verbose             bool                              // Used in some cases to show field un/marshalling errors
	Strict              bool                              // Determines whether a failure to un/marshal a field results in a failure overall
	Debug               bool                              // Print debug logs
	DefaultMarshaller   func(interface{}) ([]byte, error) // Override the default marshaller for types with no custom marshal function
	DefaultUnmarshaller func([]byte, interface{}) error   // Override the default unmarshaller for types with no custom unmarshal function
}

// RawUnmarshal is a wrapper around json.RawMessage which
// provides some helpers for unmarshalling at runtime.
type RawUnmarshal struct {
	Data  json.RawMessage
	March March
}

// MarshalFrom allows updating the underlying data via the marshaller
func (ru RawUnmarshal) MarshalFrom(v interface{}) error {
	data, err := ru.March.Marshal(v)
	if err != nil {
		return err
	}
	ru.Data = json.RawMessage(data)
	return nil
}

// UnmarshalTo allows convenient unmarshalling to a given type
func (ru RawUnmarshal) UnmarshalTo(v interface{}) error {
	return ru.March.Unmarshal(ru.Data, v)
}

// MarshalJSON ensures consistent behavior with json.RawMessage
func (ru RawUnmarshal) MarshalJSON() ([]byte, error) {
	return ru.Data.MarshalJSON()
}

// UnmarshalJSON ensures consistent behavior with json.RawMessage
func (ru RawUnmarshal) UnmarshalJSON(data []byte) error {
	return ru.Data.UnmarshalJSON(data)
}

// Marshal provides convenient defaults for March{}.Marshal
// Given any type, returns the representative []byte according to
// marshaller precedence.
// Refer to type March for details.
func Marshal(v interface{}) (data []byte, err error) {
	return March{}.Marshal(v)
}

// Unmarshal provides convenient defaults for March{}.Unmarshal
// Given []byte and any type, updates the pointed type according to
// unmarshaller precedence.
// Refer to type March for details.
func Unmarshal(data []byte, v interface{}) (err error) {
	return March{}.Unmarshal(data, v)
}

// Marshal takes any type with tags at the given tag key (determined
// by the value of M.TagKey()) and returns a recursively marshalled []byte,
// by default in JSON, or by a custom marshal method if one exists on
// the given type.
func (M March) Marshal(v interface{}) (data []byte, err error) {
	if M.Debug {
		fmt.Printf("Marshal: %#v\n", v)
	}

	// Sanity check
	if !IsValidTagName(M.TagKey()) {
		err = fmt.Errorf("Malformed tag")
		return
	}

	// Check the target type
	target := reflect.ValueOf(v)
	V := target
	// if target.Kind() == reflect.Ptr {
	// V = target.Elem() // Don't do this :)
	// }

	// Check if there is a method to call instead
	var ok bool
	data, ok, err = tryMarshal(V.Type(), V, M.MarshalMethodName())
	if M.Debug {
		fmt.Printf("Marshal: %t, %#v\n", ok, err)
	}
	if err != nil || ok {
		return
	}

	return M.MarshalDefault(v)
}

// MarshalDefault represents the absence of a MarshalX method
// (where X is determined by the March instance). It is the "sane default"
// of marshalers, and is based on JSON.
func (M March) MarshalDefault(v interface{}) (data []byte, err error) {
	if M.Debug {
		fmt.Printf("MarshalDefault: %#v\n", v)
	}
	if M.DefaultMarshaller != nil {
		return M.DefaultMarshaller(v)
	}

	return M.MarshalJSON(v)
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

	// Check the type of v
	V := reflect.ValueOf(v)
	kind := V.Kind()
	if kind != reflect.Ptr || V.IsNil() {
		return fmt.Errorf("Value is not a nonzero pointer or slice")
	}

	// Check if there is a method to call instead
	var ok bool
	ok, err = tryUnmarshal(V.Type(), V, data, M.UnmarshalMethodName())
	if err != nil || ok {
		return
	}

	return M.UnmarshalDefault(data, v)
}

// UnmarshalDefault represents the absence of an UnmarshalX method
// (where X is determined by the March instance). It is the "sane default"
// of unmarshalers, and is based on a JSON implementation of ReadFields.
func (M March) UnmarshalDefault(data []byte, v interface{}) (err error) {
	if M.DefaultUnmarshaller != nil {
		return M.DefaultUnmarshaller(data, v)
	}
	return M.UnmarshalJSON(data, v)
}
