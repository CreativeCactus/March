package example

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	march "github.com/CreativeCactus/March"
)

func TestUnmarshalStrictFail(t *testing.T) {
	M := march.March{Tag: "March", Strict: true}
	data := `{
		"embed":3,
		"nest":{"nest":4},
		"custom": "abc",
		"int": 1,
		"ptrs":["~"],
		"m1":{"a":"3","b":123,"c":999999999999999999999999999999999999999999999999999, "d": null},
		"m2": [{"x":"y"}],
		"s1": "A",
		"s2":null,
		"NotOnStruct":9
	}`

	// Example error from JSON...
	j := T{}
	jerr := json.Unmarshal([]byte(data), &j)
	if jerr == nil {
		t.Fatalf("No error from json unmarshal, but expected one")
	}
	expect := strings.Replace(jerr.Error(), "struct field T.m1", "value", 1)

	// Test...
	m := T{}
	merr := M.Unmarshal([]byte(data), &m)
	if merr == nil {
		t.Fatalf("No error from march unmarshal, but expected something like:\n\t%s", expect)
	}

	// Compare...
	if merr.Error() != expect {
		t.Logf("\tMARCH: %s\n", merr.Error())
		t.Logf("\tJSON : %s\n", jerr.Error())
		t.Logf("Unmarshal error does not match json, do they look similar?")
		return
	}

	t.Logf("Error closely resembles that of json")
}

func TestUnmarshal(t *testing.T) {
	M := march.March{Tag: "March"}
	data := `{
		"nest":{"nest":4},
		"custom": "a[]",
		"int": 1,
		"ptrs":["~"],
		"m1":{"b":123,"c":99999, "d": null},
		"m2": [{"x":"y"}],
		"s": "A",
		"NotOnStruct":9
	}`

	// Test...
	v := T{}
	err := M.Unmarshal([]byte(data), &v)
	if err != nil {
		t.Fatalf("Unable to march unmarshal: %s", err.Error())
	}
	s, err := M.Marshal(v)
	if err != nil {
		t.Fatalf("Unable to marshal the result: %s", err.Error())
	}
	t.Logf("March: %s\n", s)

	// Compare... Note that the last line is the only change
	// There is no NotOnStruct and - is marshaled
	// Also, .m2.d is represented as 0, not null
	js := []byte(`{
		"nest":{"nest":4},
		"custom": "a[]",
		"int": 1,
		"ptrs":["~"],
		"m1":{"b":123,"c":99999, "d": 0},
		"m2": [{"x":"y"}],
		"s": "A",
		"-":0
	}`)
	if match, err := CompareJSON(js, s); err != nil {
		t.Fatalf("Failed to compare marshaled JSON: %s", err.Error())
	} else if !match {
		t.Logf("JSON: %s\n", js)
		t.Logf("Note: JSON returned %d bytes, March returned %d", len(js), len(s))
		t.Fatalf("Unmarshaled string does not match JSON")
	}
}

// TestUnmarshalStrictNoEmbedded confirms that embedded fields CANNOT
// be unmarshaled onto.
func TestUnmarshalStrictNoEmbedded(t *testing.T) {
	M := march.March{Tag: "March"}
	data := `{ "embedded": 3, "deep": 4 }`

	// Test...
	v := Simple{}
	err := M.Unmarshal([]byte(data), &v)
	if err != nil {
		t.Fatalf("Unable to march unmarshal: %s", err.Error())
	}
	s, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Unable to marshal the result: %s", err.Error())
	}
	t.Logf("March: %s\n", s)

	// Compare...
	js := []byte(`{"deep":0,"embedded":0}`)
	if match, err := CompareJSON(js, s); err != nil {
		t.Fatalf("Failed to compare JSON: %s", err.Error())

	} else if !match {
		t.Logf("JSON: %s\n", js)
		t.Logf("Note: JSON returned %d bytes, March returned %d", len(js), len(s))
		t.Fatalf("Unmarshaled string does not match expectation")
	}
}

func TestUnmarshalRelax(t *testing.T) {
	M := march.March{Tag: "March", Verbose: true}
	data := `{
		"custom": "asd",
		"int": 1,
		"ptrs":["~"],
		"m1":{"b":123,"c":99, "d": null},
		"m2": [{"x":"y"}],
		"s": "both",
		"NotOnStruct":9
	}` // m1 fails with overflows and strings
	// Does not work with nested fields, nor embedded

	// Test...
	v := T{}
	err := M.Unmarshal([]byte(data), &v)
	if err != nil {
		t.Fatalf("Unable to march unmarshal: %s", err.Error())
	}
	s, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("Unable to marshal the result: %s", err.Error())
	}
	t.Logf("March: %s\n", s)
	// s := []byte("")

	// Compare...
	j := T{}
	err = json.Unmarshal([]byte(data), &j)
	if err != nil {
		t.Fatalf("Unable to marshal JSON: %s", err.Error())
	}
	js, err := json.Marshal(j)
	if err != nil {
		t.Fatalf("Unable to marshal the JSON result: %s", err.Error())
	}
	if !bytes.Equal([]byte(js), s) {
		t.Logf("JSON: %s\n", js)
		t.Fatalf("Unmarshaled string does not match json")
	}
}

func _TestUnmarshalMap(t *testing.T) {
	M := march.March{Tag: "March", Verbose: true}
	data := `{
		"deep": 4,
		"embedded":3,
		"nest":{"nest":4},
		"custom": "asd",
		"int": 1,
		"ptrs":["~"],
		"m1":{"a":"3","b":123,"c":999999999999999999999999999999999999999999999999999, "d": null},
		"m2": [{"x":"y"}],
		"NotOnStruct":9
	}`

	// Test...
	v := map[string][]byte{}
	err := M.Unmarshal([]byte(data), &v)
	if err != nil {
		t.Fatalf("Unable to march unmarshal: %s", err.Error())
	}
	s, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("Unable to marshal the result: %s", err.Error())
	}
	t.Logf("March: %s\n", s)
	// s := []byte("")

	// Compare...
	js := `NOJSON`
	if !bytes.Equal([]byte(js), s) {
		t.Logf("JSON: %s\n", js)
		t.Fatalf("Unmarshaled string does not match json")
	}
}

func TestUnmarshalArrayFail(t *testing.T) {
	M := march.March{Tag: "March", Verbose: true}
	data := `[]`
	// Test...
	var v [2]string
	err := M.Unmarshal([]byte(data), &v)
	// Compare...
	expect := `Default JSON unmarshaler does not support array types. Use a slice`
	if err == nil {
		t.Fatalf("No error from march unmarshal, expected:\n\t%s", expect)
	} else if got := err.Error(); got != expect {
		t.Fatalf("Error \n\t%s\n\tfrom march unmarshal, expected:\n\t%s", got, expect)
	}
}

func TestUnmarshalSlice(t *testing.T) {
	M := march.March{Tag: "March", Verbose: true}
	data := `[{"int":1},{"int":2}]`
	{ // Test []*
		v := []*T{}
		err := M.Unmarshal([]byte(data), &v)
		// Compare...
		if err != nil {
			t.Fatalf("Error from march unmarshal:\n\t%s", err.Error())
		}
		if got := len(v); got != 2 {
			t.Fatalf("Expected %d elements, got %d", 2, got)
		}
		if got := v[0].Int; got != 1 {
			t.Fatalf("Expected first element: %d, got %d", 1, got)
		}
	}
	{ // Test []
		v := []T{}
		err := M.Unmarshal([]byte(data), &v)
		// Compare...
		if err != nil {
			t.Fatalf("Error from march unmarshal:\n\t%s", err.Error())
		}
		if got := len(v); got != 2 {
			t.Fatalf("Expected %d elements, got %d", 2, got)
		}
		if got := v[0].Int; got != 1 {
			t.Fatalf("Expected first element: %d, got %d", 1, got)
		}
	}
	{ // Test *[]
		vv := []T{}
		v := &vv
		err := M.Unmarshal([]byte(data), &v)
		// Compare...
		if err != nil {
			t.Fatalf("Error from march unmarshal:\n\t%s", err.Error())
		}
		if got := len(*v); got != 2 {
			t.Fatalf("Expected %d elements, got %d", 2, got)
		}
		if got := (*v)[0].Int; got != 1 {
			t.Fatalf("Expected first element: %d, got %d", 1, got)
		}
	}
}

type CustomReadFields struct {
	A byte `March:"0"`
	B byte `March:"2" json:"b"`
}

func (CustomReadFields) ReadFieldsMarch(data []byte) (fields map[string][]byte, err error) {
	// A strange ReadFields which always interprets the incoming data as a map of bytes
	fields = map[string][]byte{}
	for i, b := range data {
		fields[fmt.Sprintf("%d", i)] = []byte(fmt.Sprintf(`%d`, b))
	}
	return
}

func TestUnmarshalReadFields(t *testing.T) {
	M := march.March{Tag: "March", Verbose: true}
	data := `{...}` // CustomReadFields.ReadFieldsMarch will interpret this as 5 single-byte fields

	// Test...
	v := CustomReadFields{}
	err := M.Unmarshal([]byte(data), &v)
	if err != nil {
		t.Fatalf("Unable to march unmarshal: %s", err.Error())
	}

	t.Logf("March: %#v\n", v)
}

type UnmarshalMap map[string]int

func (um UnmarshalMap) UnmarshalMarch(data []byte) (err error) {
	// A strange Unmarshal which always assigns the number of bytes it was given
	um["a"] = len(data)
	return
}

func TestUnmarshalMap(t *testing.T) {
	M := march.March{Tag: "March", Verbose: true}
	data := `{}`

	// Test...
	v := UnmarshalMap{}
	err := M.Unmarshal([]byte(data), &v)
	if err != nil {
		t.Fatalf("Unable to march unmarshal: %s", err.Error())
	}
	t.Logf("March: %#v\n", v)
}

func TestUnmarshalComposite(t *testing.T) {
	M := march.March{Tag: "March", Strict: true}
	v := Composite{}

	{ // TODO Fix this double quoting issue
		data := `{"time":"2020-02-02T06:06:60+25:00"}`
		err := M.Unmarshal([]byte(data), &v)
		expect := `parsing time ""2020-02-02T06:06:60+25:00"": second out of range`
		if err == nil {
			t.Fatalf("No error from march unmarshal, expected:\n\t%s", expect)
		} else if got := err.Error(); got != expect {
			t.Fatalf("Error \n\t%s\n\tfrom march unmarshal, expected:\n\t%s", got, expect)
		}
		t.Logf("March error as expected: %s\n", err.Error())
	}

	{
		data := `{"time":"2020-02-02T06:06:06+00:00"}`
		err := M.Unmarshal([]byte(data), &v)
		got := (v.Time.String())
		expect := "2020-02-02 06:06:06 +0000 +0000"
		if err != nil {
			t.Fatalf("Unable to march unmarshal: %s", err.Error())
		} else if got != expect {
			t.Fatalf("Error \n\t%s\n\tfrom march unmarshal, expected:\n\t%s", got, expect)
		}
		t.Logf("March parsed time as expected: %s\n", got)
	}
}
