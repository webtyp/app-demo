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

type catalogMod struct {
	p   *platformd.Platform
	env *config.Env
}

func NewCatalog(p *platformd.Platform, env *config.Env) *catalogMod {
	return &catalogMod{p: p, env: env}
}

var _ platformd.UIModule = (*catalogMod)(nil)

func (m *catalogMod) ModelName() string { return "catalog" }
func (m *catalogMod) Label() string     { return "Catálogo" }
func (m *catalogMod) Icon() svg.Icon    { return IconCatalog }

// View monta el catálogo real con crudview.New(Config{}) SIN config custom —
// el mismo layout que devices/medicalhistory, la prueba de que reusable module
// + crudview = cero bifurcación por módulo.
func (m *catalogMod) View() dom.Component {
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

type specialtyMod struct {
	p   *platformd.Platform
	env *config.Env
}

func NewSpecialties(p *platformd.Platform, env *config.Env) *specialtyMod {
	return &specialtyMod{p: p, env: env}
}

var _ platformd.UIModule = (*specialtyMod)(nil)

func (m *specialtyMod) ModelName() string { return "specialties" }
func (m *specialtyMod) Label() string     { return "Especialidades" }
func (m *specialtyMod) Icon() svg.Icon    { return IconSpecial }

func (m *specialtyMod) View() dom.Component {
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
