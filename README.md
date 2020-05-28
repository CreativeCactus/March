# March

Highly flexible tag-based marshalling framework for Go.

<img src="./docs/march.png"></img>

[![GoDoc](https://godoc.org/github.com/CreativeCactus/March?status.svg)](http://godoc.org/github.com/CreativeCactus/March)
[![Go Report](https://goreportcard.com/badge/github.com/CreativeCactus/March)](https://goreportcard.com/report/github.com/CreativeCactus/March)
[![Sourcegraph](https://sourcegraph.com/github.com/CreativeCactus/March/-/badge.svg)](https://sourcegraph.com/github.com/CreativeCactus/March?badge)

## Useful defaults

Simple:

```
type T struct {
    MyField int `March:"blah"`
}
func Test(){
    t := T{}
    err := march.Unmarshal(&t, `{"blah":123}`)
}
```

```
    // See ./example/main_test.go Example function
```

### Use any tag key

This feature allows for multiple instances of March,
if you need to handle different kinds of translation on a single type.

```
type T struct {
    MyField int `A:"blah" B:"blEh"`
}
func Test(){
    ma := march.March{ Tag:"A" }
    me := march.March{ Tag:"B" }
    t := T{}
    err := ma.Unmarshal(&t, `{"blah":123}`)
    data, err := me.Marshal(t)
    // string(data) == `{"blEh":123}`
}
```

Note that `March{}` is identical to `March{Tag:"March"}`.

## Overview

Before going further into the more advanced features, here is a primer:

**Tags** are annotations on struct fields in Go. A field can have several tags: `Tag1:"Blah" json:"blah"`

**Tag key** and **Tag value** refer to the two parts of a tag `Key:"Value"`

**Tag name** is the first part of a Tag value.
The remaining *comma-separated* parts are called **Tag flags**: `Key:"Name,Flag,Flag"`

A `march.March` is configured to use a tag key, and will use the name and flags found there to perform some un/marshalling operation.

`march.Marshal` and `.Unmarshal` functions also have differing levels of support for types other than structs.
In general, they support `map[string]*` and most primitive types quite well, even at this stage of development.

March also has some unique features, such as hoisting, explained below.

### Stability

This library is in early stages, but aims to demonstrate reliability via a thorough test suite. If you are comfortable that your use case will fit within the existing tests, then you should be able to safely use this library.

`./task test` or `go test -v ./...`

Note that this library uses reflection, so code coverage % does not give a good indication of actual test coverage. As of writing coverage is around 70%-80%, but it cannot be seen when running tests because they are run from a different package namespace (to test public/private property access accurately).

For anything serious, be sure to use [good `recover` practices](https://blog.golang.org/defer-panic-and-recover).

## Precedence

Any type can specify any number of custom un/marshaller methods.

See below: `Extensibility`

Calls to `Un/Marshal` or `March{}.Un/Marshal` with a value of type `T` will try the following:

- `T.Un/MarshalSUFFIX` where `SUFFIX` is `M.Suffix` OR `M.Tag` OR `"March"`
- `M.DefaultUn/Marshaller`
- `M.Un/MarshalDefault` AKA `M.Un/MarshalJSON`

## Features

### Flags

```
    type T struct {
        V int64 `March:"v,hoist,remains,somethingElse"`
    }
```

This functionality is experimental, and is without a solid reference implementation, so some details may take time to solidify.

This is actually a feature of the default Un/Marshal methods (`M.MarshalDefault` and  `M.UnmarshalDefault`), so custom implementations will each need to implement flags.
March provides functions like `GetTagPart`, `GetTagFlags`, `FlagsContain`, `IsValidTagName` to help with this.

Note that some flags can result in Un/Marshallers for which some valid JSON is structured differently when re-marshalled.
If that sounds confusing then it is probably not a concern for your use case.
See `TestIdempotent` in [./example/flags_test.go](./example/flags_test.go).

#### Hoist

See [./example/flags_test.go](./example/flags_test.go).

```
    type T struct {
        V T2 `March:"_,hoist,"`
    }
    type T2 struct {
        V2 int `March:"toplevel"`
    }
```

`{"toplevel":0}`

The `hoist` flag is used to tell `MarshalDefault` (AKA `MarshalJSON`) to bring the contents of some field into the top level scope.

It currently supports the following types: `struct`, `map[string]*`

Currently more than one "level" of hoisting is supported but not tested.

Note that the tag name (the first part of the tag, `_` in the above example) is ignored, but must be valid (non empty).

Note that `remains` might make Un/Marshal calls no longer idempotent.
See `TestIdempotent` in [./example/flags_test.go](./example/flags_test.go).

#### Remains

See [./example/flags_test.go](./example/flags_test.go).

```
    type T struct {
        V map[string]json.RawMessage `March:"_,remains,,"`
    }
```

The `remains` flag is used to tell `UnmarshalDefault` (AKA `UnmarshalJSON`) to put any unassigned fields (fields without tags) into this property as a `map[string]json.RawMessage`.

Note that no `UnmarshalX` operations are called on these fields, they are the the value of the `map[string][]byte` provided by `ReadFieldsX` with assigned fields removed, and converted to the type of the field, which must be one of the following: 

- `map[string][]byte` Note that []byte will be marshalled in base64 by default, resulting in potentially asymmetric un/marshallers.
- `map[string]json.RawMessage` A thin wrapper around []byte with marshalling built in.
- `map[string]march.RawData` A thin wrapper around RawMessage with unmarshalling helpers.

Multiple remain fields will receive copies.

Note that the tag name (the first part of the tag, `_` in the above example) is ignored, but must be valid.

Note that the type of a `remains` flagged field (in the case of the `UnmarshalDefault` AKA `UnmarshalJSON` implementation) must be `map[string]json.RawMessage`

Note that `remains` might make Un/Marshal calls no longer idempotent.

#### ~~Lazy~~

Not yet supported. A good first issue.
It is implemented by remains, and only needs to be set up as a flag.

#### ~~Dot notation~~

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

#### ~~Omit empty~~

Currently not implemented. A good first issue.
Trivial to implement and shows the end to end process of building a flag.

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

### ~~Custom flag handlers~~

Maybe some day. In the meantime, you can still implement custom un/marshallers
to handle any kind of flag you want. See [./example/custom_test.go](./example/custom_test.go).

### ~~Configuration Checking~~

Not currently implemented.

Checker functions can be used at runtime to determine things like:

- Whether a type supports custom UnmarshalX calls
- Whether a type has an idempotent Marshal/Unmarshal cycle

## ~~Value Checking~~

There are no plans to support value checking,
since types can be aliased and provided with their own un/marshallers.

There are plans to demonstrate the recommended patterns for this, however.

## Nested structs

Well supported.

## ~~Embedded structs~~

It can be done, but [see here](https://stackoverflow.com/a/28977064) for some reasons why it is currently excluded. 

## Concepts

Where `M` is the March instance and `T` is the type being passed to it.
`X` is the tag of `M` (see `Precedence` section).

### Marshal consist of the following stages:

- Iterate over all fields on `T` (using reflection) if it is a struct or map, otherwise marshal directly
    - Try to call `T.MarshalX`, otherwise call `M.MarshalJSON`
    - Store values as named fields: `map[string][]byte`
- Pass fields to `T.WriteFieldsX` or `WriteFieldsJSON` and return

### Unmarshal consist of the following stages:

- Call `T.ReadFieldsX` or `ReadFieldsJSON`
    - Convert the given `[]byte` into named fields: `map[string][]byte`
- Iterate over fields on `T`
    - Attempt to call `T.UnmarshalX`, otherwise call `M.UnmarshalJSON` on fields
    - Assign values on `T`

### ReadFieldsX and WriteFieldsX

Think of these functions as `Read from JSON` and `Write to JSON`.

They are responsible for listing the fields and values of a JSON (or "X") blob, and turning a list of fields and values back into a JSON (or "X") blob.

## Tips and FAQ

### Always match function signatures used by March

March will panic if a reflection-invoked function does not match the expected signature. In future versions they will simply be ignored.

### My custom methods aren't being used

Be careful to provide the value which matches the function signature. If your function takes a pointer, you will need to provide a pointer value to Un/Marshalling functions.
