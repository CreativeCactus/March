package march

import (
	"fmt"
	"strconv"
)

// TODO Test accross import boundaries by moving tests out
// TODO Test with more nulls
// 	Raw    json.RawMessage          `March:"raw" json:"raw"`

type T struct {
	Embed
	Nest   Nested                   `March:"nest" json:"nest"`
	Custom Custom                   `March:"custom" json:"custom"`
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

type Simple struct {
	Embed
}
type Embed struct {
	Embed2
	Embeded int32 `March:"embeded" json:"embeded"`
}
type Embed2 struct {
	Deep int32 `March:"deep" json:"deep"`
}
type Nested struct {
	Nested int16 `March:"nest" json:"nest"`
}
type Custom struct {
	Custom int      `March:"custom" json:"custom"`
	Nested []Custom `March:"nested" json:"nested"`
}

func numeric(b byte) bool {
	_, err := strconv.Atoi(string(b))
	return err == nil
}
func (c *Custom) UnmarshalMarch(data []byte) error {
	// Some obscure unmarshal function
	depth := 0
	c.Custom = 0
	nested := [][]byte{}
	for _, b := range data {
		if b == '[' {
			depth++
		}
		if b == ']' {
			depth--
		}
		if depth > 1 {
			nested[len(nested)-1] = append(nested[len(nested)-1], b)
			continue
		}
		// ignore numeric
		if depth == 1 && b == '#' {
			nested = append(nested, []byte{})
		}
		if depth == 1 && b == 'a' {
			nested[len(nested)-1] = append(nested[len(nested)-1], b)
		}
		if depth == 0 && b == 'a' {
			c.Custom++
		}
	}
	c.Nested = []Custom{}
	for _, v := range nested {
		sub := Custom{}
		sub.UnmarshalMarch(v)
		c.Nested = append(c.Nested, sub)
	}
	return nil
}
func (c *Custom) UnmarshalJSON(data []byte) error {
	// Some obscure unmarshal function
	return c.UnmarshalMarch(data)
}

func (c *Custom) MarshalMarch() (data []byte, err error) {
	fmt.Printf("MarshalMarch: %#v\n", c)

	// Some obscure marshal function
	data = make([]byte, c.Custom)
	for i := range data {
		data[i] = 'a'
	}
	data = append(data, '[')
	for i, v := range c.Nested {
		// TODO: Implement this for interfaces
		// And use the tryMarshal pattern here
		// Instead of locking into a recursive MarshalMarch
		var more []byte
		more, err = v.MarshalMarch()
		if err != nil {
			return
		}
		data = append(data, '#')
		data = append(data, []byte(fmt.Sprintf("%d", i))...)
		data = append(data, more...)
	}
	data = append(data, ']')
	return
}
func (c *Custom) MarshalJSON() (data []byte, err error) {
	// Some obscure marshal function
	return c.MarshalMarch()
}
