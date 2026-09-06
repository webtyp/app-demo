// Package about es el módulo demo de una pantalla de información estática.
package about

import (
	"webtyp.com/layout/platformd"
	"webtyp.com/layout/rightpanel"
	"webtyp.com/svg"

	. "webtyp.com/dom"
	. "webtyp.com/html"
)

const Icon = svg.Icon("mod-about")

type Module struct{}

func New() *Module { return &Module{} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "about" }
func (m *Module) Label() string     { return "Acerca de" }
func (m *Module) Icon() svg.Icon    { return Icon }

func (m *Module) View() Component {
	return &rightpanel.RightPanel{
		Title:   m.Label(),
		Article: Div().Text("Contenido de " + m.Label()),
	}
}
