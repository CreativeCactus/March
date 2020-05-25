package march

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func isValidTag(tag string) bool {
	// There is currently only one rule enforced by this library.
	// The specification is unclear about tag keys.
	return len(tag) > 0
}

// MarshalJSON marshals to JSON via of WriteFieldsJSON.
func (M March) MarshalJSON(v interface{}) (data []byte, err error) {
	if M.Debug {
		fmt.Printf("MarshalJSON: %#v\n", v)
	}

	// Sanity check
	if !isValidTag(M.Tag) {
		err = fmt.Errorf("Malformed tag")
	}

	// Check the target type
	target := reflect.ValueOf(v)
	V := target
	if target.Kind() == reflect.Ptr {
		// err = fmt.Errorf("Value is not a pointer")
		V = target.Elem()
		// return
	}

	// Get json fields
	nf := V.Type().NumField()
	output := map[string][]byte{}

	// Iterate over all fields
	for i := 0; i < nf; i++ {
		vfield := V.Field(i)
		tfield := V.Type().Field(i)

		if !vfield.CanInterface() {
			if M.Debug {
				fmt.Println("CANNOT ACCESS", tfield.Name)
			}
			continue
		}

		// Check for reasons to skip this field
		tag, ok := tfield.Tag.Lookup(M.Tag)
		if !ok {
			continue // The struct did not tell us to write any data from this field
		}

		// Marshal onto the output
		// fv := reflect.New(tfield.Type).Interface()
		ok, data, err = M.tryMarshal(tfield.Type, vfield)

		if err == nil && !ok {
			output[tag], err = json.Marshal(vfield.Interface())
		}

		if err != nil {
			return
		}
		if err != nil {

		}
	}

	// Finally marshal to the result
	data, err = WriteFieldsJSON(output)
	return
}

// WriteFieldsJSON is the JSON implementation of WriteFields*.
// It represents a way of encoding the top level of a message
// into bytes. It is the last stage of marshalling.
func WriteFieldsJSON(fields map[string][]byte) (data []byte, err error) {
	// Set json fields
	firstField := true
	data = []byte("{")
	for k, v := range fields {
		if !firstField {
			data = append(data, ',')
		}
		data = append(data, '"')
		data = append(data, []byte(k)...)
		data = append(data, '"')
		data = append(data, ':')
		data = append(data, v...)
		firstField = false
	}
	data = append(data, '}')
	return
}

// UnmarshalJSON unmarshals from JSON via of ReadFieldsJSON.
func (M March) UnmarshalJSON(data []byte, v interface{}) (err error) {
	var input map[string][]byte
	V := reflect.ValueOf(v)

	// Sanity check
	target := V.Elem()
	if target.Type().Kind() != reflect.Struct {
		return fmt.Errorf("Default unmarshaller does not support non-struct types yet. Implement %s or use a struct", M.UnmarshalMethodName())
	}

	if M.Debug {
		fmt.Printf("\tUnmarshaling %s\n", target.Type().Name())
	}

	{ // Get input fields using a custom method or the default JSON
		var ok bool
		ok, input, err = M.tryReadFields(V.Type(), V, data)
		if err != nil {
			return fmt.Errorf("%s failed: %s%w", M.ReadFieldsMethodName(), err.Error(), err)
		}
		if !ok {
			input, err = ReadFieldsJSON(data)
		}
		if err != nil {
			return fmt.Errorf("ReadFieldsJSON failed: %s%w", err.Error(), err)
		}
	}

	// Iterate over all fields
	nf := target.Type().NumField()
	for i := 0; i < nf; i++ {
		vfield := target.Field(i)
		tfield := target.Type().Field(i)

		// Check for reasons to skip this field
		if !vfield.CanSet() {
			continue // Unassignable or unexported field
		}
		if M.Debug {
			fmt.Printf("\tUnmarshaling %s\n", tfield.Name)
		}
		tag, ok := tfield.Tag.Lookup(M.Tag)
		if !ok {
			continue // The struct did not tell us to read any fields from data
		}
		ifield, ok := input[tag]
		if !ok {
			continue // There is no data to put here
		}

		// Unmarshal onto a new value of the same type as field, then assign it
		fv := reflect.New(tfield.Type).Interface()
		ok, err = M.tryUnmarshal(tfield.Type, vfield, data)

		if err == nil && !ok {
			err = json.Unmarshal(ifield, &fv)
		}

		if err != nil {
			if M.Verbose {
				fmt.Printf("Field %s encountered: %s", tag, err.Error())
			}
			if !M.Relax {
				return err
			}
			err = nil
		}

		result := reflect.ValueOf(fv)
		if kind := result.Kind(); kind != reflect.Interface && kind != reflect.Ptr {
			continue // Seems like a zero value
		}
		vfield.Set(result.Elem())
	}
	return nil
}

// ReadFieldsJSON is the JSON implementation of ReadFields*.
// It represents a way of decoding the top level of a message
// into parts which can be indexed by tags. It is the first stage of unmarshaling.
func ReadFieldsJSON(data []byte) (fields map[string][]byte, err error) {
	// Get json fields
	fields = map[string][]byte{}
	inputs := map[string]json.RawMessage{}
	err = json.Unmarshal(data, &inputs)
	if err != nil {
		return
	}
	for k, field := range inputs {
		fields[k] = field
	}
	return
}
