package example

import (
	"fmt"
	"testing"

	march "github.com/CreativeCactus/March"
)

type X struct {
	Value  march.RawUnmarshal            `abc:"v"`
	Extras map[string]march.RawUnmarshal `abc:"_,hoist,remains"`
}

func TestRuntimeReadRemains(t *testing.T) {
	M := march.March{Tag: "abc"}
	x := X{}
	data := `{"v":200,"string":"test","int":2, "map":{"array":[3]}}`

	if err := M.Unmarshal([]byte(data), &x); err != nil {
		t.Fatalf("March Unmarshal: %s", err.Error())
	}

	{ // Lazy unmarshal top level int
		n := 0
		if err := x.Value.UnmarshalTo(&n); err != nil {
			t.Fatalf("Lazy Unmarshal: %s", err.Error())
		} else if want := 200; n != want {
			t.Fatalf("Lazy Unmarshal: %d, wanted %d", n, want)
		}
		t.Logf("Lazy Unmarshaled %d\n", n)
	}
	{ // Lazy unmarshal a string
		s := ""
		if err := x.Extras["string"].UnmarshalTo(&s); err != nil {
			t.Fatalf("Lazy Unmarshal: %s", err.Error())
		} else if want := "test"; s != want {
			t.Fatalf("Lazy Unmarshal: %s, wanted %s", s, want)
		}
		t.Logf("Lazy Unmarshaled %s\n", s)
	}
	{ // Lazy unmarshal an int
		n := 0
		if err := x.Extras["int"].UnmarshalTo(&n); err != nil {
			t.Fatalf("Lazy Unmarshal: %s", err.Error())
		} else if want := 2; n != want {
			t.Fatalf("Lazy Unmarshal: %d, wanted %d", n, want)
		}
		t.Logf("Lazy Unmarshaled %d\n", n)
	}
	{ // Lazy unmarshal a complex type
		m := map[string][]int{}
		if err := x.Extras["map"].UnmarshalTo(&m); err != nil {
			t.Fatalf("Lazy Unmarshal: %s", err.Error())
		}
		got := fmt.Sprintf("%#v", m)
		if want := `map[string][]int{"array":[]int{3}}`; got != want {
			t.Fatalf("Lazy Unmarshal: %s, wanted %s", got, want)
		}
		t.Logf("Lazy Unmarshaled %#v\n", m)
	}
	{ // Lazy unmarshal invalid type
		n := 0
		expect := `json: cannot unmarshal string into Go value of type int`
		err := x.Extras["string"].UnmarshalTo(&n)
		if err == nil {
			t.Fatalf("Expected error(%s), got nil", expect)
		}
		if got := err.Error(); got != expect {
			t.Fatalf("Expected error(%s), got %s", expect, got)
		}
		t.Logf("Lazy Unmarshaled %d\n", n)
		t.Logf("Lazy Unmarshal got expected error %s\n", expect)
	}
}
