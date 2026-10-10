# DataCatalog

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/grokify/datacatalog/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/grokify/datacatalog/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/grokify/datacatalog/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/grokify/datacatalog/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/grokify/datacatalog/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/grokify/datacatalog/actions/workflows/go-sast-codeql.yaml
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/datacatalog
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/datacatalog
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/datacatalog/blob/main/LICENSE

A small, dependency-free data catalog for Go: a model that describes which
entities and fields an application stores and how each one is classified.

## Why

Several concerns need the same facts about stored data, and each tends to
re-invent them as scattered tags and allow-lists:

- a **query guard** needs to know which fields user-authored queries may touch;
- **privacy tooling** needs to know which entities hold personal data, how to
  find a data subject's rows, and what to do with each field on erasure;
- **documentation and governance** need an inventory, ownership and retention
  notes.

DataCatalog is the single, validated description they can all read. It is plain
data with a handful of predicates, so the catalog can be authored in Go or JSON,
checked in CI, and consumed by whatever needs it.

## What it is not

The core package depends only on the standard library. It has no database
driver, ORM or query-language dependency. **Adapters for ORMs and query guards
live in their own modules**; they import this one, not the other way round.

## Model

| Type | Purpose |
|------|---------|
| `Catalog` | A named set of entities. |
| `Entity` | One record type: `Name`, `Owner`, `Description`, `SubjectKey`, `PrivacyScope`, `HoldApplicable`, `Versioned`, `Retention`, `Fields`. |
| `Field` | One attribute: `Name`, `Description`, `Type`, `Visibility`, `PII`, `Secret`, `Erasure`, `RetainReason`, `Export`. |

- `Type` is coarse: `string`, `number`, `bool`, `time`, `json`, `bytes`.
- `Visibility` is `public` or `internal` and has **no default**: an unset
  visibility is a validation error, so every field is classified on purpose.
- `Erasure` is one of `clear`, `tombstone`, `hash`, `delete_row`, `shred`,
  `retain`.
- `PrivacyScope`, `HoldApplicable` and `Versioned` are entity-level: they
  describe the record, not a column.

`Validate()` reports **all** problems at once, each with a stable `Rule`:

| Rule | Meaning |
|------|---------|
| `name_required`, `name_unique` | Catalog, entity and field names are non-empty; entity names are unique in a catalog and field names within an entity. |
| `visibility_required` | Visibility is explicitly `public` or `internal`. |
| `secret_internal` | A secret field must be internal. |
| `subject_key_required` / `subject_key_unknown` | A privacy-scope entity with PII needs a `SubjectKey`, and a `SubjectKey` must name an existing field. |
| `pii_erasure_required` | A PII field in a privacy-scope entity needs an erasure strategy. |
| `retain_reason_required` | Erasure `retain` needs a `RetainReason`. |
| `type_invalid`, `erasure_invalid` | Optional enumerations, when set, hold a known value. |

## Example

```go
cat := &datacatalog.Catalog{
    Name: "shop",
    Entities: []datacatalog.Entity{{
        Name:         "account",
        Owner:        "identity",
        SubjectKey:   "id",
        PrivacyScope: true,
        Fields: []datacatalog.Field{
            {Name: "id", Type: datacatalog.TypeString,
                Visibility: datacatalog.VisibilityPublic, Export: true},
            // PII: queryable only when the caller opts in; cleared on erasure.
            {Name: "email", Type: datacatalog.TypeString,
                Visibility: datacatalog.VisibilityPublic,
                PII: true, Erasure: datacatalog.ErasureClear, Export: true},
            // Secret: never queryable, never exported; the key is shredded.
            {Name: "password_hash", Type: datacatalog.TypeBytes,
                Visibility: datacatalog.VisibilityInternal,
                Secret: true, Erasure: datacatalog.ErasureShred},
            // Internal flag: hidden from the application audience.
            {Name: "is_staff", Type: datacatalog.TypeBool,
                Visibility: datacatalog.VisibilityInternal},
        },
    }},
}

if err := cat.Validate(); err != nil {
    log.Fatal(err) // lists every problem
}

cat.AllowedInQuery("account", "id", false)            // true
cat.AllowedInQuery("account", "email", false)         // false: PII
cat.AllowedInQuery("account", "email", true)          // true
cat.AllowedInQuery("account", "password_hash", true)  // false: secret
cat.AllowedInQuery("account", "is_staff", true)       // false: internal
cat.AllowedInQuery("account", "missing", true)        // false: default-deny
```

The same catalog as JSON (entities are written sorted by name so diffs stay
small):

```json
{
  "name": "shop",
  "entities": [
    {
      "name": "account",
      "owner": "identity",
      "subject_key": "id",
      "privacy_scope": true,
      "fields": [
        {"name": "id", "type": "string", "visibility": "public", "export": true},
        {"name": "email", "type": "string", "visibility": "public", "pii": true, "erasure": "clear", "export": true},
        {"name": "password_hash", "type": "bytes", "visibility": "internal", "secret": true, "erasure": "shred"},
        {"name": "is_staff", "type": "bool", "visibility": "internal"}
      ]
    }
  ]
}
```

## Using it

**Query guard.** Give the guard a field policy backed by the catalog. The
predicate is default-deny: unknown entities and fields are rejected, internal
and secret fields are always blocked, and PII is blocked unless the caller is
allowed to see it.

```go
func (g *Guard) FieldAllowed(entity, field string) bool {
    return g.cat.AllowedInQuery(entity, field, g.allowPII)
}
```

**Ent-based (or any ORM) application.** Keep the catalog next to the schema,
validate it in a test, and let a small adapter module check that the catalog and
the generated schema agree (every Ent field is classified, no stale entries).
At runtime, privacy code walks `PrivacyScopeEntities()`, finds a subject's rows
through `SubjectKey`, and applies each PII field's `Erasure` strategy.

**Documentation.** `cat.RenderMarkdown()` produces one section per entity with
a field table, suitable for embedding in a docs site.

**Many services.** `datacatalog.Merge(a, b, c)` combines catalogs; identical
definitions of the same entity collapse and conflicting ones are an error.

## JSON Schema

The document format has a JSON Schema generated from the Go types (the Go
structs are the source of truth). It is embedded and available as
`schema.JSON` from `github.com/grokify/datacatalog/schema`, and committed at
[`schema/datacatalog.schema.json`](schema/datacatalog.schema.json). Regenerate
with `go generate ./schema`; a test fails if the committed file is stale.

## More

See the [guides](docs/guides/) for the validation and erasure reference.

## License

MIT. See [LICENSE](LICENSE).
