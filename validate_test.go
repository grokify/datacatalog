package datacatalog

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateOK(t *testing.T) {
	if err := exampleCatalog().Validate(); err != nil {
		t.Fatalf("example catalog should be valid: %v", err)
	}
}

func TestValidateNilCatalog(t *testing.T) {
	var c *Catalog
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for nil catalog")
	}
}

func TestValidateRules(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(c *Catalog)
		want   []Rule
	}{
		{"catalog name empty", func(c *Catalog) { c.Name = "" }, []Rule{RuleNameRequired}},
		{"entity name empty", func(c *Catalog) { c.Entities[1].Name = "" }, []Rule{RuleNameRequired}},
		{"entity name duplicate", func(c *Catalog) { c.Entities[1].Name = "account" }, []Rule{RuleNameUnique}},
		{"field name empty", func(c *Catalog) { c.Entities[1].Fields[0].Name = "" }, []Rule{RuleNameRequired}},
		{"field name duplicate", func(c *Catalog) { c.Entities[1].Fields[1].Name = "id" }, []Rule{RuleNameUnique}},
		{"visibility unset", func(c *Catalog) { c.Entities[1].Fields[0].Visibility = "" }, []Rule{RuleVisibilityRequired}},
		{"visibility unknown", func(c *Catalog) { c.Entities[1].Fields[0].Visibility = "private" }, []Rule{RuleVisibilityRequired}},
		{"type unknown", func(c *Catalog) { c.Entities[1].Fields[1].Type = "decimal" }, []Rule{RuleTypeInvalid}},
		{"erasure unknown", func(c *Catalog) { c.Entities[0].Fields[1].Erasure = "burn" }, []Rule{RuleErasureInvalid}},
		{"secret but public", func(c *Catalog) { c.Entities[0].Fields[2].Visibility = VisibilityPublic }, []Rule{RuleSecretInternal}},
		{"secret and visibility unset", func(c *Catalog) { c.Entities[0].Fields[2].Visibility = "" }, []Rule{RuleVisibilityRequired, RuleSecretInternal}},
		{"retain without reason", func(c *Catalog) { c.Entities[0].Fields[4].RetainReason = " " }, []Rule{RuleRetainReasonMissing}},
		{"subject key missing", func(c *Catalog) { c.Entities[0].SubjectKey = "" }, []Rule{RuleSubjectKeyRequired}},
		{"subject key unknown", func(c *Catalog) { c.Entities[0].SubjectKey = "nope" }, []Rule{RuleSubjectKeyUnknown}},
		{"subject key unknown without privacy scope", func(c *Catalog) { c.Entities[1].SubjectKey = "nope" }, []Rule{RuleSubjectKeyUnknown}},
		{"pii without erasure", func(c *Catalog) { c.Entities[0].Fields[1].Erasure = "" }, []Rule{RulePIIErasureRequired}},
		{
			"several problems reported together",
			func(c *Catalog) {
				c.Name = ""
				c.Entities[0].SubjectKey = ""
				c.Entities[0].Fields[1].Erasure = ""
				c.Entities[1].Fields[0].Visibility = ""
			},
			[]Rule{RuleNameRequired, RuleVisibilityRequired, RulePIIErasureRequired, RuleSubjectKeyRequired},
		},
		{
			"pii without privacy scope needs nothing",
			func(c *Catalog) {
				c.Entities[0].PrivacyScope = false
				c.Entities[0].SubjectKey = ""
				c.Entities[0].Fields[1].Erasure = ""
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := exampleCatalog()
			tt.mutate(c)
			err := c.Validate()
			if len(tt.want) == 0 {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("want *ValidationError, got %v", err)
			}
			got := map[Rule]int{}
			for _, p := range ve.Problems {
				got[p.Rule]++
			}
			if len(ve.Problems) != len(tt.want) {
				t.Fatalf("want %d problems %v, got %d: %v", len(tt.want), tt.want, len(ve.Problems), err)
			}
			for _, r := range tt.want {
				if got[r] == 0 || !ve.Has(r) {
					t.Errorf("missing rule %s in %v", r, err)
				}
			}
		})
	}
}

func TestValidateErrorText(t *testing.T) {
	c := exampleCatalog()
	c.Entities[0].Fields[1].Erasure = ""
	c.Entities[1].Name = ""
	err := c.Validate()
	msg := err.Error()
	for _, want := range []string{"2 validation problem(s)", "account.email: ", "(pii_erasure_required)", "(name_required)"} {
		if !strings.Contains(msg, want) {
			t.Errorf("error %q missing %q", msg, want)
		}
	}
	var ve *ValidationError
	if errors.As(err, &ve) && ve.Has(RuleSecretInternal) {
		t.Error("Has reported a rule that is not present")
	}
}

func TestEntityValidate(t *testing.T) {
	e := exampleCatalog().Entities[0]
	if err := e.Validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	e.Fields[2].Visibility = VisibilityPublic
	if err := e.Validate(); err == nil {
		t.Fatal("expected secret_internal problem")
	}
}

func TestProblemString(t *testing.T) {
	tests := []struct {
		p    Problem
		want string
	}{
		{Problem{Rule: RuleNameRequired, Message: "m"}, "m (name_required)"},
		{Problem{Entity: "e", Rule: RuleNameUnique, Message: "m"}, "e: m (name_unique)"},
		{Problem{Entity: "e", Field: "f", Rule: RuleTypeInvalid, Message: "m"}, "e.f: m (type_invalid)"},
	}
	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("got %q want %q", got, tt.want)
		}
	}
}

func TestEnumValidity(t *testing.T) {
	for _, v := range []Type{TypeString, TypeNumber, TypeBool, TypeTime, TypeJSON, TypeBytes} {
		if !v.Valid() {
			t.Errorf("type %q should be valid", v)
		}
	}
	for _, v := range []Erasure{ErasureClear, ErasureTombstone, ErasureHash, ErasureDeleteRow, ErasureShred, ErasureRetain} {
		if !v.Valid() {
			t.Errorf("erasure %q should be valid", v)
		}
	}
	if Type("").Valid() || Erasure("").Valid() || Visibility("").Valid() {
		t.Error("zero values must not be valid")
	}
}
