package datacatalog

import (
	"strings"
	"testing"
)

func TestMerge(t *testing.T) {
	a := exampleCatalog()
	b := &Catalog{Name: "other", Entities: []Entity{
		exampleCatalog().Entities[1], // identical to a's invoice
		{Name: "audit", Fields: []Field{{Name: "id", Visibility: VisibilityInternal}}},
	}}
	got, err := Merge(nil, a, b)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "shop" {
		t.Errorf("name = %q", got.Name)
	}
	var names []string
	for _, e := range got.Entities {
		names = append(names, e.Name)
	}
	if strings.Join(names, ",") != "account,audit,invoice" {
		t.Errorf("entities = %v", names)
	}
	if err := got.Validate(); err != nil {
		t.Errorf("merged catalog invalid: %v", err)
	}
	got.Entities[0].Fields[0].Name = "changed"
	if a.Entities[0].Fields[0].Name == "changed" {
		t.Error("merge result aliases its input")
	}
}

func TestMergeNameFromFirstNamed(t *testing.T) {
	got, err := Merge(&Catalog{}, &Catalog{Name: "b"}, &Catalog{Name: "c"})
	if err != nil || got.Name != "b" {
		t.Fatalf("got %+v %v", got, err)
	}
	empty, err := Merge()
	if err != nil || empty.Name != "" || len(empty.Entities) != 0 {
		t.Fatalf("empty merge: %+v %v", empty, err)
	}
}

func TestMergeConflicts(t *testing.T) {
	base := exampleCatalog()
	tests := []struct {
		name   string
		mutate func(c *Catalog)
	}{
		{"owner differs", func(c *Catalog) { c.Entities[0].Owner = "billing" }},
		{"field differs", func(c *Catalog) { c.Entities[0].Fields[1].PII = false }},
		{"field added", func(c *Catalog) {
			c.Entities[0].Fields = append(c.Entities[0].Fields, Field{Name: "x", Visibility: VisibilityPublic})
		}},
		{"hold differs", func(c *Catalog) { c.Entities[1].HoldApplicable = false }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			other := exampleCatalog()
			tt.mutate(other)
			_, err := Merge(base, other)
			if err == nil || !strings.Contains(err.Error(), "defined differently") {
				t.Fatalf("want conflict error, got %v", err)
			}
		})
	}
}

func TestMergeReportsEveryConflictOnce(t *testing.T) {
	a := exampleCatalog()
	b := exampleCatalog()
	c := exampleCatalog()
	b.Entities[0].Owner, b.Entities[1].Retention = "x", "y"
	c.Entities[0].Owner = "z"
	_, err := Merge(a, b, c)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if strings.Count(msg, `"account"`) != 1 || !strings.Contains(msg, `"invoice"`) {
		t.Errorf("unexpected error text: %s", msg)
	}
}

func TestMergeNilAndEmptyFieldsEqual(t *testing.T) {
	a := &Catalog{Name: "a", Entities: []Entity{{Name: "e"}}}
	b := &Catalog{Name: "b", Entities: []Entity{{Name: "e", Fields: []Field{}}}}
	if _, err := Merge(a, b); err != nil {
		t.Fatal(err)
	}
}
