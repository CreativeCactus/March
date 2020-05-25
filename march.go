package march

import (
	"fmt"
	"reflect"
)

type March struct {
	Tag     string
	Suffix  string // eg JSON for UnmarshalJSON to be checked for during unmarshalling
	Verbose bool
	Relax   bool // Determines whether a failure to un/marshal a field results in a failure overall
	Debug   bool
}

func (M March) Marshal(v interface{}) (data []byte, err error) {
	if M.Debug {
		fmt.Printf("Marshal: %#v\n", v)
	}

	// Sanity check
	if !isValidTag(M.Tag) {
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
	ok, data, err = M.tryMarshal(V.Type(), V)
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

	return M.MarshalJSON(v)
}

func (M March) Unmarshal(data []byte, v interface{}) (err error) {
	// Sanity check
	if !isValidTag(M.Tag) {
		return fmt.Errorf("Malformed tag")
	}

	// Check the type of v
	V := reflect.ValueOf(v)
	if V.Kind() != reflect.Ptr || V.IsNil() {
		return fmt.Errorf("Value is not a nonzero pointer")
	}

	// Check if there is a method to call instead
	var ok bool
	ok, err = M.tryUnmarshal(V.Type(), V, data)
	if err != nil || ok {
		return
	}

	return M.UnmarshalDefault(data, v)
}

// UnmarshalDefault represents the absence of an UnmarshalX method
// (where X is determined by the March instance). It is the "sane default"
// of unmarshalers, and is based on a JSON implementation of ReadFields.
func (M March) UnmarshalDefault(data []byte, v interface{}) (err error) {
	return M.UnmarshalJSON(data, v)
}
