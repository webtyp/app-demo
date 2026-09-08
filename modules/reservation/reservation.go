// Package reservation es el módulo demo de reserva de hora v2: sobre el módulo
// REAL appointment_booking, con un selector de área (especialidad) + médico en
// el slot Context del crudview.
//
// Restricción clave: NO se usa appointmentbooking.NewView. Su Presenter lista
// appointmentbooking.Reservation, cuyos campos son kinds base (model.Text()/
// model.Int()) sin widgets de form — y crudview hace form.New(record) que
// falla "loud" con un record sin widgets. Por eso este módulo mantiene su
// propio record local (reservation.Reservation, con input.Text() en los campos
// que el usuario edita) y un lister/saver a medida traduce a/desde las ops
// reales. El crudview no cambia de forma.
//
// Brecha conocida — sin módulo Directory: appointment_booking referencia al
// paciente por client_id (validado por DirectoryReader). La demo no tiene
// Directory, así que el formulario mantiene los campos de paciente como texto
// libre y al crear se usa el run como client_id sintético; config usa un
// stubDirectory que devuelve siempre true. Un despliegue real resolvería el
// cliente contra un Directory de verdad (como mjosefa-cms documenta su brecha
// de picker de paciente).
package reservation

import (
	"webtyp.com/components/calendarslider"
	"webtyp.com/components/targethour"
	"webtyp.com/events"
	"webtyp.com/layout/crudview"
	"webtyp.com/layout/platformd"
	"webtyp.com/model"
	"webtyp.com/svg"
	tintime "webtyp.com/time"
	"webtyp.com/unixid"
	"webtyp.com/view"

	. "webtyp.com/dom"
	. "webtyp.com/fmt"
	. "webtyp.com/html"

	ab "github.com/veltylabs/appointment_booking"

	"webtyp.com/app-demo/config"
)

const Icon = svg.Icon("mod-reservation")

type Module struct {
	p   *platformd.Platform
	env *config.Env
}

func New(p *platformd.Platform, env *config.Env) *Module { return &Module{p: p, env: env} }

var _ platformd.UIModule = (*Module)(nil)

func (m *Module) ModelName() string { return "reservation" }
func (m *Module) Label() string     { return "Reserva Hora" }
func (m *Module) Icon() svg.Icon    { return Icon }

// View monta el crudview sobre el módulo real, envuelto en un componente propio
// (reservationView) cuyo Init suscribe a los eventos de appointment_booking: al
// cambiar la agenda del médico en foco (schedule.changed) o al crearse/
// cancelarse una reserva, la lista y los huecos libres se refrescan sin recargar
// la página — la sincronía por eventos del DEMO_AGENDA (§4.4).
func (m *Module) View() Component {
	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err)
	}

	return &reservationView{
		p:   m.p,
		env: m.env,
		ids: ids,
	}
}

// reservationView envuelve el crudview de reservas. Init corre UNA vez (antes
// del primer render): construye el crudview, suscribe al broker y deja el bridge
// que los eventos usan para refrescar lista + huecos.
type reservationView struct {
	Element // value embed
	p       *platformd.Platform
	env     *config.Env
	ids     model.IDGenerator

	cv         *crudview.CrudView
	target     *targethour.TargetHour // el list concreto, guardado para FreeSlots
	ctx        *reservationContext
	store      *reservationStore
	daySig     *SignalString // día activo, lo setea el calendario via OnFilterChange
	subscribed bool
	// refreshFn, si no es nil, reemplaza el refresh() por defecto — el único hook
	// que los tests necesitan para contar recargas sin montar el DOM.
	refreshFn func()
}

// currentStaff devuelve el médico scopeado actualmente, o "".
func (v *reservationView) currentStaff() string {
	if v.ctx == nil || v.ctx.staff == nil {
		return ""
	}
	return v.ctx.staff.Get()
}

// currentDay devuelve el día activo ("YYYY-MM-DD"), o "".
func (v *reservationView) currentDay() string {
	if v.daySig == nil {
		return ""
	}
	return v.daySig.Get()
}

// refresh recarga la lista y re-llena los huecos libres del día activo. Es el
// camino único que usan los eventos (schedule.changed / reserva) y el cambio de
// médico. FreeSlots se setea ANTES de Reload para que filter()→list.SetItems
// ya los dibuje dentro de la misma pasada.
func (v *reservationView) refresh() {
	if v.refreshFn != nil {
		v.refreshFn()
		return
	}
	if v.target != nil && v.store != nil {
		v.target.FreeSlots = v.store.freeSlotsForDay(v.currentDay())
	}
	if v.cv != nil {
		_ = v.cv.Reload()
	}
}

func (v *reservationView) Init(_ Ctx) {
	if v.subscribed {
		return
	}
	v.subscribed = true

	cal := &calendarslider.CalendarSlider{}
	v.daySig = NewString("")
	cal.OnFilterChange(func(term string) {
		v.daySig.Set(term)
		v.refresh()
	})

	v.ctx = &reservationContext{
		env:          v.env,
		staffOptions: v.env.Staff(),
	}
	v.store = &reservationStore{env: v.env, staff: v.ctx.staff}

	cv, err := crudview.New(crudview.Config{
		ParentID:  "reservation",
		Presenter: byDay{view.New(v.store, &Reservation{}, view.WithTitle("Reserva Hora"))},
		IDs:       v.ids,
		Filter:    cal,
		Context:   v.ctx,
		List: func(selected *SignalString, onSelect func(view.Item)) crudview.ListView {
			th := &targethour.TargetHour{
				Selected: selected,
				OnSelect: onSelect,
				StatusOf: func(it view.Item) targethour.Status {
					switch it.Description {
					case "Confirmada":
						return targethour.StatusConfirmed
					case "Atendida":
						return targethour.StatusAttended
					}
					return targethour.StatusPending
				},
				// Hueco libre clicado: anota la hora elegida (el día ya lo trae
				// el calendario). Pre-llenar el form queda a un engagement del
				// form; aquí se informa al usuario de qué hora quedó marcada.
				OnPickFree: func(hhmm string) {
					v.p.Notify(Msg.Info, "Hora libre "+hhmm, platformd.Auto())
				},
			}
			v.target = th
			return th
		},
	})
	if err != nil {
		panic(err)
	}
	v.cv = cv

	v.ctx.onStaffChange = func(staffId string) {
		v.store.staffId = staffId
		v.refresh()
	}

	cv.OnNew = func() { v.p.Notify(Msg.Info, "Nueva reserva", platformd.Auto()) }
	cv.OnSaved = func(err error) {
		if err == nil {
			v.p.Notify(Msg.Success, "Reserva creada", platformd.Auto())
			v.refresh()
			return
		}
		v.p.Notify(Msg.Error, err.Error(), platformd.Auto())
	}

	// Suscripción in-proc al broker de la demo (en producción mjosefa-cms:
	// webtyp/sse). El broker in-proc entrega el payload concreto sin codificar.
	broker := v.env.Broker()
	broker.Subscribe(ab.EventScheduleChanged, v.onScheduleChanged)
	for _, topic := range []string{
		ab.EventReservationCreated,
		ab.EventReservationCancelled,
		ab.EventReservationCompleted,
		ab.EventReservationNoShow,
	} {
		broker.Subscribe(topic, func(events.Event) { v.refresh() })
	}
}

// onScheduleChanged maneja schedule.changed: solo refresca si el evento es del
// médico en foco. Separado del Subscribe para poder probarlo unitariamente.
func (v *reservationView) onScheduleChanged(e events.Event) {
	pl, ok := e.Payload.(*ab.ScheduleChangedPayload)
	if !ok {
		return
	}
	if pl.StaffId != v.currentStaff() {
		return // solo si es el médico en foco
	}
	v.refresh()
}

func (v *reservationView) Render() *Element {
	if v.cv == nil {
		return Div()
	}
	return Div().Child(v.cv)
}

// reservationContext es el slot Context del crudview: selects de área
// (especialidad) y médico. Elegir área filtra las opciones de médico; elegir
// médico re-acota los datos. No es widget.Filterable (ver Etapa E): cambiar de
// médico es scope, no un término de búsqueda.
type reservationContext struct {
	Element       // value embed
	area          *SignalString
	staff         *SignalString
	env           *config.Env
	staffOptions  []config.StaffOption
	onStaffChange func(staffId string)
	staffNode     *SignalNodes // <option> del select de médico, recalculado por área
}

func (c *reservationContext) Init(_ Ctx) {
	if c.area == nil {
		c.area = NewString("")
	}
	if c.staff == nil {
		c.staff = NewString("")
	}
	if c.staffNode == nil {
		c.staffNode = NewNodes()
	}
	c.rebuildStaffOptions()
}

// staffForArea devuelve las opciones de médico acotadas por el área elegida
// ("" = todas). Scan lineal sobre la lista corta de la demo.
func (c *reservationContext) staffForArea() []config.StaffOption {
	out := []config.StaffOption{}
	for _, so := range c.staffOptions {
		if c.area.Get() == "" || so.SpecialtySlug == c.area.Get() {
			out = append(out, so)
		}
	}
	return out
}

// rebuildStaffOptions re-renderiza los <option> del select de médico acotados
// por el área — el segundo select sí necesita re-render de sus opciones, así
// que se ata a un SignalNodes.
func (c *reservationContext) rebuildStaffOptions() {
	opts := make([]*Element, 0, 4)
	for _, so := range c.staffForArea() {
		if so.ID == c.staff.Get() {
			opts = append(opts, SelectedOption(so.ID, so.Name))
		} else {
			opts = append(opts, Option(so.ID, so.Name))
		}
	}
	c.staffNode.Set(opts)
}

func (c *reservationContext) Render() *Element {
	if c.staffNode == nil {
		c.Init(nil)
	}

	area := NewElement("select").Attr("name", "reservation-area")
	area.Child(Option("", "Todas"))
	for _, sp := range c.env.Specialties() {
		if sp == c.area.Get() {
			area.Child(SelectedOption(sp, specialtyLabel(sp)))
		} else {
			area.Child(Option(sp, specialtyLabel(sp)))
		}
	}
	area.On("change", func(ev Event) {
		c.area.Set(ev.TargetValue())
		c.staff.Set("") // el médico anterior puede no pertenecer al área nueva
		c.rebuildStaffOptions()
	})

	staff := NewElement("select").Attr("name", "reservation-staff").
		BindChildren(c.staffNode)
	staff.On("change", func(ev Event) {
		c.staff.Set(ev.TargetValue())
		if c.onStaffChange != nil {
			c.onStaffChange(c.staff.Get())
		}
	})

	return Div().Attr("class", "reservation-context").
		Child(Span().Text("Área")).Child(area).
		Child(Span().Text("Médico")).Child(staff)
}

// specialtyLabel humaniza el slug canónico de item_catalog para el select.
func specialtyLabel(slug string) string {
	switch slug {
	case "medicina-general":
		return "Medicina General"
	case "dental":
		return "Dental"
	case "traumatologia":
		return "Traumatología"
	case "podologia":
		return "Podología"
	case "laboratorio":
		return "Laboratorio"
	case "ecografia":
		return "Ecografía"
	case "ginecologia-y-obstetricia":
		return "Ginecología y Obstetricia"
	case "cardiologia":
		return "Cardiología"
	case "gastroenterologia":
		return "Gastroenterología"
	case "oftalmologia":
		return "Oftalmología"
	case "neurologia":
		return "Neurología"
	case "psicologia":
		return "Psicología"
	case "dermatologia":
		return "Dermatología"
	case "radiologia":
		return "Radiología"
	default:
		return slug
	}
}

// reservationStore es el lister/saver a medida: mantiene el record local del
// formulario, pero lista y crea contra las ops reales de appointment_booking
// vía config.Caller().
type reservationStore struct {
	env *config.Env
	// staff apunta al signal del Context: el scope lo gobierna el médico
	// elegido, y Reload() (con el filtro día del calendario) lee s.staff.
	staff *SignalString
	// staffId es una copia plana que List/Save leen (evita re-lookup).
	staffId string
}

func (s *reservationStore) List() ([]model.Model, error) {
	if s.staffId == "" {
		// Sin médico scopeado no hay qué listar (mismo modelo que
		// medicalhistory sin paciente).
		return nil, nil
	}
	from := unixDay("2026-01-01")
	to := unixDay("2026-12-31")
	out := &ab.ReservationList{}
	var callErr error
	s.env.Caller().Call(
		ab.OpListReservationsByStaff,
		&ab.ListReservationsByStaffArgs{TenantId: s.env.TenantID(), StaffId: s.staffId, From: from, To: to},
		out,
		func(err error) { callErr = err },
	)
	if callErr != nil {
		return nil, callErr
	}
	rows := make([]model.Model, 0, out.Len())
	for i := 0; i < out.Len(); i++ {
		rows = append(rows, toLocal(out.At(i).(*ab.Reservation)))
	}
	return rows, nil
}

func (s *reservationStore) Save(recs ...model.Model) error {
	if len(recs) == 0 {
		return Errf("reservationStore: save: empty records")
	}
	if s.staffId == "" {
		return Errf("elige un médico para reservar")
	}
	escID := s.env.ESCForStaff(s.staffId)
	if escID == "" {
		return Errf("no hay servicio configurado para este profesional en la demo")
	}
	for _, m := range recs {
		r := m.(*Reservation)
		slotUTC := slotStartUTC(r.Day, r.Hour)
		if slotUTC == 0 {
			return Errf("faltan día y hora para la reserva")
		}
		var callErr error
		s.env.Caller().Call(
			ab.OpCreateReservation,
			&ab.CreateReservationArgs{
				TenantId:                s.env.TenantID(),
				ClientId:                r.PatientRun,
				CreatorUserId:           "demo",
				EmployeeServiceConfigId: escID,
				SlotStartUtc:            slotUTC,
				Notes:                   r.PatientName,
			},
			nil,
			func(err error) { callErr = err },
		)
		if callErr != nil {
			return callErr
		}
	}
	return nil
}

// Deletes/Updates: las reservas solo mutan por transiciones FSM-gated en
// appointment_booking; la demo no expone borrado directo ni bulk-edit. El
// Presenter no declara estas capacidades, así que crudview no pinta los botones.
func (s *reservationStore) Delete(ids ...string) error {
	return Errf("reservationStore: las reservas no se eliminan en la demo")
}

func (s *reservationStore) Update(ids []string, rec model.Model, fields []string) error {
	return Errf("reservationStore: las reservas no se editan directo en la demo")
}

// toLocal convierte la reserva del módulo a la forma local del formulario.
func toLocal(r *ab.Reservation) *Reservation {
	status := ""
	switch r.Status {
	case ab.StatusConfirmed:
		status = "confirmed"
	case ab.StatusCompleted, ab.StatusNoShow:
		status = "attended"
	}
	return &Reservation{
		Id:          r.Id,
		PatientRun:  r.ClientId,
		PatientName: r.Notes,
		Day:         r.LocalStringDate,
		Hour:        r.LocalStringTime,
		Detail:      r.Notes,
		Status:      status,
	}
}

// slotStartUTC arma el slot_start_utc (segundos epoch) desde "YYYY-MM-DD" +
// "HH:MM". 0 si falta algo o no se puede parsear.
// slotStartUTC arma el slot_start_utc (segundos epoch) desde "YYYY-MM-DD" +
// "HH:MM". "HH:MM" es la hora LOCAL que el usuario elige en el form — la misma
// conversión (LocalIntToUnixUTC con la zona del work_calendar_config) que hace
// list_availability al alinear slots y reservas. ParseDateTime sin zona
// produciría un UTC bruto, desalineado con los huecos (bug corregido).
func slotStartUTC(day, hour string) int64 {
	if day == "" || hour == "" {
		return 0
	}
	minOfDay, err := tintime.ParseTime(hour)
	if err != nil {
		return 0
	}
	return ab.LocalIntToUnixUTC(unixDay(day), int(minOfDay), timezoneDemo)
}

// timezoneDemo es la zona de los work_calendar_config sembrados en config —
// misma constante que el seed de reservas usa.
const timezoneDemo = "America/Santiago"

func unixDay(dateStr string) int64 {
	nano, err := tintime.ParseDate(dateStr)
	if err != nil {
		return 0
	}
	return nano / 1000000000
}

// freeSlotsForDay consulta list_availability para el médico scopeado en el día
// dado y devuelve los huecos como "HH:MM" en hora local. El arg config_id es el
// id del employee_service_config (no del work_calendar_config) — service.go:
// ListAvailability hace GetEmployeeServiceConfig(configId). Sin médico o sin
// servicio, devuelve nil.
func (s *reservationStore) freeSlotsForDay(day string) []string {
	if s.staffId == "" || day == "" {
		return nil
	}
	escID := s.env.ESCForStaff(s.staffId)
	if escID == "" {
		return nil
	}
	out := &ab.TimeSlotList{}
	var callErr error
	s.env.Caller().Call(
		ab.OpListAvailability,
		&ab.ListAvailabilityArgs{
			TenantId: s.env.TenantID(),
			StaffId:  s.staffId,
			ConfigId: escID,
			From:     unixDay(day),
			To:       unixDay(day),
		},
		out,
		func(err error) { callErr = err },
	)
	if callErr != nil {
		return nil
	}
	slots := make([]string, 0, out.Len())
	for i := 0; i < out.Len(); i++ {
		slot := out.At(i).(*ab.TimeSlot)
		// FormatTime espera UnixNano; la hora local del inicio del hueco.
		hhmm := tintime.FormatTime(slot.StartUtc * 1000000000)
		if len(hhmm) >= 5 {
			hhmm = hhmm[:5]
		}
		slots = append(slots, hhmm)
	}
	return slots
}

// byDay adapta el Presenter: el filtro (term = "YYYY-MM-DD" del calendario)
// filtra la lista ya cargada por fecha.
type byDay struct {
	view.Presenter
}

func (p byDay) Filter(term string) []view.Item {
	if term == "" {
		return nil
	}
	var items []view.Item
	for _, it := range p.Presenter.Items() {
		if it.ID != "" && it.LeadMain != "" {
			items = append(items, it)
		}
	}
	return items
}

// Save/Update/Delete delegan las capacidades al Presenter subyacente: view
// re-expone las del lister, y crudview las lee desde byDay (el tipo que monta).
func (p byDay) Save(recs ...model.Model) error {
	if s, ok := p.Presenter.(view.Saver); ok {
		return s.Save(recs...)
	}
	return Errf("byDay: underlying presenter cannot save")
}

func (p byDay) Update(ids []string, rec model.Model, fields []string) error {
	if u, ok := p.Presenter.(view.Updater); ok {
		return u.Update(ids, rec, fields)
	}
	return Errf("byDay: underlying presenter cannot update")
}

func (p byDay) Delete(ids ...string) error {
	if d, ok := p.Presenter.(view.Deleter); ok {
		return d.Delete(ids...)
	}
	return Errf("byDay: underlying presenter cannot delete")
}
