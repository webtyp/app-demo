// Package itemcatalogdemo es el módulo demo que monta el módulo REAL
// github.com/veltylabs/item_catalog con el layout crudview SIN configuración
// custom (sin Filter/List/Context). Es la prueba de reutilización del §1 del
// usuario: "no un layout por módulo". Dos UIModule separados — "Catálogo" e
// "Ítems" — como hace mjosefa-cms, sin inventar sub-navegación en platformd.
package itemcatalogdemo

import (
	"webtyp.com/dom"
	"webtyp.com/layout/crudview"
	"webtyp.com/layout/platformd"
	"webtyp.com/svg"

	itemcatalog "github.com/veltylabs/item_catalog"

	"webtyp.com/app-demo/config"
)

const (
	IconCatalog = svg.Icon("mod-catalog")
	IconSpecial = svg.Icon("mod-specialties")
)

type CatalogModule struct {
	p   *platformd.Platform
	env *config.Env
}

func NewCatalog(p *platformd.Platform, env *config.Env) *CatalogModule {
	return &CatalogModule{p: p, env: env}
}

var _ platformd.UIModule = (*CatalogModule)(nil)

func (m *CatalogModule) ModelName() string { return "catalog" }
func (m *CatalogModule) Label() string     { return "Catálogo" }
func (m *CatalogModule) Icon() svg.Icon    { return IconCatalog }

// View monta el catálogo real con crudview.New(Config{}) SIN config custom —
// el mismo layout que devices/medicalhistory, la prueba de que reusable module
// + crudview = cero bifurcación por módulo.
func (m *CatalogModule) View() dom.Component {
	v, err := crudview.New(crudview.Config{
		ParentID:  "catalog_item",
		Presenter: itemcatalog.NewView(m.env.Caller()),
		IDs:       m.env.IDs(),
	})
	if err != nil {
		panic(err)
	}
	return v
}

type SpecialtyModule struct {
	p   *platformd.Platform
	env *config.Env
}

func NewSpecialties(p *platformd.Platform, env *config.Env) *SpecialtyModule {
	return &SpecialtyModule{p: p, env: env}
}

var _ platformd.UIModule = (*SpecialtyModule)(nil)

func (m *SpecialtyModule) ModelName() string { return "specialties" }
func (m *SpecialtyModule) Label() string     { return "Especialidades" }
func (m *SpecialtyModule) Icon() svg.Icon    { return IconSpecial }

func (m *SpecialtyModule) View() dom.Component {
	v, err := crudview.New(crudview.Config{
		ParentID:  "specialty",
		Presenter: itemcatalog.NewSpecialtyView(m.env.Caller()),
		IDs:       m.env.IDs(),
	})
	if err != nil {
		panic(err)
	}
	return v
}
