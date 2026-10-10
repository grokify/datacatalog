package datacatalog

import (
	"bytes"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestJSONRoundTrip(t *testing.T) {
	in := exampleCatalog()
	data, err := in.MarshalJSONIndent()
	if err != nil {
		t.Fatal(err)
	}
	out, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	in.Normalize()
	if !reflect.DeepEqual(in, out) {
		t.Errorf("round trip mismatch:\n in: %+v\nout: %+v", in, out)
	}
	if !bytes.HasSuffix(data, []byte("}\n")) {
		t.Error("output should end with a newline")
	}
}

func TestEncodeStableOrdering(t *testing.T) {
	a := exampleCatalog()
	b := exampleCatalog()
	b.Entities[0], b.Entities[1] = b.Entities[1], b.Entities[0]
	var ba, bb bytes.Buffer
	if err := a.Encode(&ba); err != nil {
		t.Fatal(err)
	}
	if err := b.Encode(&bb); err != nil {
		t.Fatal(err)
	}
	if ba.String() != bb.String() {
		t.Error("encoding depends on entity order")
	}
	if b.Entities[0].Name != "invoice" {
		t.Error("Encode must not reorder the receiver")
	}
	if strings.Index(ba.String(), `"account"`) > strings.Index(ba.String(), `"invoice"`) {
		t.Error("entities not sorted by name")
	}
}

func TestDecodeErrors(t *testing.T) {
	tests := []struct {
		name, in string
	}{
		{"malformed", `{"name":`},
		{"unknown catalog field", `{"name":"x","extra":1}`},
		{"unknown field property", `{"name":"x","entities":[{"name":"e","fields":[{"name":"f","visibility":"public","pii_flag":true}]}]}`},
		{"field level versioned", `{"name":"x","entities":[{"name":"e","fields":[{"name":"f","visibility":"public","versioned":true}]}]}`},
		{"trailing data", `{"name":"x"} {"name":"y"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Parse([]byte(tt.in)); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestDecodeDoesNotDefaultVisibility(t *testing.T) {
	c, err := Parse([]byte(`{"name":"x","entities":[{"name":"e","fields":[{"name":"f"}]}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if c.Entities[0].Fields[0].Visibility != "" {
		t.Error("visibility must stay unset")
	}
	if err := c.Validate(); err == nil {
		t.Error("unset visibility must fail validation")
	}
}

func TestFileRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catalog.json")
	in := exampleCatalog()
	if err := in.WriteFile(path); err != nil {
		t.Fatal(err)
	}
	out, err := ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := out.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, err := ReadFile(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("expected error for missing file")
	}
	if err := in.WriteFile(filepath.Join(t.TempDir(), "no", "dir", "c.json")); err == nil {
		t.Error("expected error writing into a missing directory")
	}
}

func TestNilCatalogIO(t *testing.T) {
	var c *Catalog
	if _, err := c.MarshalJSONIndent(); err == nil {
		t.Error("expected error")
	}
	if err := c.Encode(&bytes.Buffer{}); err == nil {
		t.Error("expected error")
	}
	if err := c.WriteFile(filepath.Join(t.TempDir(), "x.json")); err == nil {
		t.Error("expected error")
	}
	if c.Clone() != nil {
		t.Error("nil clone should be nil")
	}
	c.Normalize() // must not panic
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, bytes.ErrTooLarge }

func TestEncodeWriteError(t *testing.T) {
	if err := exampleCatalog().Encode(failWriter{}); err == nil {
		t.Error("expected write error")
	}
}

func TestCloneIsDeep(t *testing.T) {
	a := exampleCatalog()
	b := a.Clone()
	b.Entities[0].Fields[0].Name = "changed"
	if a.Entities[0].Fields[0].Name == "changed" {
		t.Error("clone shares field storage")
	}
}
