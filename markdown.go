package datacatalog

import (
	"fmt"
	"strings"
)

// RenderMarkdown renders the catalog as Markdown documentation: one section
// per entity (sorted by name) with a field table. The receiver is not
// modified and output is deterministic.
func (c *Catalog) RenderMarkdown() string {
	if c == nil {
		return ""
	}
	n := c.Clone()
	n.Normalize()

	var b strings.Builder
	fmt.Fprintf(&b, "# Data catalog: %s\n", mdInline(n.Name))
	for _, e := range n.Entities {
		fmt.Fprintf(&b, "\n## %s\n", mdInline(e.Name))
		if e.Description != "" {
			fmt.Fprintf(&b, "\n%s\n", mdInline(e.Description))
		}
		if meta := entityMeta(e); len(meta) > 0 {
			b.WriteString("\n")
			for _, m := range meta {
				fmt.Fprintf(&b, "- %s\n", m)
			}
		}
		if len(e.Fields) == 0 {
			continue
		}
		b.WriteString("\n| Field | Type | Visibility | PII | Secret | Erasure | Export | Description |\n")
		b.WriteString("|---|---|---|---|---|---|---|---|\n")
		for _, f := range e.Fields {
			erasure := string(f.Erasure)
			if f.RetainReason != "" {
				erasure += fmt.Sprintf(" (%s)", mdCell(f.RetainReason))
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s | %s | %s | %s | %s |\n",
				f.Name, string(f.Type), string(f.Visibility),
				yesNo(f.PII), yesNo(f.Secret), erasure, yesNo(f.Export), mdCell(f.Description))
		}
	}
	return b.String()
}

func entityMeta(e Entity) []string {
	var m []string
	if e.Owner != "" {
		m = append(m, "Owner: "+mdInline(e.Owner))
	}
	if e.SubjectKey != "" {
		m = append(m, fmt.Sprintf("Subject key: `%s`", e.SubjectKey))
	}
	m = append(m, "Privacy scope: "+yesNo(e.PrivacyScope))
	m = append(m, "Hold applicable: "+yesNo(e.HoldApplicable))
	m = append(m, "Versioned: "+yesNo(e.Versioned))
	if e.Retention != "" {
		m = append(m, "Retention: "+mdInline(e.Retention))
	}
	return m
}

func yesNo(v bool) string {
	if v {
		return "yes"
	}
	return "no"
}

// mdInline collapses whitespace so a value stays on one line.
func mdInline(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// mdCell makes a value safe for a table cell.
func mdCell(s string) string {
	return strings.ReplaceAll(mdInline(s), "|", `\|`)
}
