package datacatalog

// Audience is who is looking at the catalog or issuing a query.
type Audience string

// Audiences.
const (
	// Application is application code and user-facing queries. It sees only
	// public, non-secret fields.
	Application Audience = "application"
	// Internal is platform code. It sees every field.
	Internal Audience = "internal"
)

// VisibleTo reports whether the field is visible to the audience. Secret
// fields are never visible to Application, even if misclassified as public.
func (f Field) VisibleTo(aud Audience) bool {
	switch aud {
	case Internal:
		return true
	case Application:
		return f.Visibility == VisibilityPublic && !f.Secret
	}
	return false
}

// AllowedInQuery reports whether a query guard may let the field appear in a
// user-authored query. Internal and secret fields are always blocked; PII
// fields are blocked unless allowPII is true.
func (f Field) AllowedInQuery(allowPII bool) bool {
	if f.Secret || f.Visibility != VisibilityPublic {
		return false
	}
	return allowPII || !f.PII
}

// VisibleFields returns the fields visible to aud, in declaration order.
func (e Entity) VisibleFields(aud Audience) []Field {
	var out []Field
	for _, f := range e.Fields {
		if f.VisibleTo(aud) {
			out = append(out, f)
		}
	}
	return out
}

// QueryableFields returns the fields for which AllowedInQuery(allowPII) holds.
func (e Entity) QueryableFields(allowPII bool) []Field {
	var out []Field
	for _, f := range e.Fields {
		if f.AllowedInQuery(allowPII) {
			out = append(out, f)
		}
	}
	return out
}

// PIIFields returns the PII fields in declaration order.
func (e Entity) PIIFields() []Field {
	var out []Field
	for _, f := range e.Fields {
		if f.PII {
			out = append(out, f)
		}
	}
	return out
}

// Field looks up a field by name.
func (e Entity) Field(name string) (Field, bool) {
	for _, f := range e.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return Field{}, false
}

// Entity looks up an entity by name.
func (c *Catalog) Entity(name string) (Entity, bool) {
	if c == nil {
		return Entity{}, false
	}
	for _, e := range c.Entities {
		if e.Name == name {
			return e, true
		}
	}
	return Entity{}, false
}

// Field looks up a field by entity and field name.
func (c *Catalog) Field(entity, name string) (Field, bool) {
	e, ok := c.Entity(entity)
	if !ok {
		return Field{}, false
	}
	return e.Field(name)
}

// AllowedInQuery is the default-deny lookup for a query guard: unknown
// entities and fields are not allowed.
func (c *Catalog) AllowedInQuery(entity, field string, allowPII bool) bool {
	f, ok := c.Field(entity, field)
	return ok && f.AllowedInQuery(allowPII)
}

// PrivacyScopeEntities returns the entities subject to access and erasure
// requests, in catalog order.
func (c *Catalog) PrivacyScopeEntities() []Entity {
	var out []Entity
	if c == nil {
		return out
	}
	for _, e := range c.Entities {
		if e.PrivacyScope {
			out = append(out, e)
		}
	}
	return out
}
