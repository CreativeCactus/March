# Remaining

![FEATURE STATUS](https://img.shields.io/badge/FEATURE-STABLE-green)

See [./example/flags_test.go](./example/flags_test.go).

```
    type T struct {
        V map[string]json.RawMessage `March:"_,remains,,"`
    }
```

The `remains` flag is used to tell `UnmarshalDefault` (AKA `UnmarshalAsJSON`) to put any unassigned fields (fields without tags) into this property as a `map[string]json.RawMessage`.

Note that no `UnmarshalAsX` operations are called on these fields, they are the the value of the `map[string][]byte` provided by `ReadFieldsX` with assigned fields removed, and converted to the type of the field, which must be one of the following: 

- `map[string][]byte` Note that []byte will be marshaled in base64 by default, resulting in potentially asymmetric un/marshalers.
- `map[string]json.RawMessage` A thin wrapper around []byte with marshaling built in.
- `map[string]march.RawUnmarshal` A thin wrapper around RawMessage with unmarshaling helpers.

Multiple remain fields will receive copies.

Note that the tag name (the first part of the tag, `_` in the above example) is ignored, but must be valid.

Note that the type of a `remains` flagged field (in the case of the `UnmarshalDefault` AKA `UnmarshalAsJSON` implementation) must be `map[string]json.RawMessage`

Note that `remains` might make Un/Marshal calls no longer idempotent unless used carefully.

## Examples

```golang
package main

import (
	"fmt"
	march "github.com/CreativeCactus/March"
)

type X struct {
	Name   string                        `March:"name"`
	Value  march.RawUnmarshal            `March:"value"`
	Extras map[string]march.RawUnmarshal `March:"_,hoist,remains"`
}

func main() {
	var err error
	x := X{}

	err = march.Unmarshal([]byte(`{"a":1, "b":2.0, "v": { "nested": true }}`), &x)
	if err != nil {
		return
	}
	for k, v := range x.Extras {
		fmt.Printf("%s: %T (%s)", k, v, v.TypeOfJSON())
	}

	err = x.Value.UnmarshalTo(&x) // Replace x with the unmarshaled contents of x.V
	if err != nil {
		return
	}
	for k, v := range x.Extras {
		fmt.Printf("%s: %T (%s)", k, v, v.TypeOfJSON())
	}
}
```

[![STATUS](https://img.shields.io/badge/TRY%20IT%20OUT-green)](https://play.golang.org/p/cJM0FsVPKhy)


- [Can NOT remain valus onto a `map[int]`](https://play.golang.org/p/HV8ez9ZlgYm)