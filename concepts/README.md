# Concepts

## Terminology

- **Value** refers to an argument (being un/marshaled) to a March method.
- **Encode** is used synonymously with us/marshal.
- **Tags & Flags** describe components of the golang tags used by March. See below for details.

## Encoders

The highest level component of March is the encoder.
It is implemented as **Encoder Methods** (see below) `MarshalAs*` and `UnmarshalAs*`.

By default March has an intuitive JSON encoder, which supports most of the features described in the [README](../README.md).

Values can be marshaled with `march.Marshal(Value)`, which is the same as `march.March{}.Marshal(Value)`.
In the absense of **Encoder Methods** (see below) on the value, this will call `march.March{}.MarshalAsJSON(Value)`.
See **JSON Encoder**. See also [**Precedence**](../README.md#precedence) for more information.

### JSON Encoder

```golang
data, err := march.MarshalAsJSON(v)
err = march.UnmarshalAsJSON(data, &v)
```

The default `marshalAsJSON` implementation automatically supports `Slice`, `Array`, `Ptr`, `Struct`.
Other types will be checked for methods or passed directly to `json.Marshal`.
Tests show this working with `time.Time` in `./unmarshal_test.go TestUnmarshalComposite`. Note that the value passed to `time.Parse` is currently nested in quotes.
`map` is supported via `json.Marshal`, but the underlying types are not yet handled by March.

The default `unmarshalAsJSON` implementation automatically supports `Slice`, `Ptr`, `Struct`.
Other types will be checked for methods or passed directly to `json.Unmarshal`.
`Array` could be supported if provided as a `Ptr`. Type aliasing can be used to implement this now.
`Map` is supported via `json.Unmarshal`, but the underlying types are not yet handled by March.

Some flags have specific requirements, see below.

#### Features

Features can also be implemented as `Extensions` (see below).

- [`hoist` flag](./hoisting.md)
- [`remains` flag](./remaining.md)

#### JSON Fallback

By default, the JSON encoder methods (*AsJSON) will automatically
use Un/MarshalJSON methods if they exist on the given value or a pointer to it.
This is used to give expected functionality when using types like `time.Time`.
This can be disabled with the following:

```golang
march.March{
    NoMarshalJSON:   false,
    NoUnmarshalJSON: false,
}
```

Until this is supported by flags at the field level, this will prevent
calls from `encoding/json` to March methods (because it would invite recursion).

## Encoder Methods

Encoder methods are methods on a value to be Un/Marshaled.
They are called by March at runtime, using [reflection](https://golang.org/pkg/reflect/).

```golang
type Custom struct {}
func (c *Custom) MarshalMarch() (data []byte, err error) {}
func (c *Custom) UnmarshalMarch(data []byte) error {}
```

[See examples](../example/custom_test.go).

[![STATUS](https://img.shields.io/badge/TRY%20IT%20OUT-green)](https://play.golang.org/p/qY7D9AYjkiU)

Where `T` is the type of value provided to the Un/Marshal function.
`X` is the `Tag` for your March instance.

Note: These methods will currently panic if they are written with the correct signature. TODO: use NumIn, NumOut.

### T.MarshalX, T.UnmarshalX

If implemented, any `T` will be un/marshaled using this method instead of the default for the March instance.

```
    // See ./example/custom_test.go Custom type
```

### T.WriteFieldsX, T.ReadFieldsX

If implemented, any `T` will use these low level methods to write (to `[]byte`) or read (to `map[string][]byte`) the provided value.

This is useful for changing the behavior of an existing encoder without rewriting it completely.

```
    // See ./example/unmarshal_test.go TestUnmarshalReadFields
```

## Encoder Extensions

![FEATURE STATUS](https://img.shields.io/badge/FEATURE-EXPERIMENTAL-red)

Encoder Extensions are methods which can be called *within* encoder methods, usually according to the flags on the value being encoded.

Currently only `hoist` is implemented as an encoder extension, and only `MarshalAsJSON` supports calling encoder extensions.

## Tags & Flags

**Tags** are annotations on struct fields in Go. A field can have several tags: `Tag1:"Blah" json:"blah"`

**Tag key** and **Tag value** refer to the two parts of a tag `Key:"Value"`

**Tag name** is the first part of a Tag value.
The remaining *comma-separated* parts are called **Tag flags**: `Key:"Name,Flag,Flag"`

## Encoder Architecture

Where `M` is the March instance and `T` is the type being passed to it.
`X` is the tag of `M` (see [**Precedence**](../README.md#precedence)).

### Marshal consist of the following stages:

- Recurse into all fields on `T` (using reflection) if it is a struct or map, otherwise marshal directly
    - Try to call `T.MarshalAsX`, otherwise call `M.MarshalAsJSON`
    - Store values as named fields: `map[string][]byte`
- Pass fields to `T.WriteFieldsX` or `WriteFieldsJSON` and return

### Unmarshal consist of the following stages:

- Call `T.ReadFieldsX` or `ReadFieldsJSON`
    - Convert the given `[]byte` into named fields: `map[string][]byte`
- Iterate over fields on `T`
    - Attempt to call `T.UnmarshalAsX`, otherwise call `M.UnmarshalAsJSON` on fields
    - Assign values on `T`

### ReadFieldsX and WriteFieldsX

Think of these functions as `Read from JSON` and `Write to JSON`.

They are responsible for listing the fields and values of a JSON (or "X") blob, and turning a list of fields and values back into a JSON (or "X") blob.
