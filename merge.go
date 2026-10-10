package datacatalog

import (
	"errors"
	"fmt"
	"reflect"
)

// Merge combines catalogs into a new one, for example one catalog per
// service into an organization-wide view. The merged catalog takes the name
// of the first catalog that has one. Nil catalogs are skipped.
//
// An entity defined in more than one catalog must be defined identically
// (the duplicate is then collapsed); otherwise Merge returns an error naming
// every conflicting entity. The result has entities sorted by name.
func Merge(catalogs ...*Catalog) (*Catalog, error) {
	out := &Catalog{}
	index := map[string]int{}
	var conflicts []error
	reported := map[string]bool{}
	for _, c := range catalogs {
		if c == nil {
			continue
		}
		if out.Name == "" {
			out.Name = c.Name
		}
		for _, e := range c.Clone().Entities {
			i, ok := index[e.Name]
			switch {
			case !ok:
				index[e.Name] = len(out.Entities)
				out.Entities = append(out.Entities, e)
			case !reflect.DeepEqual(normalizeEntity(out.Entities[i]), normalizeEntity(e)):
				if !reported[e.Name] {
					reported[e.Name] = true
					conflicts = append(conflicts, fmt.Errorf("entity %q is defined differently in more than one catalog", e.Name))
				}
			}
		}
	}
	if len(conflicts) > 0 {
		return nil, fmt.Errorf("datacatalog: merge: %w", errors.Join(conflicts...))
	}
	out.Normalize()
	return out, nil
}

// normalizeEntity makes nil and empty field slices compare equal.
func normalizeEntity(e Entity) Entity {
	if len(e.Fields) == 0 {
		e.Fields = nil
	}
	return e
}
