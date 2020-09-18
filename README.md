# March

Highly flexible tag-based marshaling framework for Go.

Intuitive defaults for JSON and support for arbitrary encodings.

<img src="./docs/march.png"></img>

[![GoDoc](https://godoc.org/github.com/CreativeCactus/March?status.svg)](http://godoc.org/github.com/CreativeCactus/March)
[![Go Report](https://goreportcard.com/badge/github.com/CreativeCactus/March)](https://goreportcard.com/report/github.com/CreativeCactus/March)
[![Sourcegraph](https://sourcegraph.com/github.com/CreativeCactus/March/-/badge.svg)](https://sourcegraph.com/github.com/CreativeCactus/March?badge)

[![PROJECT STATUS](https://img.shields.io/badge/PROJECT%20STATUS-STABLIZING-yellow)](./README.md#stability)

[Show me the code!](./concepts/remaining.md#examples)

## Useful defaults

Simple:

```
type T struct {
    MyField int `March:"blah"`
}
func Test(){
    t := T{}
    err := march.Unmarshal(&t, `{"blah":123}`)
    // t == { MyField: 123 }
}
```

See also [`./example/main_test.go`](./example/main_test.go) and [`./example/cloudevents_test.go`](./example/cloudevents_test.go).

### Use any tag key

This feature allows for multiple instances of March,
if you need to handle different encodings on a single type.

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

A `march.March` is configured to use a tag key, and will use the name and flags found there to perform some un/marshaling operation.

`march.Marshal` and `.Unmarshal` functions also have differing levels of support for types other than structs.
In general, they support `map[string]*` and most primitive types quite well, even at this stage of development.

March also has some unique features, such as hoisting, explained below.

### Stability

This library is in early stages, but aims to demonstrate reliability with a thorough test suite.
If you are comfortable that your use case will fit within the existing tests, then you should be able to safely use this library.

`./task test` or `go test -v ./...`

Note that this library uses reflection, so code coverage % may not give a good indication of actual test coverage.
See `./task test` (using `./task type test` in bash) to generate coverage.
`-coverpkg=./...` is needed because tests are run from different package namespaces (to test public/private property access accurately).

For anything serious, be sure to use [good `recover` practices](https://blog.golang.org/defer-panic-and-recover).

### Speed

Speed is a secondary priority to Stability,
however I aim to compete with `encoding/json`.

## Bugs

- `UnmarshalAsJSON` and `UnmarshalTo` will crash if given an `interface{}` or a complex type containing it (eg. `[]interface{}`).

Future versions will support intelligent un/marshaling of `interface{}` (like `encoding/json` does).

Instead, use `json.RawMessage` or `march.RawUnmarshal`.

## Type Support

March itself can support any type.

The default `Un/MarshalAsJSON` functions recurse into the given type to run custom un/marshal methods.

See [**JSON Encoder**](./concepts/README.md#json-encoder).

### Nested structs

Well supported.

### ~~Embedded types~~

It can be done, but [see here](https://stackoverflow.com/a/28977064) for some reasons why it is currently excluded. 

## Precedence

Any type can specify any number of custom un/marshaler methods.

See below: `Extensibility`

Calls to `march.Un/Marshal` or `(M := March{}).Un/Marshal` with a value of type `T` will try the following:

- `T.Un/MarshalSUFFIX` where `SUFFIX` is `M.Suffix` OR `M.Tag` OR `"March"`
- `M.DefaultUn/Marshaler`
- `M.Un/MarshalDefault` AKA `M.Un/MarshalAsJSON`

## Features

### Options

See [GoDoc](https://godoc.org/github.com/CreativeCactus/March#March).

### RawUnmarshal

![FEATURE STATUS](https://img.shields.io/badge/FEATURE-UNSTABLE-yellow)

`RawUnmarshal` is a wrapper type which acts like `json.RawMessage`.
It is useful for lazily unmarshaling.
It can be used to `.UnmarshalTo(&v)` some value,
and should behave exactly like a call to `march.Unmarshal`.

It also supports `.TypeOfJSON` for checking the underlying type of the message.

It still has some limitations, for example:

```
    type T struct {
        Val    march.RawUnmarshal     `March:"val"`
    }
    data := []byte(`{ "val":{"x":[1,2,"3"]} }`)
    V := T{}
    S := ""
    err := march.Unmarshal(data, &V).Val.UnmarshalTo(&S)
    // err: json: cannot unmarshal object into Go value of type string
```

### Flags

![FEATURE STATUS](https://img.shields.io/badge/FEATURE-STABLE-green)

```
    type T struct {
        V int64 `March:"v,hoist,remains,somethingElse"`
    }
```

While some details and flag names may change, the overall design is now stable.

Flag implementation is actually a feature of the encoder (Un/Marshal methods `M.MarshalDefault` and `M.UnmarshalDefault`), so custom implementations will each need to implement flags.
March provides functions like `GetTagPart`, `GetTagFlags`, `FlagsContain`, `IsValidTagName` to help with this.

Note that some flags can result in Un/Marshalers for which some valid JSON is structured differently when re-marshaled (non-idempotent encoders).
If that sounds confusing then it is probably not a concern for your use case.
See `TestIdempotent` in [./example/flags_test.go](./example/flags_test.go).

#### [Hoisting](./concepts/hoisting.md)

#### [Remaining](./concepts/remaining.md)

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

See [**Encoder Methods**](./concepts/README.md#encoder-methods)

See also [**Encoder Extensions**](./concepts/README.md#encoder-extensions)

### ~~Custom flag handlers~~

See [**Encoder Extensions**](./concepts/README.md#encoder-extensions)

In the meantime, you can still implement custom un/marshalers
to handle any kind of flag you want. See [./example/custom_test.go](./example/custom_test.go).

### ~~Configuration Checking~~

Not currently implemented.

Checker functions could be used at runtime to determine things like:

- Whether a type supports custom UnmarshalAsX calls
- Whether a type has an idempotent Marshal/Unmarshal cycle for a given value

## ~~Value Checking~~

There are no plans to support value checking,
since types can be aliased and provided with their own un/marshalers.

There are plans to demonstrate the recommended patterns for this, however.

## Tips and FAQ

### Always match function signatures used by March

March will panic if a reflection-invoked function does not match the expected signature. In future versions they will simply be ignored.

### My custom methods aren't being used

Be careful to provide the value which matches the function signature. If your function takes a pointer, you might need to provide a pointer value to Un/Marshaling functions. This also applies to properties on structs.

Note that encoders MAY wrap values in pointers automatically (see [JSON Fallback](./concepts/README.md#json-fallback))

## Spelling

Marshal, marshaling, marshaled, marshaler.

One `l` is used everywhere for consistency with `encoding/json`.
