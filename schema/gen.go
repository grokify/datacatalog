//go:build ignore

// Command gen writes schema/datacatalog.schema.json from the Go types in
// package datacatalog. Run it with `go generate ./schema`.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/invopop/jsonschema"

	"github.com/grokify/datacatalog"
)

func main() {
	out := flag.String("out", "datacatalog.schema.json", "output file, or - for stdout")
	flag.Parse()
	if err := run(*out); err != nil {
		fmt.Fprintln(os.Stderr, "gen:", err)
		os.Exit(1)
	}
}

func run(out string) error {
	r := &jsonschema.Reflector{
		FieldNameTag:               "json",
		RequiredFromJSONSchemaTags: false,
	}
	s := r.Reflect(&datacatalog.Catalog{})
	s.ID = "https://github.com/grokify/datacatalog/schema/datacatalog.schema.json"
	s.Title = "Data Catalog"
	s.Description = "Entities and fields an application stores, and how each is classified."
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if out == "-" {
		_, err = os.Stdout.Write(b)
		return err
	}
	return os.WriteFile(out, b, 0o644) //nolint:gosec // generated public schema
}
