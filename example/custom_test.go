package example

import (
	"fmt"
	"log"
	"strconv"

	march "github.com/CreativeCactus/March"
)

// Custom demonstrates the implementation of a custom marshal/unmarshal syntax
type Custom struct {
	Custom int      `March:"custom" json:"custom"`
	Nested []Custom `March:"nested" json:"nested"`
}

func numeric(b byte) bool {
	_, err := strconv.Atoi(string(b))
	return err == nil
}

// UnmarshalMarch is the unmarshaler for March{Tag:"March"}
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

// UnmarshalJSON is the unmarshaler for JSON
func (c *Custom) UnmarshalJSON(data []byte) error {
	// Some obscure unmarshal function
	return c.UnmarshalMarch(data)
}

// MarshalMarch is the marshaler for March{Tag:"March"}
func (c *Custom) MarshalMarch() (data []byte, err error) {
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

// MarshalJSON is the marshaler for JSON
func (c *Custom) MarshalJSON() (data []byte, err error) {
	// Some obscure marshal function
	return c.MarshalMarch()
}

// Example shows an example of using March
func Example() {
	M := march.March{Tag: "March"}
	m := Custom{
		Custom: 4,
		Nested: []Custom{
			{
				Custom: 8,
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
		log.Fatalf("March Marshal Error: %s", err.Error())
	}

	fmt.Printf("March    Marshalled data: %s\n", string(data))

	m = Custom{}
	err = M.Unmarshal(data, &m)
	if err != nil {
		log.Fatalf("March Re-Unmarshal Error: %s", err.Error())
	}

	data, err = M.Marshal(&m)
	if err != nil {
		log.Fatalf("March Re-Marshal Error: %s", err.Error())
	}

	fmt.Printf("March Re-Marshalled data: %s\n", string(data))
	// Output:
	// March    Marshalled data: aaaa[#0aaaaaaaa[#0aa[]#1aa[]]]
	// March Re-Marshalled data: aaaa[#0aaaaaaaa[#0aa[]#1aa[]]]
}
