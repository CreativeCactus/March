package march

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

// MarshalJSON marshals to JSON via of WriteFieldsJSON.
func (M March) MarshalJSON(v interface{}) (data []byte, err error) {
	if M.Debug {
		fmt.Printf("MarshalJSON: %#v\n", v)
	}

	// Check the target type
	target := reflect.ValueOf(v)
	V := target
	kind := target.Kind()
	if kind == reflect.Ptr {
		V = target.Elem()
		// TODO Mark a recursive call to allow arbitrary depths of pointers
	}
	if kind == reflect.Slice || kind == reflect.Array {
		datas := [][]byte{}
		nested := []byte{}
		for i := 0; i < target.Len(); i++ {
			nested, err = M.Marshal(target.Index(i).Interface())
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

	output := map[string][]byte{}
	{ // Iterate over all fields
		values := Values{V}
		for i := 0; i < values.TotalFields(); i++ {
			vfield, tfield, ok := values.FieldAt(i, M.TagKey())

			{ // Check for issues
				if !ok {
					err = fmt.Errorf("Failed to get field %d/%d from %d values", i, values.TotalFields(), len(values))
					return
				}

				if !vfield.CanInterface() {
					if M.Debug {
						fmt.Printf("Access error: %s\n", tfield.TagName)
					}
					continue
				}
			}

			tag := tfield.TagName
			if len(tag) <= 0 {
				continue // The struct did not tell us to write any data from this field
			}
			tagName := tfield.TagName
			if len(tagName) <= 0 {
				continue // The tag lacks a primary value
			}

			{ // Check flags
				if tfield.FlagsContain(FlagHoist) {
					before := values.TotalFields()
					values = append(values, vfield)
					if debug {
						fmt.Printf("Appended %s (total %d => %d) %#v\n", tagName, before, values.TotalFields(), vfield)
					}
					continue // Handled later in the loop
				}
				if debug {
					fmt.Printf("Iterating %d %s %#v\n", i, tagName, vfield.Interface())
				}
			}

			{ // Marshal onto the output
				data, ok, err = tryMarshal(tfield.Type, vfield, M.MarshalMethodName())

				if err == nil && !ok {
					output[tagName], err = json.Marshal(vfield.Interface())
				}
				if err != nil {
					if M.Verbose {
						fmt.Printf("Marshalling field %s: %s", tagName, err.Error())
					}
					if M.Strict {
						return
					}
					err = nil
				}

			}
		}
	}
	if debug {
		fmt.Printf("Output %#v\n", output)
	}
	{ // Write out fields using a custom method or the default JSON
		var ok bool
		data, ok, err = tryWriteFields(V.Type(), V, output, M.WriteFieldsMethodName())
		if err != nil {
			err = fmt.Errorf("%s failed: %s%w", M.WriteFieldsMethodName(), err.Error(), err)
			return
		}
		if !ok {
			// TODO? Prevent marshalling duplicate keys
			data, err = WriteFieldsJSON(output)
		}
		if err != nil {
			err = fmt.Errorf("WriteFieldsJSON failed: %s%w", err.Error(), err)
			return
		}
	}

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
	didUnmarshal := []string{}
	remainsReceiver := []reflect.Value{}
	V := reflect.ValueOf(v)
	T := reflect.TypeOf(v)

	if T.Kind() != reflect.Ptr {
		return fmt.Errorf("Must be pointer")
	}

	target := V.Elem()
	if M.Debug {
		fmt.Printf("\tUnmarshaling TOP LEVEL %s\n", target.Type().Name())
	}

	// TODO modularize these blocks and produce a better model for Un/Marshaller implementation
	// EG. the value/struct distinction.
	{ // Sanity check and type switching
		switch target.Type().Kind() {
		case reflect.Struct: // Carry on, the rest of the function is for structs
		case reflect.Map: // Carry on, the rest of the function should work for maps
		case reflect.Array:
			return fmt.Errorf("Default unmarshaller does not support array types. Use a slice")
		case reflect.Slice:
			elems := []json.RawMessage{}
			if err = json.Unmarshal(data, &elems); err != nil {
				return fmt.Errorf("Failed to unmarshal slice: %s%w", err.Error(), err)
			}

			//(Re)initialize the slice
			elemType := target.Type().Elem()
			x := reflect.New((T.Elem()))
			for _, e := range elems {
				elem := reflect.New(elemType).Elem()
				_, err = M.UnmarshalValueJSON(elemType, elem, (e))
				x.Elem().Set(reflect.Append(x.Elem(), elem))
			}
			target.Set(x.Elem())
			return
		default: // Perform some primitive unmarshalling
			_, err = M.UnmarshalValueJSON(target.Type(), target, data)
			return
		}
	}

	{ // Get input fields using a custom method or the default JSON
		var ok bool
		input, ok, err = tryReadFields(V.Type(), V, data, M.ReadFieldsMethodName())
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
		// TODO abstract this block out
		nf := NumField(target)
		for i := 0; i < nf; i++ {
			vfield, tfield, ok := NthField(target, i, M.TagKey())
			if M.Debug {
				fmt.Printf("\tUnmarshaling %s: %s\n", tfield.TagName, tfield.Type.Name())
			}

			// Check for reasons to skip this field
			if !vfield.CanSet() {
				continue // Unassignable or unexported field
			}
			if !ok || !IsValidTagName(tfield.TagName) {
				continue // The tag lacks a primary value
			}
			if debug {
				fmt.Printf("DEBUG FLAGS: %+v\n", tfield.TagFlags)
			}

			if tfield.FlagsContain(FlagRemain) {
				remainsReceiver = append(remainsReceiver, vfield)
				if debug {
					fmt.Printf("DEBUG : %s\n", tfield.Type.Name())
				}

				continue // Handled in another loop
			}

			didUnmarshal = append(didUnmarshal, tfield.TagName)
			ifield, ok := input[tfield.TagName]
			if !ok {
				continue // There is no data to put here
			}

			{ // Unmarshal onto a new value of the same type as field, then assign it
				_, err = M.UnmarshalValueJSON(tfield.Type, vfield, ifield)
				if err != nil && M.Verbose {
					fmt.Printf("Error Unmarshalling %s: %s", tfield.TagName, err.Error())
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
			if debug {
				fmt.Printf("Assigning remains to type %s\n", value.Type().Name())
			}
			k := value.Kind()

			switch k {
			case reflect.Map:
				k, v, _ := mapType(value)
				if debug {
					fmt.Printf("DEBUG NOTE: [%s]%s\n", k.Name(), v.Name())
				}
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
				// Note that support for Struct would be rendered obsolete by support for dot notation.
			default:
				panic(fmt.Sprintf("Unmarshal remaining fields onto unknown type %s", value.Type().Name()))
			}

		}
	}

	return nil
}

// UnmarshalValueJSON unmarshals a single value (optionally using a custom unmarshaller).
// custom is true if a custom unmarshaller was used.
func (M March) UnmarshalValueJSON(t reflect.Type, v reflect.Value, data []byte) (custom bool, err error) {
	fv := reflect.New(t).Interface()
	custom, err = tryUnmarshal(t, reflect.ValueOf(fv), data, M.UnmarshalMethodName())

	if err == nil && !custom {
		err = json.Unmarshal(data, &fv)
	}

	if err != nil {
		return
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
			Data:  v,
			march: m,
		}
	}
	return
}
