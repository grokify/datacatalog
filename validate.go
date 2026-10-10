package datacatalog

import (
	"errors"
	"fmt"
	"strings"
)

// Rule identifies a validation rule so callers can react to specific
// problems without parsing messages.
type Rule string

// Validation rules.
const (
	RuleNameRequired        Rule = "name_required"
	RuleNameUnique          Rule = "name_unique"
	RuleVisibilityRequired  Rule = "visibility_required"
	RuleTypeInvalid         Rule = "type_invalid"
	RuleErasureInvalid      Rule = "erasure_invalid"
	RuleSecretInternal      Rule = "secret_internal"
	RuleSubjectKeyRequired  Rule = "subject_key_required"
	RuleSubjectKeyUnknown   Rule = "subject_key_unknown"
	RulePIIErasureRequired  Rule = "pii_erasure_required"
	RuleRetainReasonMissing Rule = "retain_reason_required"
)

// Problem is a single validation finding. Entity and Field are empty when the
// problem is not about a specific entity or field.
type Problem struct {
	Entity  string `json:"entity,omitempty"`
	Field   string `json:"field,omitempty"`
	Rule    Rule   `json:"rule"`
	Message string `json:"message"`
}

// String renders the problem as "entity.field: message (rule)".
func (p Problem) String() string {
	loc := p.Entity
	if p.Field != "" {
		loc += "." + p.Field
	}
	if loc != "" {
		loc += ": "
	}
	return fmt.Sprintf("%s%s (%s)", loc, p.Message, p.Rule)
}

// ValidationError carries every problem found by Validate.
type ValidationError struct {
	Problems []Problem
}

// Error lists all problems, one per line.
func (e *ValidationError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "datacatalog: %d validation problem(s)", len(e.Problems))
	for _, p := range e.Problems {
		b.WriteString("\n  - ")
		b.WriteString(p.String())
	}
	return b.String()
}

// Has reports whether any problem matches the rule.
func (e *ValidationError) Has(r Rule) bool {
	for _, p := range e.Problems {
		if p.Rule == r {
			return true
		}
	}
	return false
}

// Validate checks the catalog against every rule and returns a
// *ValidationError listing all problems, or nil when the catalog is valid.
func (c *Catalog) Validate() error {
	if c == nil {
		return errors.New("datacatalog: nil catalog")
	}
	var probs []Problem
	if c.Name == "" {
		probs = append(probs, Problem{Rule: RuleNameRequired, Message: "catalog name is required"})
	}
	seen := map[string]bool{}
	for i, e := range c.Entities {
		switch {
		case e.Name == "":
			probs = append(probs, Problem{
				Rule:    RuleNameRequired,
				Message: fmt.Sprintf("entity at index %d has no name", i),
			})
		case seen[e.Name]:
			probs = append(probs, Problem{Entity: e.Name, Rule: RuleNameUnique, Message: "duplicate entity name"})
		}
		seen[e.Name] = true
		probs = append(probs, e.problems()...)
	}
	if len(probs) == 0 {
		return nil
	}
	return &ValidationError{Problems: probs}
}

// Validate checks a single entity against the entity and field rules.
func (e Entity) Validate() error {
	probs := e.problems()
	if len(probs) == 0 {
		return nil
	}
	return &ValidationError{Problems: probs}
}

func (e Entity) problems() []Problem {
	var probs []Problem
	add := func(field string, r Rule, format string, args ...any) {
		probs = append(probs, Problem{Entity: e.Name, Field: field, Rule: r, Message: fmt.Sprintf(format, args...)})
	}

	names := map[string]bool{}
	for i, f := range e.Fields {
		if f.Name == "" {
			add("", RuleNameRequired, "field at index %d has no name", i)
		} else if names[f.Name] {
			add(f.Name, RuleNameUnique, "duplicate field name")
		}
		names[f.Name] = true

		switch {
		case f.Visibility == "":
			add(f.Name, RuleVisibilityRequired, "visibility must be set explicitly to %q or %q",
				VisibilityPublic, VisibilityInternal)
		case !f.Visibility.Valid():
			add(f.Name, RuleVisibilityRequired, "unknown visibility %q", f.Visibility)
		}
		if f.Type != "" && !f.Type.Valid() {
			add(f.Name, RuleTypeInvalid, "unknown type %q", f.Type)
		}
		if f.Erasure != "" && !f.Erasure.Valid() {
			add(f.Name, RuleErasureInvalid, "unknown erasure strategy %q", f.Erasure)
		}
		if f.Secret && f.Visibility != VisibilityInternal {
			add(f.Name, RuleSecretInternal, "secret field must have visibility %q", VisibilityInternal)
		}
		if f.Erasure == ErasureRetain && strings.TrimSpace(f.RetainReason) == "" {
			add(f.Name, RuleRetainReasonMissing, "erasure %q requires a retain_reason", ErasureRetain)
		}
		if e.PrivacyScope && f.PII && f.Erasure == "" {
			add(f.Name, RulePIIErasureRequired, "PII field in a privacy-scope entity needs an erasure strategy")
		}
	}

	switch {
	case e.SubjectKey != "" && !names[e.SubjectKey]:
		add("", RuleSubjectKeyUnknown, "subject_key %q does not name a field of the entity", e.SubjectKey)
	case e.SubjectKey == "" && e.PrivacyScope && e.HasPII():
		add("", RuleSubjectKeyRequired, "privacy-scope entity with PII fields needs a subject_key")
	}
	return probs
}
