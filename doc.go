// Package datacatalog is a small, dependency-free data catalog: a model for
// describing which entities and fields an application stores and how each
// field is classified (visibility, PII, secret, erasure strategy).
//
// The catalog is plain data. It is meant to be authored (in Go or JSON),
// validated once, and then consumed by query guards, privacy tooling
// (access and erasure requests), documentation generators and management
// tooling. Adapters for ORMs and query guards live in their own modules; this
// package depends only on the standard library.
//
// Typical use:
//
//	cat, err := datacatalog.ReadFile("catalog.json")
//	if err != nil { ... }
//	if err := cat.Validate(); err != nil { ... } // reports every problem
//	if cat.AllowedInQuery("account", "email", false) { ... }
package datacatalog
