# March

Highly flexible tag-based marshalling framework for Go.

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

## Extensibility

Where `T` is the type provided to the Un/Marshal function.
`X` is the `Tag` for your March instance.

Note: These methods will currently panic if they are written with the correct signature. TODO: use NumIn, NumOut.

### T.MarshalX, T.UnmarshalX

If implemented, any `T` will be un/marshalled using this method instead of the default for the March instance.

```
    // See ./example/main_test.go Custom type
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

## Tips and FAQ

### Always match function signatures used by March

March will panic if a function does not match the expected signature. In future versions they will simply be ignored.

### My custom methods aren't being used

Be careful to provide the value which matches the function signature. If your function takes a pointer, you will need to provide a pointer value to Un/Marshalling functions.

