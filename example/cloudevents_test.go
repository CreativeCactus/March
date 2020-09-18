package example

import (
	"encoding/json"
	"testing"
	"time"

	march "github.com/CreativeCactus/March"
)

type CloudEvent struct {
	SpecVersion     string                        `ce:"specversion"`
	Type            string                        `ce:"type"`
	Source          string                        `ce:"source"`
	ID              string                        `ce:"id"`
	Time            time.Time                     `ce:"time"`
	Data            march.RawUnmarshal            `ce:"data"`
	DataContentType string                        `ce:"datacontenttype"`
	Extensions      map[string]march.RawUnmarshal `ce:"_,hoist,remains"`
}

// TestCloudEvents shows the use of the same test for a cloudevent
// with literal JSON and a JSON string for .data
func TestCloudEvents(t *testing.T) {
	m := march.March{Tag: "ce"}
	UnmarshalRemarshalCloudEvent(m, t, applicationJSON)
	UnmarshalRemarshalCloudEvent(m, t, textJSON)
}

func UnmarshalRemarshalCloudEvent(m march.March, t *testing.T, example string) {
	ce := CloudEvent{}

	{ // Can unmarshal base struct
		if err := m.Unmarshal([]byte(example), &ce); err != nil {
			t.Fatalf("Unmarshal Error: %s", err.Error())
		}
	}

	{ // Can lazy unmarshal string to string
		s := ""
		if err := ce.Extensions["comexampleextension1"].UnmarshalTo(&s); err != nil {
			t.Fatalf("UnmarshalTo Error: %s", err.Error())
		}
	}

	{ // Can lazy unmarshal number to json.RawMessage
		b := json.RawMessage{}
		if err := ce.Extensions["comexampleothervalue"].UnmarshalTo(&b); err != nil {
			t.Fatalf("UnmarshalTo Error: %s", err.Error())
		}
	}

	{ // Can lazy unmarshal data
		{ // To RawMessage
			b := json.RawMessage{}
			if err := ce.Data.UnmarshalTo(&b); err != nil {
				t.Fatalf("UnmarshalTo Error: %s", err.Error())
			}
		}

		// Note that data is used as a receiving interface later, so type matters
		data := string(ce.Data.Bytes)
		jt, err := ce.Data.TypeOfJSON()
		if err != nil {
			t.Fatalf("Failed to determine data type: %s", err.Error())
		}

		t.Logf("ce.Data has type: %s", string(jt))
		if jt == march.JSONString {
			if err := ce.Data.UnmarshalTo(&data); err != nil {
				t.Fatalf("UnmarshalTo Error: %s", err.Error())
			}
		}

		if jt == march.JSONObject {
			{ // To Struct, using UnmarshalTo
				b := struct {
					A int64 `ce:"a"`
				}{}
				if err := ce.Data.UnmarshalTo(&b); err != nil {
					t.Fatalf("UnmarshalTo Error: %s", err.Error())
				}
			}
		}

		{ // To Struct
			b := struct {
				A int64 `ce:"a"`
			}{}
			if err := m.Unmarshal([]byte(data), &b); err != nil {
				t.Fatalf("Unmarshal Error: %s", err.Error())
			}
		}
		{ // To Struct, using a different March
			b := struct {
				A int64 `March:"a"`
			}{}
			if err := (march.March{}).Unmarshal([]byte(data), &b); err != nil {
				t.Fatalf("Unmarshal Error: %s", err.Error())
			}
		}
	}

	{ // Correct values
		if want := "A234-1234-1234"; want != ce.ID {
			t.Fatalf("ID: Want %s Got %s", want, ce.ID)
		}
		if want := "/mycontext"; want != ce.Source {
			t.Fatalf("Source: Want %s Got %s", want, ce.Source)
		}
		if want := "com.example.someevent"; want != ce.Type {
			t.Fatalf("Type: Want %s Got %s", want, ce.Type)
		}
	}

	{ // Can re-marshal idempotently
		if data, err := m.Marshal(&ce); err != nil {
			t.Fatalf("UnmarshalTo Error: %s", err.Error())
		} else {
			t.Logf("Remarshaled: %s", string(data))
			if match, err := CompareJSON(data, []byte(example)); err != nil {
				t.Fatalf("Failed to compare marshaled JSON: %s", err.Error())
			} else if !match {
				t.Fatalf("Value mismatch: Got %s, Want %s", string(data), example)
			}
		}
		t.Logf("Remarshaled idempotently")
	}
}

// examples adapted from the v1.0 CloudEvents spec:
// https://github.com/cloudevents/spec/blob/v1.0/json-format.md#32-examples

const applicationJSON = `{
    "specversion" : "1.0",
    "type" : "com.example.someevent",
    "source" : "/mycontext",
    "id" : "A234-1234-1234",
    "time" : "2018-04-05T17:31:00Z",
    "comexampleextension1" : "value",
    "comexampleothervalue" : 5,
    "datacontenttype" : "application/json",
    "data" : {
		"a": 1
	}
}`

const textJSON = `{
    "specversion" : "1.0",
    "type" : "com.example.someevent",
    "source" : "/mycontext",
    "id" : "A234-1234-1234",
    "time" : "2018-04-05T17:31:00Z",
    "comexampleextension1" : "value",
    "comexampleothervalue" : 5,
    "datacontenttype" : "text/json",
    "data" : "{\"a\": 1}"
}`
