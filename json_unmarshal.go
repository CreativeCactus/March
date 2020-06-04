package march

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// UnmarshalAsJSON unmarshals from JSON via of ReadFieldsJSON.
// v must be either a *reflect.Value or the value to unmarshal.
func (M March) UnmarshalAsJSON(data []byte, v interface{}) (err error) {
	pV, ok := v.(*reflect.Value)
	var V reflect.Value
	if !ok {
		V = reflect.ValueOf(v)
	} else {
		V = *pV
	}
	T := V.Type()
	// If v was passed as a non-reflect.Value, then it must be a pointer
	// Otherwise unmarshal will have no effect.
	if !ok && T.Kind() != reflect.Ptr {
		return fmt.Errorf("Must be pointer")
	}

	if !M.NoUnmarshalJSON { // No matter what V is, if it already has an UnmarshalJSON method
		// Then use that instead of the default march JSON unmarshaler.
		// First, check *T, since having a Un/Marshal methods on the base type is rare
		pV := ptr(V)
		ok, err = tryUnmarshal(pV.Type(), pV, data, "UnmarshalJSON")
		if err != nil || ok {
			V.Set(pV.Elem())
			return
		}

		// Try on the base type, just in case
		ok, err = tryUnmarshal(T, V, data, "UnmarshalJSON")
		if err != nil || ok {
			return
		}
	}

	{ // Sanity check and type switching
		switch k := T.Kind(); k {
		case reflect.Ptr:
			return M.unmarshalJSONPtr(T, V, data)
		case reflect.Struct:
			return M.unmarshalJSONStruct(T, V, data)
		// case reflect.Map:
		// TODO implement specific support for custom types under maps
		// https://golang.org/ref/spec#Map_types
		case reflect.Array:
			return fmt.Errorf("Default JSON unmarshaler does not support array types. Use a slice")
		case reflect.Slice:
			return M.unmarshalJSONSlice(T, V, data)
		default: // Perform some primitive unmarshaling
			if V.Type().Kind() == reflect.Interface {
				V = V.Elem()
			}
			return M.unmarshalJSONValue(V.Type(), V, data)
		}
	}
}
func (M March) unmarshalJSONSlice(t reflect.Type, v reflect.Value, data json.RawMessage) (err error) {
	elems := []json.RawMessage{}
	if err = json.Unmarshal(data, &elems); err != nil {
		return fmt.Errorf("Failed to unmarshal slice: %s%w", err.Error(), err)
	}

	elemType := v.Type().Elem()
	slice := reflect.New(t)
	{ // (Re)initialize the slice
		for _, e := range elems {
			elem := reflect.New(elemType)
			err = M.Unmarshal(e, &elem)
			if err != nil {
				return
			}
			slice.Elem().Set(reflect.Append(slice.Elem(), elem.Elem()))
		}
	}

	v.Set(slice.Elem())
	return
}
func (M March) unmarshalJSONStruct(t reflect.Type, v reflect.Value, data json.RawMessage) (err error) {
	var input map[string][]byte
	didUnmarshal := []string{}
	remainsReceiver := []reflect.Value{}

	{ // Get input fields using a custom method or the default JSON
		var ok bool
		input, ok, err = tryReadFields(v.Type(), v, data, M.ReadFieldsMethodName())
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

	{ // Iterate over all fields
		nf := NumField(v)
		for i := 0; i < nf; i++ {
			vfield, tfield, ok := NthField(v, i, M.TagKey())

			{ // Pre checks
				// Check for reasons to skip this field
				if !vfield.CanSet() {
					continue // Unassignable or unexported field
				}
				if !ok || !IsValidTagName(tfield.TagName) {
					continue // The tag lacks a primary value
				}

				if tfield.FlagsContain(FlagRemain) {
					remainsReceiver = append(remainsReceiver, vfield)
					continue // Handled in another loop
				}

				// Otherwise carry on unmarshaling
				didUnmarshal = append(didUnmarshal, tfield.TagName)
			}

			{ // Unmarshal onto a new value of the same type as field, then assign it
				ifield, ok := input[tfield.TagName]
				if !ok {
					continue // There is no data to put here
				}

				if tfield.Kind == reflect.Ptr {
					field := reflect.New(tfield.Type.Elem())
					err = M.Unmarshal(ifield, &field)
					vfield.Set(field)

				} else {
					field := reflect.New(tfield.Type).Elem()
					err = M.Unmarshal(ifield, &field)
					vfield.Set(field)
				}

				if err != nil && M.Verbose {
					fmt.Printf("Error Unmarshaling %s: %s", tfield.TagName, err.Error())
				}
				if err != nil && M.Strict {
					return
				}
			}
		}
	}

	{ // Perform a second stage for the remains flag(s)
		for _, field := range didUnmarshal {
			delete(input, field)
		}
		for _, value := range remainsReceiver {
			k := value.Kind()

			switch k {
			case reflect.Map:
				k, v, _ := mapType(value)
				{ // Check the type of map
					if k.Kind() != reflect.String {
						panic(fmt.Sprintf("Unmarshal remaining fields onto map with unsupported key type %s", k.Name()))
					}
					if v == reflect.TypeOf([]byte{}) {
						value.Set(reflect.ValueOf(input))
						continue
					}
					if v == reflect.TypeOf(json.RawMessage{}) {
						value.Set(reflect.ValueOf(toJSONMap(input)))
						continue
					}
					if v == reflect.TypeOf(RawUnmarshal{}) {
						value.Set(reflect.ValueOf(toRawMap(input, M)))
						continue
					}
					panic(fmt.Sprintf("Unmarshal remaining fields onto map with unsupported value type %s", v.Name()))
				}
			case reflect.Struct, reflect.Array, reflect.Slice:
				panic(fmt.Sprintf("Unmarshal remaining fields onto unsupported type %s", value.Type().Name()))
				// Note that support for struct would be rendered obsolete by support for dot notation.
			default:
				panic(fmt.Sprintf("Unmarshal remaining fields onto unknown type %s", value.Type().Name()))
			}

		}
	}
	return nil
}

func (M March) unmarshalJSONPtr(t reflect.Type, v reflect.Value, data json.RawMessage) (err error) {
	// T := v.Type().Elem()
	// for T.Kind() == reflect.Ptr()
	// E := reflect.New(reflect.PtrTo(v.Type())) //.Elem()
	ct := t.Elem() //.Type()

	{ // Check if the underlying ptr is zero
		if ct.Kind() == reflect.Ptr {
			// Create the value under **v
			E := reflect.New(ct.Elem())
			err = M.Unmarshal(data, &E)
			if err != nil {
				return
			}
			if v.Elem().IsZero() {
				// Implied that this will never happen with the top level pointer, checked in M.Unmarshal
				v.Elem().Set(E)
			} else {
				v.Elem().Elem().Set(E.Elem())
			}
			return
		}
	}
	E := reflect.New(ct).Elem()
	err = M.Unmarshal(data, &E)
	if err != nil {
		return
	}
	v.Elem().Set(E)
	return
}

// unmarshalJSONValue unmarshals a single primitive value (optionally using a custom unmarshaler).
// custom is true if a custom unmarshaler was used.
func (M March) unmarshalJSONValue(t reflect.Type, v reflect.Value, data json.RawMessage) (err error) {
	fv := reflect.New(t).Interface()
	{ // Call an unmarshaler
		var ok bool
		ok, err = tryUnmarshal(t, reflect.ValueOf(fv), data, M.UnmarshalMethodName())
		if err == nil && !ok {
			err = json.Unmarshal(data, &fv)
		}
		if err != nil {
			return
		}
	}

	result := reflect.ValueOf(fv)
	if kind := result.Kind(); kind != reflect.Interface && kind != reflect.Ptr {
		return // Seems like a zero value
	}
	v.Set(result.Elem())
	return
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

// toJSONMap is just a type casting helper to use map[string][]byte as map[string]json.RawMessage
func toJSONMap(input map[string][]byte) (output map[string]json.RawMessage) {
	output = map[string]json.RawMessage{}
	for k, v := range input {
		output[k] = v
	}
	return
}

// toRawMap is just a type casting helper to use map[string][]byte as map[string]RawUnmarshal
func toRawMap(input map[string][]byte, m March) (output map[string]RawUnmarshal) {
	output = map[string]RawUnmarshal{}
	for k, v := range input {
		output[k] = RawUnmarshal{
			Bytes: v,
			March: m,
		}
	}
	return
}
