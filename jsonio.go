package datacatalog

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

// Clone returns a deep copy of the catalog.
func (c *Catalog) Clone() *Catalog {
	if c == nil {
		return nil
	}
	out := &Catalog{Name: c.Name}
	if c.Entities != nil {
		out.Entities = make([]Entity, len(c.Entities))
		for i, e := range c.Entities {
			e.Fields = slices.Clone(e.Fields)
			out.Entities[i] = e
		}
	}
	return out
}

// Normalize sorts entities by name, in place. Field order is kept because
// declaration order is meaningful (it usually mirrors column order).
func (c *Catalog) Normalize() {
	if c == nil {
		return
	}
	slices.SortStableFunc(c.Entities, func(a, b Entity) int { return strings.Compare(a.Name, b.Name) })
}

// Decode reads a JSON catalog. Unknown fields are rejected so typos in
// classifications are not silently ignored. Decode does not validate; call
// Validate on the result.
func Decode(r io.Reader) (*Catalog, error) {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	var c Catalog
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("datacatalog: decode: %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("datacatalog: decode: trailing data after catalog")
	}
	return &c, nil
}

// Parse decodes a JSON catalog from bytes. See Decode.
func Parse(data []byte) (*Catalog, error) {
	return Decode(bytes.NewReader(data))
}

// ReadFile loads a JSON catalog from a file. See Decode.
func ReadFile(path string) (*Catalog, error) {
	f, err := os.Open(path) //nolint:gosec // path is chosen by the caller
	if err != nil {
		return nil, fmt.Errorf("datacatalog: %w", err)
	}
	c, decErr := Decode(f)
	if closeErr := f.Close(); decErr == nil && closeErr != nil {
		return nil, fmt.Errorf("datacatalog: %w", closeErr)
	}
	return c, decErr
}

// MarshalJSONIndent returns the catalog as indented JSON with entities sorted
// by name and a trailing newline, so output is stable and diff-friendly. The
// receiver is not modified.
func (c *Catalog) MarshalJSONIndent() ([]byte, error) {
	if c == nil {
		return nil, errors.New("datacatalog: nil catalog")
	}
	n := c.Clone()
	n.Normalize()
	b, err := json.MarshalIndent(n, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("datacatalog: encode: %w", err)
	}
	return append(b, '\n'), nil
}

// Encode writes the catalog as stable indented JSON. See MarshalJSONIndent.
func (c *Catalog) Encode(w io.Writer) error {
	b, err := c.MarshalJSONIndent()
	if err != nil {
		return err
	}
	if _, err := w.Write(b); err != nil {
		return fmt.Errorf("datacatalog: encode: %w", err)
	}
	return nil
}

// WriteFile saves the catalog as stable indented JSON.
func (c *Catalog) WriteFile(path string) error {
	b, err := c.MarshalJSONIndent()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o644); err != nil { //nolint:gosec // catalogs are documentation, not secrets
		return fmt.Errorf("datacatalog: %w", err)
	}
	return nil
}
