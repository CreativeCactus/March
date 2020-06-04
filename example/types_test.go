package example

import (
	"encoding/json"
	"reflect"
	"time"
)

// TODO Test with more nulls
// 	Raw    json.RawMessage          `March:"raw" json:"raw"`

type T struct {
	Embed
	Nest   Nested                   `March:"nest" json:"nest"`
	Custom *Custom                  `March:"custom" json:"custom"`
	Int    int                      `March:"int" json:"int"`
	PtrS   **[]**string             `March:"ptrs" json:"ptrs"`
	M1     map[string]int32         `March:"m1" json:"m1"`
	M2     []map[string]interface{} `March:"m2" json:"m2"`
	S1     string                   `March:"s" json:"s"`
	S2     string                   `March:"s" json:"s"`
	// Inaccessible
	priv  string `March:"int" json:"int"`
	None1 int    `March:"-" json:"-"`
	None2 int    `March:"" json:""`
	None3 int    `March: json:`
	None4 int    `March json`
	None5 int
}

type Composite struct {
	Time time.Time `March:"time"`
}

type Simple struct {
	Embed
}
type Embed struct {
	Embed2
	Embedded int32 `March:"embedded" json:"embedded"`
}
type Embed2 struct {
	Deep int32 `March:"deep" json:"deep"`
}
type Nested struct {
	Nested int16 `March:"nest" json:"nest"`
}

// CompareJSON is useful for checking the equality of json strings
// when go might force an unpredictable field ordering.
func CompareJSON(a, b []byte) (match bool, err error) {
	var A interface{}
	var B interface{}
	if err = json.Unmarshal(a, &A); err != nil {
		return
	}
	if err = json.Unmarshal(b, &B); err != nil {
		return
	}
	match = reflect.DeepEqual(A, B)
	return
}

// union assigns fields from b to a and returns the result
// without mutating a or b.
func union(a, b map[string]json.RawMessage) (c map[string]json.RawMessage) {
	c = map[string]json.RawMessage{}
	for k, v := range a {
		c[k] = v
	}
	for k, v := range b {
		c[k] = v
	}
	return
}
