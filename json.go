package march

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// MarshalJSON marshals to JSON via of WriteFieldsJSON.
func (M March) MarshalJSON(v interface{}) (data []byte, err error) {
	if M.Debug {
		fmt.Printf("MarshalJSON: %#v\n", v)
	}

	// Sanity check
	if !IsValidTagName(M.Tag) {
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
	// nf := V.Type().NumField()
	values := Values{V}
	output := map[string][]byte{}
	{
		// Iterate over all fields
		for i := 0; i < values.TotalFields(); i++ {
			// vfield := V.Field(i)
			// tfield := V.Type().Field(i)
			vfield, tfield, ok := values.FieldAt(i, M.Tag)

			if !ok {
				err = fmt.Errorf("Failed to get field %d/%d from %d values", i, values.TotalFields(), len(values))
				return
			}

			if !vfield.CanInterface() {
				if M.Debug {
					fmt.Println("CANNOT ACCESS", tfield.TagName)
				}
				continue
			}

			// Check for reasons to skip this field
			tag := tfield.TagName
			if len(tag) <= 0 {
				continue // The struct did not tell us to write any data from this field
			}
			tagName := tfield.TagName
			if len(tagName) <= 0 {
				continue // The tag lacks a primary value
			}
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

			// Marshal onto the output
			data, ok, err = M.tryMarshal(tfield.Type, vfield)

			if err == nil && !ok {
				output[tagName], err = json.Marshal(vfield.Interface())
			}
			if M.Verbose {
				fmt.Printf("Marshalling field %s: %s", tagName, err.Error())
			}
		}
	}
	if debug {
		fmt.Printf("Output %#v\n", output)
	}
	{ // Write out fields using a custom method or the default JSON
		var ok bool
		data, ok, err = M.tryWriteFields(V.Type(), V, output)
		if err != nil {
			err = fmt.Errorf("%s failed: %s%w", M.WriteFieldsMethodName(), err.Error(), err)
			return
		}
		if !ok {
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
		input, ok, err = M.tryReadFields(V.Type(), V, data)
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
		tagName, ok := GetTagPart(tag, 0)
		if !ok || !IsValidTagName(tagName) {
			continue // The tag lacks a primary value
		}
		if FlagsContain(tag, FlagRemain) {
			remainsReceiver = append(remainsReceiver, vfield)
			continue // Handled in another loop
		}

		didUnmarshal = append(didUnmarshal, tagName)
		ifield, ok := input[tagName]
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
				fmt.Printf("Unmarshalling Field %s: %s", tagName, err.Error())
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

	// Perform a second stage for the remains flag(s)
	for _, field := range didUnmarshal {
		delete(input, field)
	}
	for _, value := range remainsReceiver {
		if debug {
			fmt.Printf("Assigning remains to type %s\n", value.Type().Name())
		} // TODO refuse to assign to invalid types
		remain := toJSONMap(input)
		value.Set(reflect.ValueOf(remain))
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

func toJSONMap(input map[string][]byte) (output map[string]json.RawMessage) {
	output = map[string]json.RawMessage{}
	for k, v := range input {
		output[k] = v
	}
	return
}
