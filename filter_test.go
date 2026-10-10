package datacatalog

import (
	"reflect"
	"testing"
)

func names(fs []Field) []string {
	out := []string{}
	for _, f := range fs {
		out = append(out, f.Name)
	}
	return out
}

func TestFieldVisibleTo(t *testing.T) {
	tests := []struct {
		name  string
		field Field
		aud   Audience
		want  bool
	}{
		{"public to application", Field{Visibility: VisibilityPublic}, Application, true},
		{"internal to application", Field{Visibility: VisibilityInternal}, Application, false},
		{"misclassified secret to application", Field{Visibility: VisibilityPublic, Secret: true}, Application, false},
		{"unset visibility to application", Field{}, Application, false},
		{"internal to internal", Field{Visibility: VisibilityInternal}, Internal, true},
		{"secret to internal", Field{Visibility: VisibilityInternal, Secret: true}, Internal, true},
		{"unknown audience", Field{Visibility: VisibilityPublic}, Audience("other"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.field.VisibleTo(tt.aud); got != tt.want {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestFieldAllowedInQuery(t *testing.T) {
	tests := []struct {
		name     string
		field    Field
		allowPII bool
		want     bool
	}{
		{"public", Field{Visibility: VisibilityPublic}, false, true},
		{"public with allowPII", Field{Visibility: VisibilityPublic}, true, true},
		{"internal", Field{Visibility: VisibilityInternal}, true, false},
		{"secret", Field{Visibility: VisibilityInternal, Secret: true}, true, false},
		{"secret misclassified public", Field{Visibility: VisibilityPublic, Secret: true}, true, false},
		{"pii blocked", Field{Visibility: VisibilityPublic, PII: true}, false, false},
		{"pii allowed", Field{Visibility: VisibilityPublic, PII: true}, true, true},
		{"internal pii", Field{Visibility: VisibilityInternal, PII: true}, true, false},
		{"unset visibility", Field{}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.field.AllowedInQuery(tt.allowPII); got != tt.want {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
}

func TestEntityFieldLists(t *testing.T) {
	c := exampleCatalog()
	acct, ok := c.Entity("account")
	if !ok {
		t.Fatal("account not found")
	}
	tests := []struct {
		name string
		got  []Field
		want []string
	}{
		{"application", acct.VisibleFields(Application), []string{"id", "email"}},
		{"internal", acct.VisibleFields(Internal), []string{"id", "email", "password_hash", "is_staff", "tax_id"}},
		{"queryable no pii", acct.QueryableFields(false), []string{"id"}},
		{"queryable pii", acct.QueryableFields(true), []string{"id", "email"}},
		{"pii", acct.PIIFields(), []string{"email", "tax_id"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := names(tt.got); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v want %v", got, tt.want)
			}
		})
	}
	if !acct.HasPII() {
		t.Error("account should have PII")
	}
	if inv, _ := c.Entity("invoice"); inv.HasPII() {
		t.Error("invoice should not have PII")
	}
}

func TestLookups(t *testing.T) {
	c := exampleCatalog()
	if f, ok := c.Field("account", "email"); !ok || !f.PII {
		t.Errorf("email lookup failed: %+v %v", f, ok)
	}
	for _, tc := range [][2]string{{"nope", "id"}, {"account", "nope"}} {
		if _, ok := c.Field(tc[0], tc[1]); ok {
			t.Errorf("lookup %v should fail", tc)
		}
	}
	var nilCat *Catalog
	if _, ok := nilCat.Entity("account"); ok {
		t.Error("nil catalog lookup should fail")
	}
	if got := nilCat.PrivacyScopeEntities(); len(got) != 0 {
		t.Errorf("nil catalog scope entities: %v", got)
	}
}

func TestCatalogAllowedInQuery(t *testing.T) {
	c := exampleCatalog()
	tests := []struct {
		entity, field string
		allowPII      bool
		want          bool
	}{
		{"account", "id", false, true},
		{"account", "email", false, false},
		{"account", "email", true, true},
		{"account", "password_hash", true, false},
		{"account", "is_staff", true, false},
		{"account", "missing", true, false},
		{"missing", "id", true, false},
	}
	for _, tt := range tests {
		if got := c.AllowedInQuery(tt.entity, tt.field, tt.allowPII); got != tt.want {
			t.Errorf("%s.%s allowPII=%v: got %v want %v", tt.entity, tt.field, tt.allowPII, got, tt.want)
		}
	}
}

func TestPrivacyScopeEntities(t *testing.T) {
	got := exampleCatalog().PrivacyScopeEntities()
	if len(got) != 1 || got[0].Name != "account" {
		t.Errorf("got %+v", got)
	}
}
