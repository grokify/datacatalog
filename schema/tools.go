//go:build tools

package schema

// Keep the generator's dependencies in go.mod: gen.go is excluded from normal
// builds, so go mod tidy would otherwise remove them.
import _ "github.com/invopop/jsonschema"
