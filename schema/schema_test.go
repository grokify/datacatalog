package schema_test

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"testing"

	"github.com/grokify/datacatalog/schema"
)

func TestEmbeddedSchemaIsJSON(t *testing.T) {
	var doc map[string]any
	if err := json.Unmarshal(schema.JSON, &doc); err != nil {
		t.Fatalf("embedded schema is not valid JSON: %v", err)
	}
	defs, ok := doc["$defs"].(map[string]any)
	if !ok {
		t.Fatal("schema has no $defs")
	}
	for _, name := range []string{"Catalog", "Entity", "Field"} {
		if _, ok := defs[name]; !ok {
			t.Errorf("schema is missing definition %q", name)
		}
	}
}

// TestEmbeddedSchemaIsCurrent fails when the committed schema has drifted from
// the Go types; run `go generate ./schema` to fix.
func TestEmbeddedSchemaIsCurrent(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("go tool not available")
	}
	cmd := exec.Command(goBin, "run", "gen.go", "-out", "-")
	var out, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("go run gen.go: %v\n%s", err, stderr.String())
	}
	if !bytes.Equal(out.Bytes(), schema.JSON) {
		t.Fatal("datacatalog.schema.json is stale; run `go generate ./schema`")
	}
}
