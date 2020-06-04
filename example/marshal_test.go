package example

import (
	"fmt"
	"testing"
	"time"

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

	t.Logf("March Marshaled data: %s", string(data))
}

func TestMarshalCustom(t *testing.T) {
	M := march.March{Tag: "March"}
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

	t.Logf("March Marshaled data: %s", string(data))

	m = Custom{}
	err = M.Unmarshal(data, &m)
	if err != nil {
		t.Fatalf("March Re-Unmarshal Error: %s", err.Error())
	}

	data, err = M.Marshal(&m)
	if err != nil {
		t.Fatalf("March Re-Marshal Error: %s", err.Error())
	}

	t.Logf("March Re-Marshaled data: %s", string(data))

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
			{"int":1, "nest":{"nest":0},"custom":null,"ptrs":null,"m1":null,"m2":[],"s":"","-":0},
			{"int":2, "nest":{"nest":0},"custom":null,"ptrs":null,"m1":null,"m2":[],"s":"","-":0}
		]`
		if match, err := CompareJSON(data, []byte(expect)); err != nil {
			t.Fatalf("Failed to compare JSON: %s\n\tGot: %s", err.Error(), string(data))
		} else if !match {
			t.Fatalf("Expected %s,\n\tgot %s", expect, string(data))
		}
	}
}

func TestMarshalComposite(t *testing.T) {
	M := march.March{Tag: "March", Strict: true}
	expect := "2020-02-02T01:02:03Z"
	testTime, err := time.Parse(time.RFC3339, expect)
	if err != nil {
		panic(err)
	}
	v := Composite{Time: testTime}
	expect = fmt.Sprintf(`{"time":"%s"}`, expect)

	data, err := M.Marshal(v)
	if err != nil {
		t.Fatalf("Unable to march unmarshal: %s", err.Error())
	} else if string(data) != expect {
		t.Fatalf("Error \n\t%s\n\tfrom march unmarshal, expected:\n\t%s", string(data), expect)
	}
	t.Logf("March marshaled time as expected: %s\n", string(data))

}
