package example

import (
	"encoding/json"
	"testing"

	march "github.com/CreativeCactus/March"
)

type Flags struct {
	Value   rune                       `March:"V"`
	Hoisted U                          `March:"H,hoist"`
	Remains map[string]json.RawMessage `March:"R,remains"`
}

type U struct {
	Hoistable int `March:"H2"`
}

type Both struct {
	Value rune                       `March:"V"`
	Both  map[string]json.RawMessage `March:"B,hoist,remains"`
}

func TestRemains(t *testing.T) {
	M := march.March{Tag: "March"}
	data := `{ "V":128, "R":"actually remains", "Z":"also remains", "H2": "but also this" }`
	v := Flags{}

	err := M.Unmarshal([]byte(data), &v)
	if err != nil {
		t.Fatalf("March Unmarshal Error: %s", err.Error())
	}

	if want := rune(128); v.Value != want {
		t.Fatalf("Value mismatch: Got %d, Want %d", v.Value, want)
	}
	if want := 0; v.Hoisted.Hoistable != want {
		t.Fatalf("Value mismatch: Got %d, Want %d", v.Hoisted.Hoistable, want)
	}
	if want := 3; len(v.Remains) != want {
		t.Fatalf("Value mismatch: Got %d, Want %d", len(v.Remains), want)
	}
	if want := `"actually remains"`; string(v.Remains["R"]) != want {
		t.Fatalf("Value mismatch: Got %s, Want %s", string(v.Remains["R"]), want)
	}

	t.Logf("March Unmarshaled: %#v\n", v)
}

func TestHoist(t *testing.T) {
	M := march.March{Tag: "March"}
	v := Flags{
		Value: 'a',
		Hoisted: U{
			Hoistable: 10,
		},
		Remains: map[string]json.RawMessage{
			"test": []byte(`"test"`),
		},
	}

	data, err := M.Marshal(&v)
	if err != nil {
		t.Fatalf("March Unmarshal Error: %s", err.Error())
	}

	want := `{"V":97,"R":{"test":"test"},"H2":10}`

	if match, err := CompareJSON(data, []byte(want)); err != nil {
		t.Fatalf("Failed to compare marshaled JSON: %s", err.Error())
	} else if !match {
		t.Fatalf("Value mismatch: Got %s, Want %s", string(data), want)
	}

	t.Logf("March Marshaled: %#v\n", data)
}

func TestBoth(t *testing.T) {
	M := march.March{Tag: "March"}
	result := `{"V":128,"Z":"remains","ZZ":"remainZ"}`
	data := []byte(result)
	v := Both{}

	err := M.Unmarshal(data, &v)
	if err != nil {
		t.Fatalf("March Unmarshal Error: %s", err.Error())
	}

	if want := rune(128); v.Value != want {
		t.Fatalf("Value mismatch: Got %d, Want %d", v.Value, want)
	}
	if want := 2; len(v.Both) != want {
		t.Fatalf("Value mismatch: Got %d, Want %d", len(v.Both), want)
	}
	if want := `"remains"`; string(v.Both["Z"]) != want {
		t.Fatalf("Value mismatch: Got %s, Want %s", string(v.Both["Z"]), want)
	}
	t.Logf("March Unmarshaled: %#v\n", v)

	data, err = M.Marshal(&v)
	if err != nil {
		t.Fatalf("March Unmarshal Error: %s", err.Error())
	}

	if match, err := CompareJSON(data, []byte(result)); err != nil {
		t.Fatalf("Failed to compare JSON: %s", err.Error())
	} else if !match {
		t.Fatalf("Value mismatch: Got %s, Want %s", string(data), result)
	}

	t.Logf("March Marshaled: %#v\n", data)
}

func TestIdempotent(t *testing.T) {
	M := march.March{Tag: "March"}
	result := `{"V":128,"Z":"remains"}`
	data := []byte(result)
	v := Both{}

	// First unmarshal
	{
		err := M.Unmarshal(data, &v)
		if err != nil {
			t.Fatalf("March Unmarshal Error: %s", err.Error())
		}
		if want := rune(128); v.Value != want {
			t.Fatalf("Value mismatch: Got %d, Want %d", v.Value, want)
		}
		if want := 1; len(v.Both) != want {
			t.Fatalf("Value mismatch: Got %d, Want %d", len(v.Both), want)
		}
		if want := `"remains"`; string(v.Both["Z"]) != want {
			t.Fatalf("Value mismatch: Got %s, Want %s", string(v.Both["R"]), want)
		}
		t.Logf("March Unmarshaled:\n\t%#v\n", v)
	}

	// First marshal
	{
		var err error
		data, err = M.Marshal(&v)
		if err != nil {
			t.Fatalf("March Marshal Error: %s", err.Error())
		}

		if match, err := CompareJSON(data, []byte(result)); err != nil {
			t.Fatalf("Failed to compare marshaled JSON: %s", err.Error())
		} else if !match {
			t.Fatalf("Value mismatch: Got %s, Want %s", string(data), result)
		}

		t.Logf("March Marshaled:\n\t%#v\n", data)
	}
	// Second unmarshal
	{
		err := M.Unmarshal(data, &v)
		if err != nil {
			t.Fatalf("March Unmarshal Error: %s", err.Error())
		}
		if want := rune(128); v.Value != want {
			t.Fatalf("Value mismatch: Got %d, Want %d", v.Value, want)
		}
		if want := 1; len(v.Both) != want {
			t.Fatalf("Value mismatch: Got %d, Want %d", len(v.Both), want)
		}
		if want := `"remains"`; string(v.Both["Z"]) != want {
			t.Fatalf("Value mismatch: Got %s, Want %s", string(v.Both["R"]), want)
		}
		t.Logf("March Unmarshaled:\n\t%#v\n", v)
	}

	// Second marshal
	{
		var err error
		data, err = M.Marshal(&v)
		if err != nil {
			t.Fatalf("March Marshal Error: %s", err.Error())
		}

		if match, err := CompareJSON(data, []byte(result)); err != nil {
			t.Fatalf("Failed to compare marshaled JSON: %s", err.Error())
		} else if !match {
			t.Fatalf("Value mismatch: Got %s, Want %s", string(data), result)
		}

		t.Logf("March Marshaled:\n\t%#v\n", data)
	}

}
