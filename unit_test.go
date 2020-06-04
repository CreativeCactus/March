package march

import (
	"reflect"
	"testing"
)

func TestUnmarshalPtr(t *testing.T) {
	target := ""
	v := &target
	m := March{}
	{ // Direct call
		t.Logf("Direct...")
		{ // Quoted
			V := reflect.ValueOf(v)
			T := reflect.TypeOf(v)
			if err := m.unmarshalJSONPtr(T, V, []byte(`"string1"`)); err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}
		}
		if want := "string1"; target != want {
			t.Fatalf("Got %s, expected %s.", target, want)
		}
		t.Logf("Direct calls pass: %s", target)
	}

	{ // Via March
		t.Logf("March...")
		{ // Quoted
			if err := m.Unmarshal([]byte(`"string3"`), v); err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}
		}
		if want := "string3"; target != want {
			t.Fatalf("Got %s, expected %s.", target, want)
		}
		t.Logf("March calls pass: %s", target)
	}
}

func TestUnmarshalPtrPtr(t *testing.T) {
	target := ""
	vv := &target
	v := &vv
	m := March{}
	{ // Direct call
		t.Logf("Direct...")
		{ // Quoted
			V := reflect.ValueOf(v)
			T := reflect.TypeOf(v)
			if err := m.unmarshalJSONPtr(T, V, []byte(`"string1"`)); err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}
		}
		if want := "string1"; target != want {
			t.Fatalf("Got %s, expected %s.", target, want)
		}
		t.Logf("Direct calls pass: %s", target)
	}

	{ // Via March
		t.Logf("March...")
		{ // Quoted
			if err := m.Unmarshal([]byte(`"string3"`), v); err != nil {
				t.Fatalf("Unexpected error: %s", err.Error())
			}
		}
		if want := "string3"; target != want {
			t.Fatalf("Got %s, expected %s.", target, want)
		}
		t.Logf("March calls pass: %s", target)
	}
}
func TestUnmarshalPtrUninitialized(t *testing.T) {
	m := March{}
	{ // &***string
		var vv ***string
		v := &vv
		{ // Direct call
			t.Logf("Direct...")
			{ // Quoted
				V := reflect.ValueOf(v)
				T := reflect.TypeOf(v)
				if err := m.unmarshalJSONPtr(T, V, []byte(`"string1"`)); err != nil {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
			}
			if want := "string1"; ****v != want {
				t.Fatalf("Got %s, expected %s.", ****v, want)
			}
			t.Logf("Direct calls pass: %s", ****v)
		}

		{ // Via March
			t.Logf("March...")
			{ // Quoted
				if err := m.Unmarshal([]byte(`"string3"`), v); err != nil {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
			}
			if want := "string3"; ****v != want {
				t.Fatalf("Got %s, expected %s.", ****v, want)
			}
			t.Logf("March calls pass: %s", ****v)
		}
	}
	{ // &&***string
		var vvv ***string
		vv := &vvv
		v := &vv
		{ // Direct call
			t.Logf("Direct...")
			{ // Quoted
				V := reflect.ValueOf(v)
				T := reflect.TypeOf(v)
				if err := m.unmarshalJSONPtr(T, V, []byte(`"string4"`)); err != nil {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
			}
			if want := "string4"; *****v != want {
				t.Fatalf("Got %s, expected %s.", *****v, want)
			}
			t.Logf("Direct calls pass: %s", *****v)
		}

		{ // Via March
			t.Logf("March...")
			{ // Quoted
				if err := m.Unmarshal([]byte(`"string5"`), v); err != nil {
					t.Fatalf("Unexpected error: %s", err.Error())
				}
			}
			if want := "string5"; *****v != want {
				t.Fatalf("Got %s, expected %s.", *****v, want)
			}
			t.Logf("March calls pass: %s", *****v)
		}
	}
}
