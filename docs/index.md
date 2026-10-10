# DataCatalog

A small, dependency-free data catalog for Go: a model that describes which
entities and fields an application stores and how each one is classified, for
use by query guards, privacy tooling and documentation generators.

- [Model and validation reference](guides/model-reference.md)
- [Erasure strategies](guides/erasure.md)

The package depends only on the standard library. Adapters for ORMs and query
guards live in their own modules.
