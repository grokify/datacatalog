package datacatalog

// Type is the coarse storage type of a field. It is descriptive, not a
// database type: adapters map it to whatever the engine needs.
type Type string

// Field types.
const (
	TypeString Type = "string"
	TypeNumber Type = "number"
	TypeBool   Type = "bool"
	TypeTime   Type = "time"
	TypeJSON   Type = "json"
	TypeBytes  Type = "bytes"
)

// Valid reports whether t is a known type. The empty Type is not valid;
// callers that treat the type as optional check for "" first.
func (t Type) Valid() bool {
	switch t {
	case TypeString, TypeNumber, TypeBool, TypeTime, TypeJSON, TypeBytes:
		return true
	}
	return false
}

// Visibility says who may see a field. There is deliberately no default: the
// zero value is invalid so that every field is classified explicitly.
type Visibility string

// Field visibilities.
const (
	// VisibilityPublic fields may be exposed to application code and
	// user-facing queries.
	VisibilityPublic Visibility = "public"
	// VisibilityInternal fields are for the platform only.
	VisibilityInternal Visibility = "internal"
)

// Valid reports whether v is an explicit, known visibility.
func (v Visibility) Valid() bool {
	return v == VisibilityPublic || v == VisibilityInternal
}

// Erasure is the strategy applied to a field when its data subject is erased.
type Erasure string

// Erasure strategies.
const (
	// ErasureClear sets the value to its zero value or NULL.
	ErasureClear Erasure = "clear"
	// ErasureTombstone replaces the value with a fixed marker.
	ErasureTombstone Erasure = "tombstone"
	// ErasureHash replaces the value with a one-way hash.
	ErasureHash Erasure = "hash"
	// ErasureDeleteRow removes the whole row.
	ErasureDeleteRow Erasure = "delete_row"
	// ErasureShred destroys the key that protects the value (crypto-shredding).
	ErasureShred Erasure = "shred"
	// ErasureRetain keeps the value; RetainReason must say why.
	ErasureRetain Erasure = "retain"
)

// Valid reports whether e is a known strategy. The empty Erasure is not
// valid; it means "no strategy declared".
func (e Erasure) Valid() bool {
	switch e {
	case ErasureClear, ErasureTombstone, ErasureHash, ErasureDeleteRow, ErasureShred, ErasureRetain:
		return true
	}
	return false
}

// Field describes one stored attribute of an Entity.
type Field struct {
	Name        string     `json:"name" jsonschema_description:"Field (column) name; unique within the entity."`
	Description string     `json:"description,omitempty"`
	Type        Type       `json:"type,omitempty" jsonschema:"enum=string,enum=number,enum=bool,enum=time,enum=json,enum=bytes"`
	Visibility  Visibility `json:"visibility" jsonschema:"enum=public,enum=internal" jsonschema_description:"Explicit audience classification; there is no default."`
	// PII marks personal data about the data subject.
	PII bool `json:"pii,omitempty" jsonschema_description:"PII marks personal data about the data subject."`
	// Secret marks credentials and key material; it implies internal visibility.
	Secret bool `json:"secret,omitempty" jsonschema_description:"Secret marks credentials and key material; it implies internal visibility."`
	// Erasure is the strategy applied when the data subject is erased.
	Erasure Erasure `json:"erasure,omitempty" jsonschema:"enum=clear,enum=tombstone,enum=hash,enum=delete_row,enum=shred,enum=retain" jsonschema_description:"Erasure is the strategy applied when the data subject is erased."`
	// RetainReason is required when Erasure is "retain".
	RetainReason string `json:"retain_reason,omitempty" jsonschema_description:"RetainReason is required when Erasure is 'retain'."`
	// Export includes the field in data-subject access exports.
	Export bool `json:"export,omitempty" jsonschema_description:"Export includes the field in data-subject access exports."`
}

// Entity describes one stored record type (a table, collection or document
// kind) and its fields.
type Entity struct {
	Name        string `json:"name" jsonschema_description:"Entity name; unique within the catalog."`
	Owner       string `json:"owner,omitempty" jsonschema_description:"Team or component responsible for the entity."`
	Description string `json:"description,omitempty"`
	// SubjectKey names the field that identifies the data subject.
	SubjectKey string `json:"subject_key,omitempty" jsonschema_description:"SubjectKey names the field that identifies the data subject."`
	// PrivacyScope marks the entity as subject to access and erasure requests.
	PrivacyScope bool `json:"privacy_scope,omitempty" jsonschema_description:"PrivacyScope marks the entity as subject to access and erasure requests."`
	// HoldApplicable marks the entity as subject to legal holds. It is an
	// entity-level property: holds suspend erasure for whole rows.
	HoldApplicable bool `json:"hold_applicable,omitempty" jsonschema_description:"HoldApplicable marks the entity as subject to legal holds. It is an entity-level property: holds suspend erasure for whole rows."`
	// Versioned marks the entity as keeping row history. It is an
	// entity-level property: history rows must be erased with their subject.
	Versioned bool `json:"versioned,omitempty" jsonschema_description:"Versioned marks the entity as keeping row history. It is an entity-level property: history rows must be erased with their subject."`
	// Retention is a free-text retention note.
	Retention string  `json:"retention,omitempty" jsonschema_description:"Retention is a free-text retention note."`
	Fields    []Field `json:"fields,omitempty"`
}

// Catalog is a named set of entities.
type Catalog struct {
	Name     string   `json:"name" jsonschema_description:"Catalog name, typically the application or service."`
	Entities []Entity `json:"entities,omitempty"`
}

// HasPII reports whether any field of the entity is PII.
func (e Entity) HasPII() bool {
	for _, f := range e.Fields {
		if f.PII {
			return true
		}
	}
	return false
}
