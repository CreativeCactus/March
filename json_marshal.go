package march

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// MarshalAsJSON marshals to JSON via WriteFieldsJSON.
// v must be a reflect.Value or the value to marshal.
// Use reflect.ValueOf(v) twice if trying to marshal reflect.Value.
func (M March) MarshalAsJSON(v interface{}) (data []byte, err error) {
	V, ok := v.(reflect.Value)
	if !ok {
		V = reflect.ValueOf(v)
	}
	T := V.Type()

	if !M.NoMarshalJSON { // No matter what it is, if it already has a MarshalJSON method
		// Then use that instead of the default march JSON marshaler
		// First, check *T, since having a Un/Marshal methods on the base type is rare
		pV := ptr(V)
		data, ok, err = tryMarshal(pV.Type(), pV, "MarshalJSON")
		if err != nil || ok {
			return
		}

		// Try on the base type, just in case
		data, ok, err = tryMarshal(T, V, "MarshalJSON")
		if err != nil || ok {
			return
		}
	}

	switch k := V.Kind(); k {
	case reflect.Slice, reflect.Array:
		if V.IsNil() {
			return []byte("[]"), nil
		}
		return M.marshalJSONSlice(V)
	// case reflect.Map:
	// TODO implement specific support for custom types under maps
	// https://golang.org/ref/spec#Map_types
	case reflect.Ptr:
		if V.IsNil() {
			return []byte("null"), nil
		}
		V = V.Elem()
		return M.Marshal(V)
	case reflect.Struct:
		return M.marshalJSONStruct(V)
	default:
		return json.Marshal(V.Interface())
	}
}

func (M March) marshalJSONStruct(v reflect.Value) (data []byte, err error) {
	output := map[string][]byte{}
	{ // Iterate over all fields
		values := Values{v}
	fields:
		for i := 0; i < values.TotalFields(); i++ {
			vfield, tfield, ok := values.FieldAt(i, M.TagKey())

			{ // Check for issues
				if !ok {
					err = fmt.Errorf("Failed to get field %d/%d from %d values", i, values.TotalFields(), len(values))
					return
				}

				if !vfield.CanInterface() {
					continue
				}
			}

			tag := tfield.TagName
			if len(tag) <= 0 {
				continue // The struct did not tell us to write any data from this field
			}

			{ // Check flags
				for name, extension := range M.GetExtensions() {
					if tfield.FlagsContain(name) {
						skip := false
						skip, err = extension(M, &values, &vfield, &tfield, i)
						if err != nil {
							err = fmt.Errorf("Extension %s: Error on field %d/%d: %w", name, i, values.TotalFields(), err)
							return
						}
						if skip {
							continue fields // Handled later in the loop
						}
					}
				}
			}

			output[tag], err = M.Marshal(vfield)
			if err != nil {
				if M.Verbose {
					fmt.Printf("Marshaling field %s: %s", tag, err.Error())
				}
				if M.Strict {
					return
				}
				err = nil
			}
		}
	}

	{ // Write out fields using a custom method or the default JSON
		var ok bool
		data, ok, err = tryWriteFields(v.Type(), v, output, M.WriteFieldsMethodName())
		if err != nil {
			err = fmt.Errorf("%s failed: %w", M.WriteFieldsMethodName(), err)
			return
		}
		if !ok {
			// TODO? Prevent marshaling duplicate keys
			data, err = WriteFieldsJSON(output)
		}
		if err != nil {
			err = fmt.Errorf("WriteFieldsJSON failed: %w", err)
			return
		}
	}

	return
}

func (M March) marshalJSONSlice(v reflect.Value) (data []byte, err error) {
	datas := [][]byte{}
	nested := []byte{}
	for i := 0; i < v.Len(); i++ {
		nested, err = M.Marshal(v.Index(i).Interface())
		if err != nil {
			return
		}
		datas = append(datas, nested)
	}
	data = []byte("[")
	data = append(data, bytes.Join(datas, []byte(","))...)
	data = append(data, ']')

	return
}

// WriteFieldsJSON is the JSON implementation of WriteFields*.
// It represents a way of encoding the top level of a message
// into bytes. It is the last stage of marshaling.
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
