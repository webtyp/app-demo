// Package agenda es el módulo demo del editor de agenda: un profesional elige
// su plantilla semanal (7 días) y sus excepciones por fecha, persistido sobre
// el módulo REAL appointment_booking vía el loopback.Caller de config. No es
// un crudview: no hay "listado de registros"; es un editor de una sola forma,
// montado por un componente propio (ScheduleView).
package agenda

import (
	tintime "webtyp.com/time"

	"webtyp.com/components/scheduleeditor"
	"webtyp.com/layout/platformd"
	"webtyp.com/svg"
	"webtyp.com/widget"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
	. "webtyp.com/html"

	ab "github.com/veltylabs/appointment_booking"

	"webtyp.com/app-demo/config"
	"webtyp.com/app-demo/modules/staffpick"
)

const Icon = svg.Icon("mod-agenda")

// NameAgenda es la identidad de widget de la vista. La clase del módulo se
// deriva de Name/Part, nunca de una string escrita a mano.
const NameAgenda = widget.Name("agenda")

const (
	PartHeader = widget.Part("header")
	PartStaff  = widget.Part("staff")
	PartTitle  = widget.Part("title")
	PartBody   = widget.Part("body")
)

var (
	clsRoot   = NameAgenda.Root()
	clsTitle  = NameAgenda.Class(PartTitle)
	clsHeader = NameAgenda.Class(PartHeader)
	clsStaff  = NameAgenda.Class(PartStaff)
	clsBody   = NameAgenda.Class(PartBody)
)

// Module es el módulo demo del editor de agenda.
type Module struct {
	p   *platformd.Platform
	env *config.Env
}

// New construye el módulo. Recibe la plataforma (para notificar) y el entorno
// demo compartido (el caller y el seed de appointment_booking real).
func New(p *platformd.Platform, env *config.Env) *Module { return &Module{p: p, env: env} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "agenda" }
func (m *Module) Label() string     { return "Agenda" }
func (m *Module) Icon() svg.Icon    { return Icon }

// View devuelve el editor de agenda montado sobre el módulo real.
func (m *Module) View() Component {
	return &ScheduleView{p: m.p, env: m.env, sel: NewString("")}
}

// ScheduleView es el componente del editor: un picker de profesional en la
// cabecera y el scheduleeditor debajo. Al seleccionar otro profesional (o tras
// una escritura) rehace el subárbol del editor leyendo la agenda por las ops
// reales — el ScheduleEditor no es la fuente de verdad, el host lo remonta.
type ScheduleView struct {
	Element // value embed
	p       *platformd.Platform
	env     *config.Env
	sel     *SignalString // staffId seleccionado
	// editor es el contenedor cuyo hijo (el ScheduleEditor) se rehace al
	// cambiar de profesional o tras una escritura. Un SignalNodes permite
	// swap de nodo sin reconstruir el árbol del chasis.
	editor *SignalNodes
}

func (s *ScheduleView) WidgetName() widget.Name { return NameAgenda }
func (s *ScheduleView) WidgetKind() widget.Kind { return widget.Form }

func (s *ScheduleView) Init(_ Ctx) {
	if s.sel.Get() == "" {
		staff := s.env.Staff()
		if len(staff) > 0 {
			s.sel.Set(staff[0].ID)
		}
	}
	if s.editor == nil {
		s.editor = NewNodes()
	}
	s.reloadEditor()
}

// selectedName devuelve el nombre del profesional seleccionado (linear scan
// sobre la lista corta de la demo).
func (s *ScheduleView) selectedName() string {
	for _, so := range s.env.Staff() {
		if so.ID == s.sel.Get() {
			return so.Name
		}
	}
	return ""
}

// reloadEditor reconstruye el ScheduleEditor del profesional seleccionado: lee
// la agenda vía ScheduleClient (ops reales) y lo monta como único hijo del
// contenedor editor.
func (s *ScheduleView) reloadEditor() {
	s.editor.Set([]*Element{Div().Child(s.buildEditor())})
}

// buildEditor lee la agenda del profesional seleccionado (ops reales) y arma el
// ScheduleEditor con sus callbacks traducidos a escrituras. Separado de
// reloadEditor para que un test pueda inspeccionar el subárbol renderizado.
func (s *ScheduleView) buildEditor() *scheduleeditor.ScheduleEditor {
	client := ab.NewScheduleClient(s.env.Caller(), s.env.TenantID(), s.sel.Get())

	var blocks []ab.WorkCalendarBlock
	client.Blocks(func(rows []ab.WorkCalendarBlock, err error) {
		blocks = rows
		if err != nil {
			s.notifySave(err)
		}
	})

	from := unixDay("2026-01-01")
	to := unixDay("2026-12-31")
	var excs []ab.WorkCalendarException
	client.Exceptions(from, to, func(rows []ab.WorkCalendarException, err error) {
		excs = rows
		if err != nil {
			s.notifySave(err)
		}
	})

	return &scheduleeditor.ScheduleEditor{
		Pattern:    blocksToPattern(blocks),
		Exceptions: toEditorExceptions(excs),
		Holidays:   s.env.Holidays2026(),
		OnPatternChange: func(rows []scheduleeditor.PatternRow) {
			var firstErr error
			pending := 7
			for dow := 0; dow <= 6; dow++ {
				var dayBlocks []ab.WorkCalendarBlock
				for _, r := range rows {
					for _, d := range r.Days {
						if d == dow {
							dayBlocks = append(dayBlocks, ab.WorkCalendarBlock{
								TenantId:  s.env.TenantID(),
								StaffId:   s.sel.Get(),
								DayOfWeek: int64(dow),
								StartMin:  int64(r.StartMin),
								EndMin:    int64(r.EndMin),
								IsActive:  true,
							})
						}
					}
				}
				client.SaveDayBlocks(dow, dayBlocks, func(err error) {
					if err != nil && firstErr == nil {
						firstErr = err
					}
					pending--
					if pending == 0 {
						s.notifySave(firstErr)
						if firstErr == nil {
							s.reloadEditor()
						}
					}
				})
			}
		},
		OnExceptionAdd: func(x scheduleeditor.Exception) {
			client.AddException(toWCException(x), func(err error) {
				s.notifySave(err)
				if err == nil {
					s.reloadEditor()
				}
			})
		},
		OnExceptionRemove: func(id string) {
			client.RemoveException(id, func(err error) {
				s.notifySave(err)
				if err == nil {
					s.reloadEditor()
				}
			})
		},
	}
}

func blocksToPattern(blocks []ab.WorkCalendarBlock) []scheduleeditor.PatternRow {
	type rangeKey struct {
		start, end int
	}
	var keys []rangeKey
	daysMap := make([][]int, 0)

	for _, b := range blocks {
		if !b.IsActive || b.DayOfWeek < 0 || b.DayOfWeek > 6 {
			continue
		}
		sm, em := int(b.StartMin), int(b.EndMin)
		dow := int(b.DayOfWeek)

		idx := -1
		for i, k := range keys {
			if k.start == sm && k.end == em {
				idx = i
				break
			}
		}
		if idx == -1 {
			keys = append(keys, rangeKey{start: sm, end: em})
			daysMap = append(daysMap, []int{dow})
		} else {
			daysMap[idx] = append(daysMap[idx], dow)
		}
	}

	rows := make([]scheduleeditor.PatternRow, len(keys))
	for i, k := range keys {
		rows[i] = scheduleeditor.PatternRow{
			StartMin: k.start,
			EndMin:   k.end,
			Days:     daysMap[i],
		}
	}
	return rows
}

// notifySave muestra el toast de resultado de una escritura.
func (s *ScheduleView) notifySave(err error) {
	if s.p == nil {
		return
	}
	if err != nil {
		s.p.Notify(Msg.Error, "No se pudo guardar la agenda", platformd.Auto())
		return
	}
	s.p.Notify(Msg.Success, "Agenda guardada", platformd.Auto())
}

// Render arma el editor completo con el chrome de la vista.
func (s *ScheduleView) Render() *Element {
	return Div().Set(clsRoot.AsAttr()).
		Child(H1().Set(clsTitle.AsAttr()).Text("Agenda")).
		Child(Div().Set(clsHeader.AsAttr()).
			Child(Span().Text("Profesional")).
			Child(staffpick.Select(s.env.Staff(), s.sel, func(string) { s.reloadEditor() }).Set(clsStaff.AsAttr()))).
		Child(Div().Set(clsBody.AsAttr()).BindChildren(s.editor))
}

// ---------------------------------------------------------------------------
// Conversiones scheduleeditor ↔ appointment_booking (forma local, DRY en config
// si la Etapa F las repite).
// ---------------------------------------------------------------------------

func toWCException(x scheduleeditor.Exception) ab.WorkCalendarException {
	return ab.WorkCalendarException{
		SpecificDate:  unixDay(x.Date),
		ExceptionType: x.Type,
		StartTime:     int64(x.StartMin),
		EndTime:       int64(x.EndMin),
		Notes:         x.Notes,
	}
}

func toEditorExceptions(excs []ab.WorkCalendarException) []scheduleeditor.Exception {
	out := make([]scheduleeditor.Exception, 0, len(excs))
	for _, x := range excs {
		out = append(out, scheduleeditor.Exception{
			ID:       x.Id,
			Date:     unixDate(x.SpecificDate),
			Type:     x.ExceptionType,
			StartMin: int(x.StartTime),
			EndMin:   int(x.EndTime),
			Notes:    x.Notes,
		})
	}
	return out
}

// unixDay convierte "YYYY-MM-DD" a medianoche UTC en segundos (la forma que
// work_calendar_exception.specific_date almacena).
func unixDay(dateStr string) int64 {
	nano, err := tintime.ParseDate(dateStr)
	if err != nil {
		return 0
	}
	return nano / 1000000000
}

// unixDate convierte segundos desde epoch a "YYYY-MM-DD" (FormatDate espera
// UnixNano, por eso se escala ×10^9).
func unixDate(seconds int64) string {
	return tintime.FormatDate(seconds * 1000000000)
}
