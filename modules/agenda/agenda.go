// Package agenda es el módulo demo del editor de agenda: un profesional elige
// su plantilla semanal (7 días) y sus excepciones por fecha, persistido sobre
// el módulo REAL appointment_booking vía el loopback.Caller de config. No es
// un crudview: no hay "listado de registros"; es un editor de una sola forma,
// montado por un componente propio (scheduleView).
package agenda

import (
	tintime "webtyp.com/time"

	"webtyp.com/components/scheduleeditor"
	"webtyp.com/layout/platformd"
	"webtyp.com/svg"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
	. "webtyp.com/html"

	ab "github.com/veltylabs/appointment_booking"
	workschedule "github.com/veltylabs/work_schedule"

	"webtyp.com/app-demo/config"
)

const Icon = svg.Icon("mod-agenda")

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
	return &scheduleView{p: m.p, env: m.env, sel: NewString("")}
}

// scheduleView es el componente del editor: un picker de profesional en la
// cabecera y el scheduleeditor debajo. Al seleccionar otro profesional (o tras
// una escritura) rehace el subárbol del editor leyendo la agenda por las ops
// reales — el ScheduleEditor no es la fuente de verdad, el host lo remonta.
type scheduleView struct {
	Element // value embed
	p       *platformd.Platform
	env     *config.Env
	sel     *SignalString // staffId seleccionado
	// editor es el contenedor cuyo hijo (el ScheduleEditor) se rehace al
	// cambiar de profesional o tras una escritura. Un SignalNodes permite
	// swap de nodo sin reconstruir el árbol del chasis.
	editor *SignalNodes
	// schedule es el panel read-only del horario del profesional, alimentado
	// por work_schedule (lee sus tablas staff/workcalendar) — re-lleno en
	// cada cambio de staff.
	schedule *SignalNodes
}

func (s *scheduleView) Init(_ Ctx) {
	if s.sel.Get() == "" {
		staff := s.env.Staff()
		if len(staff) > 0 {
			s.sel.Set(staff[0].ID)
		}
	}
	if s.editor == nil {
		s.editor = NewNodes()
	}
	if s.schedule == nil {
		s.schedule = NewNodes()
	}
	s.reloadEditor()
	s.reloadSchedule()
}

// reloadSchedule re-llena el panel "Horario" del profesional elegido con las
// filas de work_schedule.NewView (lista read-only sobre sus tablas).
func (s *scheduleView) reloadSchedule() {
	wsStaffID := s.env.WorkScheduleStaffID(s.sel.Get())
	nodes := []*Element{}
	if wsStaffID != 0 {
		pres := workschedule.NewView(s.env.Caller(), wsStaffID)
		if err := pres.Reload(); err == nil {
			for _, it := range pres.Items() {
				nodes = append(nodes, Li().Text(it.Label+": "+it.Description))
			}
		}
	}
	if len(nodes) == 0 {
		nodes = append(nodes, Li().Text("Sin horario registrado"))
	}
	s.schedule.Set(nodes)
}

// selectedName devuelve el nombre del profesional seleccionado (linear scan
// sobre la lista corta de la demo).
func (s *scheduleView) selectedName() string {
	for _, so := range s.env.Staff() {
		if so.ID == s.sel.Get() {
			return so.Name
		}
	}
	return ""
}

// buildPicker arma el <select> de profesional.
func (s *scheduleView) buildPicker() *Element {
	sel := NewElement("select").Attr("name", "agenda-staff")
	for _, so := range s.env.Staff() {
		if so.ID == s.sel.Get() {
			sel.Child(SelectedOption(so.ID, so.Name))
		} else {
			sel.Child(Option(so.ID, so.Name))
		}
	}
	sel.On("change", func(ev Event) {
		s.sel.Set(ev.TargetValue())
		s.reloadEditor()
		s.reloadSchedule()
	})
	return sel
}

// reloadEditor reconstruye el ScheduleEditor del profesional seleccionado: lee
// la agenda vía ScheduleClient (ops reales) y lo monta como único hijo del
// contenedor editor.
func (s *scheduleView) reloadEditor() {
	s.editor.Set([]*Element{Div().Attr("class", "agenda__editor").Child(s.buildEditor())})
}

// buildEditor lee la agenda del profesional seleccionado (ops reales) y arma el
// ScheduleEditor con sus callbacks traducidos a escrituras. Separado de
// reloadEditor para que un test pueda inspeccionar el subárbol renderizado.
func (s *scheduleView) buildEditor() *scheduleeditor.ScheduleEditor {
	client := ab.NewScheduleClient(s.env.Caller(), s.env.TenantID(), s.sel.Get())

	var week []ab.WorkCalendarWeekly
	client.Weekly(func(rows []ab.WorkCalendarWeekly, err error) {
		week = rows
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
		Week:       fallbackWeek(week),
		Exceptions: toEditorExceptions(excs),
		Holidays:   s.env.Holidays2026(),
		OnWeeklyChange: func(r scheduleeditor.WeeklyRow) {
			client.SaveWeeklyRow(toWCWeekly(r), func(err error) {
				s.notifySave(err)
				if err == nil {
					s.reloadEditor()
				}
			})
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

// fallbackWeek asegura 7 filas Dom..Sáb: las que vienen de la op se mantienen,
// las que faltan van inactivas con horas 0 (un día sin fila = no agendado).
func fallbackWeek(rows []ab.WorkCalendarWeekly) []scheduleeditor.WeeklyRow {
	week := make([]scheduleeditor.WeeklyRow, 7)
	for _, r := range rows {
		dow := int(r.DayOfWeek)
		if dow >= 0 && dow < 7 {
			week[dow] = scheduleeditor.WeeklyRow{
				DayOfWeek:   dow,
				Active:      r.IsActive,
				WorkStart:   int(r.WorkStart),
				WorkFinish:  int(r.WorkFinish),
				BreakStart:  int(r.BreakStart),
				BreakFinish: int(r.BreakFinish),
			}
		}
	}
	return week
}

// notifySave muestra el toast de resultado de una escritura.
func (s *scheduleView) notifySave(err error) {
	if s.p == nil {
		return
	}
	if err != nil {
		s.p.Notify(Msg.Error, "No se pudo guardar la agenda", platformd.Auto())
		return
	}
	s.p.Notify(Msg.Success, "Agenda guardada", platformd.Auto())
}

// Render arma el editor completo.
func (s *scheduleView) Render() *Element {
	return Div().Attr("class", "agenda").
		Child(Div().Attr("class", "agenda__header").
			Child(Span().Text("Profesional")).Child(s.buildPicker())).
		Child(Div().Attr("class", "agenda__body").BindChildren(s.editor)).
		Child(Div().Attr("class", "agenda_schedule").
			Child(H2().Text("Horario")).
			Child(Ul().BindChildren(s.schedule)))
}

// ---------------------------------------------------------------------------
// Conversiones scheduleeditor ↔ appointment_booking (forma local, DRY en config
// si la Etapa F las repite).
// ---------------------------------------------------------------------------

func toWCWeekly(r scheduleeditor.WeeklyRow) ab.WorkCalendarWeekly {
	return ab.WorkCalendarWeekly{
		DayOfWeek:   int64(r.DayOfWeek),
		IsActive:    r.Active,
		WorkStart:   int64(r.WorkStart),
		WorkFinish:  int64(r.WorkFinish),
		BreakStart:  int64(r.BreakStart),
		BreakFinish: int64(r.BreakFinish),
	}
}

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
