package datacatalog

import "testing"

const wantMarkdown = "# Data catalog: shop\n" +
	"\n## account\n" +
	"\nA registered user.\n" +
	"\n- Owner: identity\n" +
	"- Subject key: `id`\n" +
	"- Privacy scope: yes\n" +
	"- Hold applicable: no\n" +
	"- Versioned: yes\n" +
	"- Retention: Deleted 30 days after closure.\n" +
	"\n| Field | Type | Visibility | PII | Secret | Erasure | Export | Description |\n" +
	"|---|---|---|---|---|---|---|---|\n" +
	"| `id` | string | public | no | no |  | yes |  |\n" +
	"| `email` | string | public | yes | no | clear | yes | Login \\| contact address. |\n" +
	"| `password_hash` | bytes | internal | no | yes | shred | no |  |\n" +
	"| `is_staff` | bool | internal | no | no |  | no | Staff flag. |\n" +
	"| `tax_id` | string | internal | yes | no | retain (Tax law) | no |  |\n" +
	"\n## invoice\n" +
	"\n- Privacy scope: no\n" +
	"- Hold applicable: yes\n" +
	"- Versioned: no\n" +
	"\n| Field | Type | Visibility | PII | Secret | Erasure | Export | Description |\n" +
	"|---|---|---|---|---|---|---|---|\n" +
	"| `id` |  | public | no | no |  | no |  |\n" +
	"| `total` | number | public | no | no |  | no |  |\n"

func TestRenderMarkdownGolden(t *testing.T) {
	c := exampleCatalog()
	// Input order must not matter.
	c.Entities[0], c.Entities[1] = c.Entities[1], c.Entities[0]
	if got := c.RenderMarkdown(); got != wantMarkdown {
		t.Errorf("markdown mismatch\n--- got ---\n%s\n--- want ---\n%s", got, wantMarkdown)
	}
}

func TestRenderMarkdownEdgeCases(t *testing.T) {
	var nilCat *Catalog
	if got := nilCat.RenderMarkdown(); got != "" {
		t.Errorf("nil catalog: %q", got)
	}
	c := &Catalog{Name: "x", Entities: []Entity{{Name: "empty", Description: "multi\nline"}}}
	want := "# Data catalog: x\n\n## empty\n\nmulti line\n\n- Privacy scope: no\n- Hold applicable: no\n- Versioned: no\n"
	if got := c.RenderMarkdown(); got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
