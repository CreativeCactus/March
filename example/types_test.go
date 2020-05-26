package example

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
