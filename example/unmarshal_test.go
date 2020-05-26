package example

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	march "github.com/CreativeCactus/March"
)

func TestUnmarshalStrictFail(t *testing.T) {
	M := march.March{Tag: "March"}
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

	// Test...
	m := T{}
	merr := M.Unmarshal([]byte(data), &m)
	if merr == nil {
		t.Fatalf("No error from march unmarshal, but expected one")
	}

	// Compare...
	j := T{}
	jerr := json.Unmarshal([]byte(data), &j)
	if jerr == nil {
		t.Fatalf("No error from json unmarshal, but expected one")
	}
	if jerr.Error() != merr.Error() {
		t.Logf("\tMARCH: %s\n", merr.Error())
		t.Logf("\tJSON : %s\n", jerr.Error())
		t.Logf("Unmarshal error does not match json, do they look similar?")
		return
	}

	t.Logf("Error matches that of json")
}

func TestUnmarshalPass(t *testing.T) {
	M := march.March{Tag: "March"}
	data := `{
		"nest":{"nest":4},
		"custom": "abc",
		"int": 1,
		"ptrs":["~"],
		"m1":{"b":123,"c":99999, "d": null},
		"m2": [{"x":"y"}],
		"s1": "A",
		"s2":null,
		"NotOnStruct":9
	}`

	// Test...
	v := T{}
	err := M.Unmarshal([]byte(data), &v)
	if err != nil {
		t.Fatalf("Unable to march unmarshal: %s", err.Error())
	}
	s, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("Unable to marshal the result: %s", err.Error())
	}
	t.Logf("March: %s\n", s)

	// Compare...
	x := T{}
	err = json.Unmarshal([]byte(data), &x)
	if err != nil {
		t.Fatalf("Unable to json unmarshal: %s", err.Error())
	}
	js, err := json.MarshalIndent(x, "", "  ")
	if err != nil {
		t.Fatalf("Unable to marshal the JSON result: %s", err.Error())
	}
	if !bytes.Equal(js, s) {
		t.Logf("JSON: %s\n", js)
		t.Logf("Note: JSON returned %d bytes, March returned %d", len(js), len(s))
		t.Fatalf("Unmarshalled string does not match json")
	}
}

func TestUnmarshalStrictPassNoEmbedded(t *testing.T) {
	M := march.March{Tag: "March"}
	data := `{ "embeded": 3, "deep": 4 }`

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
	js := []byte(`{"deep":0,"embeded":0}`)
	if !bytes.Equal(js, s) {
		t.Logf("JSON: %s\n", js)
		t.Logf("Note: JSON returned %d bytes, March returned %d", len(js), len(s))
		t.Fatalf("Unmarshalled string does not match expectation")
	}
}

func TestUnmarshalRelaxPass(t *testing.T) {
	M := march.March{Tag: "March", Verbose: true}
	data := `{
		"custom": "asd",
		"int": 1,
		"ptrs":["~"],
		"m1":{"b":123,"c":99, "d": null},
		"m2": [{"x":"y"}],
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
		t.Fatalf("Unmarshalled string does not match json")
	}
}

func _TestUnmarshalMap(t *testing.T) {
	M := march.March{Tag: "March", Verbose: true}
	data := `{
		"deep": 4,
		"embeded":3,
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
		t.Fatalf("Unmarshalled string does not match json")
	}
}

func TestUnmarshalMapFail(t *testing.T) {
	M := march.March{Tag: "March", Verbose: true}
	data := `{}`
	// Test...
	v := map[string][]byte{}
	err := M.Unmarshal([]byte(data), &v)
	if err == nil {
		t.Fatalf("No error from march unmarshal, but expected one")
	}
	// Compare...
	expect := `Default unmarshaller does not support non-struct types yet. Implement UnmarshalMarch or use a struct`
	if err.Error() != expect {
		t.Fatalf("Unexpected error from march unmarshal: %s", err.Error())
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
	data := `{...}`

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
