# March

Highly flexible tag-based marshalling framework for Go.

<img src="./docs/march.png"></img>

## Useful defaults

The only thing you *need* to specify is a tag key.

```
type T struct {
    MyField int `MyApp:"blah"`
}
func Test(){
    m := march.March{ Tag:"MyApp" }
    t := T{}
    err := m.Unmarshal(&t, `{"blah":123}`)
}
```

```
    // See ./example/main_test.go Example function
```

## Use any tag key

This feature allows for multiple instances of March,
if you need to handle different kinds of translation on a single type.

```
type T struct {
    MyField int `A:"blah" B:"bleh"`
}
func Test(){
    ma := march.March{ Tag:"blah" }
    me := march.March{ Tag:"blEh" }
    t := T{}
    err := ma.Unmarshal(&t, `{"blah":123}`)
    data, err := me.Marshal(t)
    // string(data) == `{"blEh":123}`
}
```

## Flags

```
    type T struct {
        V int64 `March:"v,hoist,remains,somethingElse"`
    }
```

These are some of the few features supported outside of `encoding/json`s feature set, and can be used in conjunction with `encoding/json` if this is the only feature you want.

This is actually a feature of the default Un/Marshal methods (`M.MarshalDefault` and  `M.UnmarshalDefault`), so custom implementations will each need to implement flags.

March provides functions like `GetTagPart`, `GetTagFlags`, `FlagsContain`, `IsValidTagName` to help with this.

This functionality is experimental, and is without a solid reference implementation, so some details may take time to solidify.

Note that some flags can result in Un/Marshallers for which some valid JSON is presented differently when re-marshalled.
If that sounds confusing then it is probably not a concern for your use case.
See `TestIdempotent` in [./example/flags_test.go](./example/flags_test.go).

### Hoist

See [./example/flags_test.go](./example/flags_test.go).

```
    type T struct {
        V T2 `March:"_,hoist,"`
    }
    type T2 struct {
        V2 int `March:"toplevel"`
    }
```

The `hoist` flag is used to tell `MarshalDefault` (AKA `MarshalJSON`) to bring the contents of some field into the top level scope.

It currently supports the following types: `struct`

Currently only one "level" of hoisting is supported.

Note that the tag name (the first part of the tag, `_` in the above example) is ignored, but must be valid.

Note that `remains` might make Un/Marshal calls no longer idempotent.

### Remains

See [./example/flags_test.go](./example/flags_test.go).

```
    type T struct {
        V map[string]json.RawMessage `March:"_,remains,,"`
    }
```

The `remains` flag is used to tell `UnmarshalDefault` (AKA `UnmarshalJSON`) to put any unassigned fields (fields without tags) into this property as a `map[string]json.RawMessage`.

Note that no `UnmarshalX` operations are called on these fields, they are the the value of the `map[string][]byte` provided by `ReadFieldsX` with assigned fields removed, and converted to `map[string]json.RawMessage`.

Multiple remain fields will receive copes.

Note that the tag name (the first part of the tag, `_` in the above example) is ignored, but must be valid.

Note that the type of a `remains` flagged field (in the case of the `UnmarshalDefault` AKA `UnmarshalJSON` implementation) must be `map[string]json.RawMessage`

Note that `remains` might make Un/Marshal calls no longer idempotent.

### ~~Dot notation~~

Dot notation is not currently supported, but would provide a more flexible alternative to `remains`.

It would allow for syntax like this:

```
    type T struct {
        V int64  `March:"data.v"`
        W int64  `March:"data.w"`
        X string `March:"list[0]"`
    }
```

To marshal to/from JSON like this:

`{ "data": { "v": 1, "w": 2 }, "list": [ "x" ] }`

### ~~Omit empty~~

Currently not implemented. A good first feature.

## Top level types

Currently supports structs only.

Plans to support maps and arrays soon.

## Extensibility

Where `T` is the type provided to the Un/Marshal function.
`X` is the `Tag` for your March instance.

Note: These methods will currently panic if they are written with the correct signature. TODO: use NumIn, NumOut.

### T.MarshalX, T.UnmarshalX

If implemented, any `T` will be un/marshalled using this method instead of the default for the March instance.

```
    // See ./example/custom_test.go Custom type
```

### T.WriteFieldsX, T.ReadFieldsX

If implemented, any `T` will use these low level methods to write (to `[]byte`) or read (to `map[string][]byte`) the provided value.

This is useful for changing the behavior of an existing `Un/MarshalX` function without rewriting it completely.

```
    // See ./example/unmarshal_test.go TestUnmarshalReadFields
```

## Stability

This library is in early stages, but aims to demonstrate reliability via a thorough test suite. If you are comfortable that your use case will fit within the existing tests, then you should be able to safely use this library.

`./task test` or `go test -v ./...`

Note that this library uses reflection, so code coverage % does not give a good indication of actual test coverage. As of writing coverage is around 75%-80%, but it cannot be seen when running tests because they are run from a different package namespace (to test public/private property access accurately).

For anything serious, be sure to use [good `recover` practices](https://blog.golang.org/defer-panic-and-recover).

## ~~Checking~~

Not currently implemented.

Checker functions can be used at runtime to determine things like:

- Whether a type supports custom UnmarshalX calls
- Whether a type has an idempotent Marshal/Unmarshal cycle

## Nested structs

Well supported.

## ~~Embedded structs~~

It can be done, but [see here](https://stackoverflow.com/a/28977064) for some reasons why it is currently excluded. 

## Concepts

Where M is the March instance and T is the type being passed to it.
X is the `M.Tag`.

### Marshal consist of the following stages:

- Try to call `T.MarshalX`, otherwise call `M.MarshalJSON`
    - Iterate over all fields on `T` (using reflection)
        - Try to call `T.MarshalX`, otherwise call `M.MarshalJSON`
        - Store values as named fields: `map[string][]byte`
    - Pass fields to `T.WriteFieldsX` or `WriteFieldsJSON` and return.

### Unmarshal consist of the following stages:

- Try to call `T.UnmarshalX`, otherwise call `M.UnmarshalJSON`
    - Call `T.ReadFieldsX` or `ReadFieldsJSON`
        - Convert the given `[]byte` into named fields: `map[string][]byte`
    - Iterate over fields on `T`
        - Attempt to call `T.UnmarshalX`, otherwise call `M.UnmarshalJSON` on fields
        - Assign values on `T`

### ReadFieldsX and WriteFieldsX

Think of these functions as `Read from JSON` and `Write to JSON`.

They are responsible for listing the fields and values of a JSON blob, and turning a list of fields and values back into a JSON blob.

## Tips and FAQ

### Always match function signatures used by March

March will panic if a function does not match the expected signature. In future versions they will simply be ignored.

### My custom methods aren't being used

Be careful to provide the value which matches the function signature. If your function takes a pointer, you will need to provide a pointer value to Un/Marshalling functions.
