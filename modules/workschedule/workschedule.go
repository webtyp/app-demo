// Package workschedule es el módulo demo que monta la vista read-only
// "Horario (sistema legado)" sobre el módulo REAL github.com/veltylabs/work_schedule,
// que lee sus tablas staff/workcalendar (ownership externo). Se separó del
// módulo agenda para no tener dos fuentes de verdad en una misma pantalla.
package workschedule

import (
	"webtyp.com/layout/platformd"
	"webtyp.com/svg"

	. "webtyp.com/dom"
	. "webtyp.com/html"

	ws "github.com/veltylabs/work_schedule"

	"webtyp.com/app-demo/config"
	"webtyp.com/app-demo/modules/staffpick"
)

const Icon = svg.Icon("mod-workschedule")

// Module es el módulo demo del horario legado.
type Module struct {
	p   *platformd.Platform
	env *config.Env
}

// New construye el módulo.
func New(p *platformd.Platform, env *config.Env) *Module { return &Module{p: p, env: env} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "workschedule" }
func (m *Module) Label() string     { return "Horario legado" }
func (m *Module) Icon() svg.Icon    { return Icon }

// View devuelve la lista read-only del horario del profesional elegido.
func (m *Module) View() Component {
	return &scheduleList{p: m.p, env: m.env, sel: NewString("")}
}

// scheduleList es el componente de la vista: un picker de profesional y la
// lista del horario legado. Al cambiar de profesional rehace el subárbol de la
// lista leyendo get_work_schedule por las ops reales — el módulo solo lee.
type scheduleList struct {
	Element // value embed
	p       *platformd.Platform
	env     *config.Env
	sel     *SignalString // staffId seleccionado
	// items es el contenedor cuya lista se rehace al cambiar de profesional.
	// Un SignalNodes permite swap de nodo sin reconstruir el árbol del chasis.
	items *SignalNodes
}

func (s *scheduleList) Init(_ Ctx) {
	if s.sel.Get() == "" {
		staff := s.env.Staff()
		if len(staff) > 0 {
			s.sel.Set(staff[0].ID)
		}
	}
	if s.items == nil {
		s.items = NewNodes()
	}
	s.reload()
}

// reload re-llena la lista del profesional elegido con las filas de
// workschedule.NewView (lista read-only sobre sus tablas).
func (s *scheduleList) reload() {
	wsStaffID := s.env.WorkScheduleStaffID(s.sel.Get())
	nodes := []*Element{}
	if wsStaffID != 0 {
		pres := ws.NewView(s.env.Caller(), wsStaffID)
		if err := pres.Reload(); err == nil {
			for _, it := range pres.Items() {
				nodes = append(nodes, Li().Text(it.Label + ": " + it.Description))
			}
		}
	}
	if len(nodes) == 0 {
		nodes = append(nodes, Li().Text("Sin horario registrado"))
	}
	s.items.Set(nodes)
}

// Render arma la vista: encabezado, picker y la lista del horario legado.
func (s *scheduleList) Render() *Element {
	return Div().
		Child(H2().Text("Horario (sistema legado)")).
		Child(staffpick.Select(s.env.Staff(), s.sel, func(string) { s.reload() })).
		Child(Ul().BindChildren(s.items))
}
