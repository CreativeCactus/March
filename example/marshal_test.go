package example

import (
	"testing"

	march "github.com/CreativeCactus/March"
)

// Get a test value for the PtrS field.
func getPtrS(s string) **[]**string {
	ps := &s
	a := []**string{&ps}
	pa := &a
	return &pa
}

func TestMarshal(t *testing.T) {
	M := march.March{Tag: "March"}
	m := T{
		Embed: Embed{
			Embedded: 10,
		},
		Nest: Nested{
			Nested: 10,
		},
		Int:  10,
		PtrS: getPtrS("test"),
		M1: map[string]int32{
			"a": 123,
		},
		M2: []map[string]interface{}{
			{
				"a": 123,
			},
		},
		S1:    "test1",
		S2:    "test2",
		priv:  "private",
		None1: 1,
		None2: 1,
		None3: 1,
		None4: 1,
		None5: 1,
	}

	data, err := M.Marshal(m)
	if err != nil {
		t.Fatalf("March Marshal Error: %s", err.Error())
	}

	t.Logf("March Marshalled data: %s", string(data))
}

func TestMarshalCustom(t *testing.T) {
	M := march.March{Tag: "March", Debug: true}
	m := Custom{
		Custom: 3,
		Nested: []Custom{
			{
				Custom: 1,
				Nested: []Custom{
					{
						Custom: 2,
					}, {
						Custom: 2,
					},
				},
			},
		},
	}

	data, err := M.Marshal(&m)
	if err != nil {
		t.Fatalf("March Marshal Error: %s", err.Error())
	}

	t.Logf("March Marshalled data: %s", string(data))

	m = Custom{}
	err = M.Unmarshal(data, &m)
	if err != nil {
		t.Fatalf("March Re-Unmarshal Error: %s", err.Error())
	}

	data, err = M.Marshal(&m)
	if err != nil {
		t.Fatalf("March Re-Marshal Error: %s", err.Error())
	}

	t.Logf("March Re-Marshalled data: %s", string(data))

}

func TestMarshalSlice(t *testing.T) {
	M := march.March{Tag: "March", Verbose: true}
	v := []T{{Int: 1}, {Int: 2}}
	{
		data, err := M.Marshal(v)
		// Compare...
		if err != nil {
			t.Fatalf("Error from march marshal:\n\t%s", err.Error())
		}
		expect := `[
			{"int":1, "nest":{"nest":0},"custom":{"custom":0,"nested":null},"ptrs":null,"m1":null,"m2":null,"s":"","-":0},
			{"int":2, "nest":{"nest":0},"custom":{"custom":0,"nested":null},"ptrs":null,"m1":null,"m2":null,"s":"","-":0}
		]`
		if match, err := CompareJSON(data, []byte(expect)); err != nil {
			t.Fatalf("Failed to compare JSON: %s\n\tGot: %s", err.Error(), string(data))
		} else if !match {
			t.Fatalf("Expected %s,\n\tgot %s", expect, string(data))
		}
	}
}
