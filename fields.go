package march

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// Utilities for dealing with tags

// GetTagPart returns the Nth part of the tag value t
// Ok is false if there was no such field.
func GetTagPart(t string, n int) (part string, ok bool) {
	parts := strings.Split(t, ",")
	ok = len(parts) > n
	if ok {
		part = parts[n]
	}
	return
}

// GetTagFlags returns the comma separated parts (except for the first)
// of the tag value t.
func GetTagFlags(t string) []string {
	parts := strings.Split(t, ",")
	if len(parts) < 2 {
		return []string{}
	}
	return parts[1:]
}

// FlagsContain returns true if any of the comma separated tag value parts
// match the given flag.
func FlagsContain(t, flag string) bool {
	for _, f := range GetTagFlags(t) {
		if f == flag {
			return true
		}
	}
	return false
}

// IsValidTagName indicates whether the given tag name is valid.
// Tag name is the first part of a value: `Key:"Name,Flag,Flag"`
func IsValidTagName(tag string) bool {
	// There is currently only one rule enforced by this library.
	// The specification is unclear about tag keys, and JSON could potentially support any string key.
	return len(tag) > 0
}

// Data types

// FieldDescriptor is a generic form of reflect.StructField which can be used for
// non-struct types' fields.
type FieldDescriptor struct {
	Type     reflect.Type
	Kind     reflect.Kind
	Tag      string
	TagName  string
	TagFlags []string
}

// FlagsContain returns true if any of the comma separated tag value parts
// match the given flag.
func (fd FieldDescriptor) FlagsContain(f string) bool {
	for _, v := range fd.TagFlags {
		if f == v {
			return true
		}
	}
	return false
}

// FieldDescriptorFromStructField converts a StructField into a FieldDescriptor.
func FieldDescriptorFromStructField(sf reflect.StructField, tagKey string) (fd FieldDescriptor, ok bool) {
	tag := sf.Tag.Get(tagKey)
	tagName := ""
	tagName, ok = GetTagPart(tag, 0)
	if !ok {
		return
	}
	t := sf.Type
	k := t.Kind()
	return FieldDescriptor{
		Tag:      tag,
		TagName:  tagName,
		TagFlags: GetTagFlags(tag),
		Type:     t,
		Kind:     k,
	}, true
}

// FieldDescriptorFromSlice converts a Slice into a FieldDescriptor.
func FieldDescriptorFromSlice(sf reflect.Value, key int, tagKey string) (fd FieldDescriptor, ok bool) {
	tag := fmt.Sprintf("%d", key)
	t := sf.Type()
	k := t.Kind()

	if k != reflect.Array && k != reflect.Slice {
		ok = false
		return
	}
	if key < 0 || key >= t.Len() {
		ok = false
		return
	}

	return FieldDescriptor{
		TagName:  tag,
		TagFlags: []string{},
		Type:     t.Elem(),
		Kind:     k,
	}, true
}

// FieldDescriptorFromMap converts a Map into a FieldDescriptor.
func FieldDescriptorFromMap(sf reflect.Value, key interface{}, tagKey string) (fd FieldDescriptor, ok bool) {
	tag := fmt.Sprintf("%v", key) //
	t := sf.Type()
	k := t.Kind()

	if k != reflect.Map {
		ok = false
		return
	}

	return FieldDescriptor{
		TagName:  tag,
		TagFlags: []string{},
		Type:     t.Elem(),
		Kind:     k,
	}, true
}

// NumField counts the fields on a reflect.Value which might not be structs
func NumField(v reflect.Value) int {
	t := v.Type()
	k := t.Kind()
	if k == reflect.Struct {
		return t.NumField()
	}
	if k == reflect.Array || k == reflect.Map || k == reflect.Slice {
		return v.Len()
	}
	return 0
}

// NthField gets the Nth field on a reflect.Value which might not be a struct
// Ok specifies whether the field was found and could be returned from the Value.
func NthField(v reflect.Value, n int, tagKey string) (f reflect.Value, fd FieldDescriptor, ok bool) {
	if n < 0 || n >= NumField(v) {
		return
	}
	t := v.Type()
	k := v.Kind()
	if debug {
		fmt.Printf("\t%dth Field of %s\n", n, k.String())
	}
	if k == reflect.Struct {
		fd, ok = FieldDescriptorFromStructField(t.Field(n), tagKey)
		return v.Field(n), fd, ok
	}
	if k == reflect.Array || k == reflect.Slice {
		f = v.Index(n) // Will panic if OOR
		fd, ok := FieldDescriptorFromSlice(v, n, tagKey)
		return f, fd, ok
	}
	if k == reflect.Map {
		key := SortKeys(v.MapKeys())[n]
		// https://golang.org/ref/spec#Map_types
		// https://golang.org/ref/spec#Comparison_operators
		f = v.MapIndex(key)
		fd, ok = FieldDescriptorFromMap(v, key, tagKey)
		///// TODO
		return f, fd, ok
	}
	return
}

// SortKeys deterministically sorts the values used to key a map
// This allows map fields to be read numerically
func SortKeys(V []reflect.Value) []reflect.Value {
	values := Values(V)
	sort.Stable(values)
	return values
	// Note that Stable sorts require more operations to guarantee that equal fields remain in order
	// However, the intended use of this function should not allow for duplicate values, because
	// they are supposed to be keys of some map.
}

// Values represents an array of reflected values,
// whose fields can be itterated upon as one list
type Values []reflect.Value

// Len implements sorting
func (V Values) Len() int { return len(V) }

// Swap implements sorting
func (V Values) Swap(i, j int) { V[i], V[j] = V[j], V[i] }

// Less implements sorting, returns true if Vi < Vj AKA A < B
func (V Values) Less(i, j int) bool {
	A := KeyToBytes(V[i])
	B := KeyToBytes(V[j])
	for n := range A {
		if len(B) <= n {
			return false // A > B in length, and has until now been equal in value
		}
		if A[n] < B[n] {
			return true // Both have a byte at this index, and that of A is lesser
		}
		if A[n] > B[n] {
			return false
		}
	}
	// We have iterated over every field in A, and B has been equal so far in value
	return len(B) > len(A) // Either A is shorter (return true), or they are equal (false).
}

// KeyToBytes handles conversion from any reflect.Value which can be
// a map key into a []byte for ordering.
func KeyToBytes(v reflect.Value) []byte {
	// If you followed a stack trace to get here, you are now cursed.
	// Please implement support for a type to be freed.

	t := v.Type()
	k := v.Kind()
	if (k == reflect.Array || k == reflect.Slice) && t.Elem().Kind() == reflect.Uint8 {
		return v.Bytes()
	}
	if k == reflect.String {
		return []byte(v.String())
	}
	panic(fmt.Sprintf("Attempted to get bytes from unknown map key type %s.", t.Name()))
}

// TotalFields counts the fields in the collective list
func (V Values) TotalFields() (nf int) {
	for _, v := range V {
		nf += NumField(v)
	}
	return
}

// FieldAt returns the Nth field in the collective list
func (V Values) FieldAt(n int, tagKey string) (vfield reflect.Value, tfield FieldDescriptor, ok bool) {
	if n < 0 || n >= V.TotalFields() {
		ok = false

		return
	}
	for _, v := range V {
		nf := NumField(v)
		if n < nf {
			vfield, tfield, ok = NthField(v, n, tagKey)
			return
		}
		n -= nf
	}
	ok = false

	return
}

// ValueAt returns the Value whose field is Nth in the collective list
func (V Values) ValueAt(n int) (vfield reflect.Value, ok bool) {
	if n < 0 || n >= V.TotalFields() {
		ok = false
		return
	}
	for _, vfield = range V {
		nf := NumField(vfield)
		if n < nf {
			ok = true
			return
		}
		n -= nf
	}
	ok = false
	return

}

// mapType is a helper for getting the key and value types of a map
// ok is false if m is not a map.
func mapType(m reflect.Value) (k, v reflect.Type, ok bool) {
	if m.Kind() != reflect.Map {
		return
	}
	k = m.Type().Key()
	v = m.Type().Elem()
	ok = true
	return
}
