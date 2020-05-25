package march

import (
	"fmt"
	"reflect"
)

// MethodSuffix returns the string which appears at the end of the following methods as X
// MarshalX
// UnmarshalX
// It determines the name of the method to be called on types during marshal operations.
func (M March) MethodSuffix() string {
	if len(M.Suffix) > 0 {
		return M.Suffix
	}
	if len(M.Tag) > 0 {
		return M.Tag
	}
	return ""
}

func (M March) UnmarshalMethodName() string {
	return fmt.Sprintf("Unmarshal%s", M.MethodSuffix())
}
func (M March) MarshalMethodName() string {
	return fmt.Sprintf("Marshal%s", M.MethodSuffix())
}
func (M March) ReadFieldsMethodName() string {
	return fmt.Sprintf("ReadFields%s", M.MethodSuffix())
}
func (M March) WriteFieldsMethodName() string {
	return fmt.Sprintf("WriteFields%s", M.MethodSuffix())
}

// tryMarshal attempts to call a custom unmarshal method on the given type/value
func (M March) tryMarshal(t reflect.Type, v reflect.Value) (ok bool, data []byte, err error) {
	self := M.MarshalMethodName()
	var res []reflect.Value
	res, ok = TryCall(t, v, self, []reflect.Value{v})
	if !ok {
		return
	}
	// TODO replace this with calls to NumIn/NumOut/In/Out
	if len(res) != 2 {
		panic(fmt.Sprintf("Implementation of %s returned %d values, wanted 2", self, len(res)))
	}
	var eok, dok bool

	d := res[0].Interface()
	data, dok = d.([]byte)
	if !dok {
		err = fmt.Errorf("%s implementation returned non map[string][]byte value as first result: %T", self, d)
	}

	ferr := res[1].Interface()
	err, eok = ferr.(error)
	if !eok && err != nil {
		err = fmt.Errorf("%s implementation returned non error/nil value as second result: %T", self, ferr)
	}

	return
}

// tryUnmarshal attempts to call a custom unmarshal method on the given type/value
func (M March) tryUnmarshal(t reflect.Type, v reflect.Value, data []byte) (ok bool, err error) {
	return M.tryCallAndGetError(t, v, M.UnmarshalMethodName(), []reflect.Value{
		v,
		reflect.ValueOf(data),
	})
}

// tryReadFields attempts to call a custom input field getter method on the given type/value
func (M March) tryReadFields(t reflect.Type, v reflect.Value, data []byte) (ok bool, fields map[string][]byte, err error) {
	self := M.ReadFieldsMethodName()
	var res []reflect.Value
	res, ok = TryCall(t, v, self, []reflect.Value{
		v,
		reflect.ValueOf(data),
	})
	if !ok {
		return
	}
	// TODO replace this with calls to NumIn/NumOut/In/Out
	if len(res) != 2 {
		panic(fmt.Sprintf("Implementation of %s returned %d values, wanted 2", self, len(res)))
	}
	var eok, fok bool

	f := res[0].Interface()
	fields, fok = f.(map[string][]byte)
	if !fok {
		err = fmt.Errorf("%s implementation returned non map[string][]byte value as first result: %T", self, f)
	}

	ferr := res[1].Interface()
	err, eok = ferr.(error)
	if !eok && err != nil {
		err = fmt.Errorf("%s implementation returned non error/nil value as second result: %T", self, ferr)
	}

	return
}

// tryWriteFields attempts to call a custom output field setter method on the given type/value
func (M March) tryWriteFields(t reflect.Type, v reflect.Value, fields map[string][]byte) (ok bool, data []byte, err error) {
	self := M.WriteFieldsMethodName()
	var res []reflect.Value
	res, ok = TryCall(t, v, self, []reflect.Value{
		v,
		reflect.ValueOf(data),
	})
	if !ok {
		return
	}
	// TODO replace this with calls to NumIn/NumOut/In/Out
	if len(res) != 2 {
		panic(fmt.Sprintf("Implementation of %s returned %d values, wanted 2", self, len(res)))
	}
	var eok, dok bool

	d := res[0].Interface()
	data, dok = d.([]byte)
	if !dok {
		err = fmt.Errorf("%s implementation returned non []byte value as first result: %T", self, d)
	}

	ferr := res[1].Interface()
	err, eok = ferr.(error)
	if !eok && err != nil {
		err = fmt.Errorf("%s implementation returned non error/nil value as second result: %T", self, ferr)
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
func (M March) tryCallAndGetError(t reflect.Type, v reflect.Value, method string, args []reflect.Value) (ok bool, err error) {
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
		err = fmt.Errorf("%s implementation returned non error/nil value as first result: %T", M.UnmarshalMethodName(), first)
	}
	return
}

// TryCall attemtps to make a call on a reflected type/value with provided args
// Returns the result of the call (if any) as reflect.Values and ok to indicate if the method was found.
func TryCall(t reflect.Type, v reflect.Value, method string, args []reflect.Value) (res []reflect.Value, ok bool) {
	var m reflect.Method
	m, ok = t.MethodByName(method)
	// fmt.Printf("TryCall: %s.%s, %t (%d total)\n", t.Name(), method, ok, t.NumMethod())

	if ok {
		// fmt.Printf("TryCall: Call: %d args\n", len(args))

		res = m.Func.Call(args)
	}
	return
}
