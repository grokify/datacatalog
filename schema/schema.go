// Package schema embeds the JSON Schema for the catalog document format. The
// schema is generated from the Go types in package datacatalog; do not edit it
// by hand.
package schema

import _ "embed"

//go:generate go run gen.go

// JSON is the JSON Schema (draft 2020-12) for a catalog document.
//
//go:embed datacatalog.schema.json
var JSON []byte
