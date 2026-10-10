# Erasure Strategies

The `erasure` of a field says what privacy tooling does with its value when the
data subject is erased.

| Strategy | Effect |
|----------|--------|
| `clear` | Set the value to its zero value or NULL. |
| `tombstone` | Replace the value with a fixed marker, keeping the row shape. |
| `hash` | Replace the value with a one-way hash, e.g. to keep uniqueness. |
| `delete_row` | Remove the whole row. |
| `shred` | Destroy the key that protects the value (crypto-shredding). |
| `retain` | Keep the value. A `retain_reason` is required. |

The catalog records the strategy; it does not execute it. Executing erasure is
the job of privacy tooling built on the catalog.

Entity-level flags interact with erasure: a `hold_applicable` entity may need
erasure suspended while a hold is in force, and a `versioned` entity keeps
history rows that must be treated with the same strategy as current rows.
