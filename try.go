package march

import (
	"fmt"
	"reflect"
)

// TagKey returns the Tag property on M or a sane default
// It should be used instead of M.Tag directly
func (M March) TagKey() string {
	if len(M.Tag) > 0 {
		return M.Tag
	}
	return "March"
}

// MethodSuffix returns the string which appears at the end of custom methods as X (see below)
// It determines the name of the method to be called on types during marshal operations.
func (M March) MethodSuffix() string {
	if len(M.Suffix) > 0 {
		return M.Suffix
	}
	if tag := M.TagKey(); len(tag) > 0 {
		return tag
	}
	return ""
}

// UnmarshalMethodName returns the name of the Unmarshal method to look for on custom types
func (M March) UnmarshalMethodName() string {
	return fmt.Sprintf("Unmarshal%s", M.MethodSuffix())
}

// MarshalMethodName returns the name of the Marshal method to look for on custom types
func (M March) MarshalMethodName() string {
	return fmt.Sprintf("Marshal%s", M.MethodSuffix())
}

// ReadFieldsMethodName returns the name of the ReadFields method to look for on custom types
func (M March) ReadFieldsMethodName() string {
	return fmt.Sprintf("ReadFields%s", M.MethodSuffix())
}

// WriteFieldsMethodName returns the name of the WriteFields method to look for on custom types
func (M March) WriteFieldsMethodName() string {
	return fmt.Sprintf("WriteFields%s", M.MethodSuffix())
}

// ptr returns a reflect.Value containing a pointer to the given Values content
// eg. v => *v, *v => **v
func ptr(v reflect.Value) reflect.Value {
	pt := reflect.PtrTo(v.Type())
	pv := reflect.New(pt.Elem())
	pv.Elem().Set(v)
	return pv
}

// tryMarshal attempts to call a custom unmarshal method on the given type/value
func tryMarshal(t reflect.Type, v reflect.Value, method string) (data []byte, ok bool, err error) {
	if v.IsZero() && v.Kind() == reflect.Ptr {
		ok = false
		return
	}
	data, ok, err = callMarshal(t, v, method, []reflect.Value{v})
	if err != nil || ok {
		return
	}
	V := reflect.New(v.Type())
	T := V.Type()
	V.Elem().Set(v)
	return callMarshal(T, V, method, []reflect.Value{V})
}
func callMarshal(t reflect.Type, v reflect.Value, method string, args []reflect.Value) (data []byte, ok bool, err error) {
	var res []reflect.Value
	res, ok = TryCall(t, v, method, args)
	if !ok {
		return
	}
	// TODO replace this with calls to NumIn/NumOut/In/Out
	if len(res) != 2 {
		panic(fmt.Sprintf("Implementation of %s returned %d values, wanted 2", method, len(res)))
	}
	var eok, dok bool

	d := res[0].Interface()
	data, dok = d.([]byte)
	if !dok {
		err = fmt.Errorf("%s implementation returned non map[string][]byte value as first result: %T", method, d)
	}

	ferr := res[1].Interface()
	err, eok = ferr.(error)
	if !eok && err != nil {
		err = fmt.Errorf("%s implementation returned non error/nil value as second result: %T", method, ferr)
	}

	return
}

// tryUnmarshal attempts to call a custom unmarshal method on the given type/value
// It will first check for a the specified method on *v and create a new pointer to v if needed.
func tryUnmarshal(t reflect.Type, v reflect.Value, data []byte, method string) (ok bool, err error) {
	return tryCallAndGetError(t, v, method, []reflect.Value{
		v,
		reflect.ValueOf(data),
	})
}

// tryReadFields attempts to call a custom input field getter method on the given type/value
func tryReadFields(t reflect.Type, v reflect.Value, data []byte, method string) (fields map[string][]byte, ok bool, err error) {
	var res []reflect.Value
	res, ok = TryCall(t, v, method, []reflect.Value{
		v,
		reflect.ValueOf(data),
	})
	if !ok {
		return
	}
	// TODO replace this with calls to NumIn/NumOut/In/Out
	if len(res) != 2 {
		panic(fmt.Sprintf("Implementation of %s returned %d values, wanted 2", method, len(res)))
	}
	var eok, fok bool

	f := res[0].Interface()
	fields, fok = f.(map[string][]byte)
	if !fok {
		err = fmt.Errorf("%s implementation returned non map[string][]byte value as first result: %T", method, f)
	}

	ferr := res[1].Interface()
	err, eok = ferr.(error)
	if !eok && err != nil {
		err = fmt.Errorf("%s implementation returned non error/nil value as second result: %T", method, ferr)
	}

	return
}

// tryWriteFields attempts to call a custom output field setter method on the given type/value
func tryWriteFields(t reflect.Type, v reflect.Value, fields map[string][]byte, method string) (data []byte, ok bool, err error) {
	var res []reflect.Value
	res, ok = TryCall(t, v, method, []reflect.Value{
		v,
		reflect.ValueOf(data),
	})
	if !ok {
		return
	}
	// TODO replace this with calls to NumIn/NumOut/In/Out
	if len(res) != 2 {
		panic(fmt.Sprintf("Implementation of %s returned %d values, wanted 2", method, len(res)))
	}
	var eok, dok bool

	d := res[0].Interface()
	data, dok = d.([]byte)
	if !dok {
		err = fmt.Errorf("%s implementation returned non []byte value as first result: %T", method, d)
	}

	ferr := res[1].Interface()
	err, eok = ferr.(error)
	if !eok && err != nil {
		err = fmt.Errorf("%s implementation returned non error/nil value as second result: %T", method, ferr)
	}

	return
}

// tryCallAndGetError attempts to call a method on the given type/value
// It expects that the function will return one value.
// The returned value must be err or nil.
// If err is nil and ok is true, then the call succeeded.
// If err is nil and ok is false, no method was found.
// If err is not nil then an attempt was made to call the method,
// which failed or returned an unexpected result.
func tryCallAndGetError(t reflect.Type, v reflect.Value, method string, args []reflect.Value) (ok bool, err error) {
	var res []reflect.Value
	res, ok = TryCall(t, v, method, args)
	if !ok {
		return
	}
	if len(res) != 1 {
		err = fmt.Errorf("%s implementation returned %d values, expected one error", method, len(res))
	}
	var eok bool
	first := res[0].Interface()
	err, eok = first.(error)
	if !eok && err != nil {
		err = fmt.Errorf("%s implementation returned non error/nil value as first result: %T", method, first)
	}
	return
}

// TryCall attempts to make a call on a reflected type/value with provided args
// Returns the result of the call (if any) as reflect.Values and ok to indicate if the method was found.
func TryCall(t reflect.Type, v reflect.Value, method string, args []reflect.Value) (res []reflect.Value, ok bool) {
	var m reflect.Method
	m, ok = t.MethodByName(method)
	if ok {
		res = m.Func.Call(args)
	}
	return
}
