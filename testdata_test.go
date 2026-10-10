package datacatalog

// exampleCatalog is a valid catalog used across tests.
func exampleCatalog() *Catalog {
	return &Catalog{
		Name: "shop",
		Entities: []Entity{
			{
				Name:         "account",
				Owner:        "identity",
				Description:  "A registered user.",
				SubjectKey:   "id",
				PrivacyScope: true,
				Versioned:    true,
				Retention:    "Deleted 30 days after closure.",
				Fields: []Field{
					{Name: "id", Type: TypeString, Visibility: VisibilityPublic, Export: true},
					{Name: "email", Type: TypeString, Visibility: VisibilityPublic, PII: true, Erasure: ErasureClear, Export: true, Description: "Login | contact address."},
					{Name: "password_hash", Type: TypeBytes, Visibility: VisibilityInternal, Secret: true, Erasure: ErasureShred},
					{Name: "is_staff", Type: TypeBool, Visibility: VisibilityInternal, Description: "Staff flag."},
					{Name: "tax_id", Type: TypeString, Visibility: VisibilityInternal, PII: true, Erasure: ErasureRetain, RetainReason: "Tax law"},
				},
			},
			{
				Name:           "invoice",
				HoldApplicable: true,
				Fields: []Field{
					{Name: "id", Visibility: VisibilityPublic},
					{Name: "total", Type: TypeNumber, Visibility: VisibilityPublic},
				},
			},
		},
	}
}
