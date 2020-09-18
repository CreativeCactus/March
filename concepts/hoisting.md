# Hoisting

![FEATURE STATUS](https://img.shields.io/badge/FEATURE-STABLE-green)

Value hosting is when a nested value is marshaled into a "top level" field.

```
    type T struct {
        V T2 `March:"_,hoist,"`
    }
    type T2 struct {
        V2 int `March:"toplevel"`
    }
    // T will marshal to `{"toplevel":0}`
```

The `hoist` flag is used to tell `MarshalDefault` (AKA `MarshalJSON`) to bring the contents of some field into the top level scope.

It currently supports the following types: `struct`, `map[string]*`

See [./example/flags_test.go](./example/flags_test.go).

Currently more than one "level" of hoisting is supported but not tested.

Note that the tag name (the first part of the tag, `_` in the above example) is ignored, but must be valid (non empty).

Note that without `remains`, `hoist` might make Un/Marshal calls no longer idempotent.
See `TestIdempotent` in [./example/flags_test.go](./example/flags_test.go).
