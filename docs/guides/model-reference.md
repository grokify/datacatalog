# Model and Validation Reference

## Entity

| Field | Meaning |
|-------|---------|
| `name` | Entity name, unique in the catalog. |
| `owner` | Team or component responsible for the entity. |
| `description` | Free text. |
| `subject_key` | Name of the field that identifies the data subject. |
| `privacy_scope` | The entity is subject to access and erasure requests. |
| `hold_applicable` | Legal holds can apply to its rows. Entity-level. |
| `versioned` | The entity keeps row history. Entity-level. |
| `retention` | Free-text retention note. |
| `fields` | The entity's fields, in declaration order. |

## Field

| Field | Meaning |
|-------|---------|
| `name` | Field name, unique in the entity. |
| `description` | Free text. |
| `type` | Optional coarse type: `string`, `number`, `bool`, `time`, `json`, `bytes`. |
| `visibility` | Required: `public` or `internal`. There is no default. |
| `pii` | Personal data about the data subject. |
| `secret` | Credential or key material. Implies `internal`. |
| `erasure` | Strategy applied on erasure; see [Erasure strategies](erasure.md). |
| `retain_reason` | Required when `erasure` is `retain`. |
| `export` | Included in data-subject access exports. |

## Audiences

| Audience | Sees |
|----------|------|
| `application` | Public, non-secret fields. |
| `internal` | Every field. |

`Field.AllowedInQuery(allowPII)` is stricter and meant for query guards:
internal and secret fields are always blocked, PII fields only when `allowPII`
is true. `Catalog.AllowedInQuery(entity, field, allowPII)` additionally denies
unknown entities and fields.

## Validation

`Catalog.Validate()` returns a `*ValidationError` holding every `Problem`
(entity, field, `Rule`, message), or nil. Use `errors.As` to inspect it.

| Rule | Triggered when |
|------|----------------|
| `name_required` | Catalog, entity or field name is empty. |
| `name_unique` | Two entities share a name, or two fields of one entity do. |
| `visibility_required` | Visibility is unset or not `public`/`internal`. |
| `type_invalid` | A set `type` is not a known type. |
| `erasure_invalid` | A set `erasure` is not a known strategy. |
| `secret_internal` | A secret field is not internal. |
| `subject_key_required` | A privacy-scope entity has PII fields but no `subject_key`. |
| `subject_key_unknown` | `subject_key` does not name a field of the entity. |
| `pii_erasure_required` | A PII field in a privacy-scope entity has no `erasure`. |
| `retain_reason_required` | `erasure` is `retain` without a `retain_reason`. |

## File format

JSON only. Decoding rejects unknown properties, so a misspelled classification
is an error rather than a silent default. Output is stable: entities are sorted
by name; fields keep declaration order. The schema is at
`schema/datacatalog.schema.json`.
