// Package workschedule es el módulo demo que monta la vista read-only
// "Horario (sistema legado)" sobre el módulo REAL github.com/veltylabs/work_schedule,
// que lee sus tablas staff/workcalendar (ownership externo). Se separó del
// módulo agenda para no tener dos fuentes de verdad en una misma pantalla.
package workschedule

import (
	"webtyp.com/layout/platformd"
	"webtyp.com/svg"
	"webtyp.com/widget"

	. "webtyp.com/dom"
	. "webtyp.com/html"

	ws "github.com/veltylabs/work_schedule"

	"webtyp.com/app-demo/config"
	"webtyp.com/app-demo/modules/staffpick"
)

const Icon = svg.Icon("mod-workschedule")

// NameWorkSchedule es la identidad de widget de la vista. La clase del módulo
// se deriva de Name/Part, nunca de una string escrita a mano — mismo patrón
// que modules/agenda.
const NameWorkSchedule = widget.Name("workschedule")

const (
	PartHeader = widget.Part("header")
	PartTitle  = widget.Part("title")
	PartList   = widget.Part("list")
)

var (
	clsRoot   = NameWorkSchedule.Root()
	clsTitle  = NameWorkSchedule.Class(PartTitle)
	clsHeader = NameWorkSchedule.Class(PartHeader)
	clsList   = NameWorkSchedule.Class(PartList)
)

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
	return &ScheduleList{p: m.p, env: m.env, sel: NewString("")}
}

// ScheduleList es el componente de la vista: un picker de profesional y la
// lista del horario legado. Al cambiar de profesional rehace el subárbol de la
// lista leyendo get_work_schedule por las ops reales — el módulo solo lee.
type ScheduleList struct {
	Element // value embed
	p       *platformd.Platform
	env     *config.Env
	sel     *SignalString // staffId seleccionado
	// items es el contenedor cuya lista se rehace al cambiar de profesional.
	// Un SignalNodes permite swap de nodo sin reconstruir el árbol del chasis.
	items *SignalNodes
}

func (s *ScheduleList) WidgetName() widget.Name { return NameWorkSchedule }
func (s *ScheduleList) WidgetKind() widget.Kind { return widget.Form }

func (s *ScheduleList) Init(_ Ctx) {
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
func (s *ScheduleList) reload() {
	wsStaffID := s.env.WorkScheduleStaffID(s.sel.Get())
	nodes := []*Element{}
	if wsStaffID != 0 {
		pres := ws.NewView(s.env.Caller(), wsStaffID)
		if err := pres.Reload(); err == nil {
			for _, it := range pres.Items() {
				nodes = append(nodes, Li().Text(it.Label+": "+it.Description))
			}
		}
	}
	if len(nodes) == 0 {
		nodes = append(nodes, Li().Text("Sin horario registrado"))
	}
	s.items.Set(nodes)
}

// Render arma la vista: encabezado, picker y la lista del horario legado.
func (s *ScheduleList) Render() *Element {
	return Div().Set(clsRoot.AsAttr()).
		Child(H2().Set(clsTitle.AsAttr()).Text("Horario (sistema legado)")).
		Child(Div().Set(clsHeader.AsAttr()).
			Child(staffpick.Select(s.env.Staff(), s.sel, func(string) { s.reload() }))).
		Child(Ul().Set(clsList.AsAttr()).BindChildren(s.items))
}
