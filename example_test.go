package datacatalog_test

import (
	"fmt"

	"github.com/grokify/datacatalog"
)

func Example() {
	cat := &datacatalog.Catalog{
		Name: "shop",
		Entities: []datacatalog.Entity{{
			Name:         "account",
			Owner:        "identity",
			SubjectKey:   "id",
			PrivacyScope: true,
			Fields: []datacatalog.Field{
				{Name: "id", Type: datacatalog.TypeString, Visibility: datacatalog.VisibilityPublic, Export: true},
				{Name: "email", Type: datacatalog.TypeString, Visibility: datacatalog.VisibilityPublic,
					PII: true, Erasure: datacatalog.ErasureClear, Export: true},
				{Name: "password_hash", Type: datacatalog.TypeBytes, Visibility: datacatalog.VisibilityInternal,
					Secret: true, Erasure: datacatalog.ErasureShred},
				{Name: "is_staff", Type: datacatalog.TypeBool, Visibility: datacatalog.VisibilityInternal},
			},
		}},
	}
	if err := cat.Validate(); err != nil {
		fmt.Println(err)
		return
	}
	for _, name := range []string{"id", "email", "password_hash", "is_staff"} {
		fmt.Printf("%-13s no PII: %-5v PII ok: %v\n", name,
			cat.AllowedInQuery("account", name, false), cat.AllowedInQuery("account", name, true))
	}
	// Output:
	// id            no PII: true  PII ok: true
	// email         no PII: false PII ok: true
	// password_hash no PII: false PII ok: false
	// is_staff      no PII: false PII ok: false
}
