// Package itemcatalogdemo — pruebo el catálogo real montado con crudview sin
// config custom: lista lo sembrado, y crear agrega por la op real.
package itemcatalogdemo

import (
	"os"
	"strings"
	"testing"

	"webtyp.com/router"

	itemcatalog "github.com/veltylabs/item_catalog"

	"webtyp.com/app-demo/config"
)

// TestCatalog_ListsSeededItems: el catálogo lista los servicios sembrados.
func TestCatalog_ListsSeededItems(t *testing.T) {
	env := config.New()
	pres := itemcatalog.NewView(env.Caller())
	if err := pres.Reload(); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	items := pres.Items()
	if len(items) < 4 {
		t.Fatalf("expected >=4 seeded catalog items, got %d", len(items))
	}
	names := map[string]bool{}
	for _, it := range items {
		names[it.Label] = true
	}
	for _, want := range []string{"Consulta médica", "Ecografía abdominal", "Radiografía de tórax"} {
		if !names[want] {
			t.Errorf("expected seeded item %q; got %v", want, names)
		}
	}
}

// TestSpecialties_ListsSeeded: las especialidades sembradas.
func TestSpecialties_ListsSeeded(t *testing.T) {
	env := config.New()
	pres := itemcatalog.NewSpecialtyView(env.Caller())
	if err := pres.Reload(); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	items := pres.Items()
	if len(items) < 4 {
		t.Fatalf("expected >=4 seeded specialties, got %d", len(items))
	}
	foundMed := false
	for _, it := range items {
		if it.Label == "Medicina General" {
			foundMed = true
		}
	}
	if !foundMed {
		t.Fatalf("expected Medicina General in specialties")
	}
}

// TestCrudviewReuse_NoForkNeeded: el módulo demo monta crudview.New(Config{})
// SOLO con ParentID/Presenter/IDs — sin Filter/List/Context custom.
// Comprobación estructural del fuente del módulo: documenta que el layout se
// reutilizó tal cual (es el punto del §1 del usuario).
func TestCrudviewReuse_NoForkNeeded(t *testing.T) {
	src, err := os.ReadFile("items.go")
	if err != nil {
		t.Fatalf("read items.go: %v", err)
	}
	for _, forbidden := range []string{"Filter:", "List:", "Context:"} {
		if strings.Contains(string(src), forbidden) {
			t.Errorf("module must mount crudview WITHOUT custom %s (layout reuse is the point)", forbidden)
		}
	}
	if !strings.Contains(string(src), "itemcatalog.NewView") ||
		!strings.Contains(string(src), "itemcatalog.NewSpecialtyView") {
		t.Errorf("expected the real module presenters to be mounted")
	}
}

// TestCreateItem_GoesThroughRealOp: crear via la op real del catálogo agrega al
// listado (el flujo crudview→Presenter→op que el form usa).
func TestCreateItem_GoesThroughRealOp(t *testing.T) {
	env := config.New()
	caller := env.Caller()

	// Resolver el id de la especialidad "medicina-general" desde el catálogo
	// real (los ids los genera el seed).
	specID := lookupTestSpecialty(t, caller, "medicina-general")
	if specID == "" {
		t.Fatal("expected a seeded medicina-general specialty")
	}

	// Upsert vía la op real (la misma que view.Ops.Save usa cuando el form
	// guarda un ítem nuevo). Sin Id: la op crea.
	var upsertErr error
	caller.Call(itemcatalog.OpUpsertItem, &itemcatalog.CatalogItem{
		TenantId:    config.TenantID,
		SpecialtyId: specID,
		Sku:         "md-nueva",
		Name:        "Consulta nueva",
		Type:        itemcatalog.ItemTypeService,
		IsActive:    true,
		Price:       25000,
		Currency:    "CLP",
	}, nil, func(err error) { upsertErr = err })
	if upsertErr != nil {
		t.Fatalf("OpUpsertItem: %v", upsertErr)
	}

	pres := itemcatalog.NewView(caller)
	if err := pres.Reload(); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	found := false
	for _, it := range pres.Items() {
		if it.Label == "Consulta nueva" {
			found = true
		}
	}
	if !found {
		t.Fatal("item created via the real op not found in the list")
	}
}

// lookupTestSpecialty lee OpListSpecialties y devuelve el id del slug dado.
func lookupTestSpecialty(t *testing.T, caller router.Caller, slug string) string {
	t.Helper()
	out := &itemcatalog.SpecialtyList{}
	var callErr error
	caller.Call(itemcatalog.OpListSpecialties,
		&itemcatalog.ListSpecialtiesArgs{TenantId: config.TenantID},
		out, func(err error) { callErr = err })
	if callErr != nil {
		t.Fatalf("OpListSpecialties: %v", callErr)
	}
	for i := 0; i < out.Len(); i++ {
		if out.At(i).(*itemcatalog.Specialty).Slug == slug {
			return out.At(i).(*itemcatalog.Specialty).Id
		}
	}
	return ""
}

// TestSpecialties_FedFromRealCatalog: config expone las especialidades desde el
// catálogo real (fuente única), que el selector de área de reservation usa.
func TestSpecialties_FedFromRealCatalog(t *testing.T) {
	env := config.New()
	specs := env.Specialties()
	if len(specs) == 0 {
		t.Fatal("expected specialties from the real catalog")
	}
	found := false
	for _, s := range specs {
		if s == "medicina-general" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected medicina-general from the catalog; got %v", specs)
	}
}
