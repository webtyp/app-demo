// Package config is the demo's composition root. A single Env (env.go) builds
// the shared runtime the real-module wrappers consume: an in-memory orm.DB
// seeded, the veltylabs domain modules mounted on a router/loopback caller,
// and an events broker. css.go carries the visual theme and lang.go the
// Spanish dictionary — this root is one place, not three.
package config

import (
	"webtyp.com/events"
	"webtyp.com/events/mock"
	"webtyp.com/model"
	"webtyp.com/orm"
	"webtyp.com/router"
	"webtyp.com/router/loopback"
	"webtyp.com/storage/mem"
	tinytime "webtyp.com/time"
	"webtyp.com/unixid"

	ab "github.com/veltylabs/appointment_booking"
	"github.com/veltylabs/item_catalog"
)

// TenantID es el id de tenant único de la demo.
const TenantID = "demo"

// StaffOption es un miembro del staff para los pickers de la demo. La lista no
// vive en una tabla (appointment_booking valida staff por el StaffReader
// inyectado, no por una tabla propia); es configuración de la demo, como lo
// sería un módulo staff en producción.
type StaffOption struct {
	ID            string
	Name          string
	SpecialtySlug string
}

// slugTraumatologia / slugMedicinaGeneral / slugEcografia son los slugs
// canónicos que item_catalog ya define — se derivan de su lista para no
// inventar especialidades (ver DEMO_AGENDA_MASTER_PLAN D6).
func slugTraumatologia() string   { return lookupSlug("traumatologia") }
func slugMedicinaGeneral() string { return lookupSlug("medicina-general") }
func slugEcografia() string       { return lookupSlug("ecografia") }

// lookupSlug devuelve el slug canónico de item_catalog, o "" si no existe.
func lookupSlug(slug string) string {
	for _, cs := range itemcatalog.CanonicalSpecialties {
		if cs.Slug == slug {
			return cs.Slug
		}
	}
	return ""
}

// Specialties son las especialidades (áreas) canónicas del catálogo — la forma
// canónica de la demo para el selector de área. La Etapa H sigue leyéndolas de
// item_catalog (fuente única); aquí se derivan de CanonicalSpecialties.
func (e *Env) Specialties() []string {
	out := make([]string, 0, len(itemcatalog.CanonicalSpecialties))
	for _, cs := range itemcatalog.CanonicalSpecialties {
		out = append(out, cs.Slug)
	}
	return out
}

// ESCForStaff devuelve el employee_service_config.id primario del médico, o ""
// si no tiene. La demo ofrece UN servicio por médico (no hay picker de servicio
// — el Context del crudview es área+médico). Este id es el que consume
// create_reservation (employee_service_config_id) y list_availability
// (config_id — NO el del work_calendar_config, ver service.go).
func (e *Env) ESCForStaff(staffId string) string {
	switch staffId {
	case "staff-natasha":
		return "escNatashaConsulta"
	case "staff-tony":
		return "escTonyCirugia"
	case "staff-thor":
		return "escThorEcografia"
	}
	return ""
}

// Env es la composición de la demo: un orm.DB en memoria sembrado, los módulos
// de dominio reales montados, un router.Caller in-proc y un events broker.
// Se construye UNA vez en web/client.go y se comparte entre módulos.
type Env struct {
	caller   router.Caller
	broker   events.Broker
	ids      model.IDGenerator
	ab       *ab.Module
	db       *orm.DB
	staff    []StaffOption
	holidays []string
}

// New construye el entorno: DB en memoria, módulos reales en un loopback.Caller
// y seed realista. Compila a WASM: el Caller vive client-side.
func New() *Env {
	db := orm.New(mem.New())

	ids, err := unixid.NewUnixID()
	if err != nil {
		panic(err)
	}
	broker := &mock.Broker{}

	abMod, err := ab.New(db, ab.Deps{
		// La demo confía en su propio seed (lista de staff en memoria + servicio
		// sembrado) — los readers solo validan existencia al cerrar una reserva,
		// y aquí todo lo que la UI ofrece existe. En un despliegue real estos
		// serían los módulos staff / service_catalog / directory de verdad.
		Staff:     stubStaff{},
		Catalog:   stubCatalog{},
		Directory: stubDirectory{},
		IDs:       ids,
		Publisher: broker,
	})
	if err != nil {
		panic(err)
	}

	env := &Env{
		caller:   loopback.New(abMod),
		broker:   broker,
		ids:      ids,
		ab:       abMod,
		db:       db,
		holidays: holidaysCL2026(),
		staff: []StaffOption{
			{ID: "staff-tony", Name: "Dr. Tony Stark", SpecialtySlug: slugTraumatologia()},
			{ID: "staff-natasha", Name: "Dra. Natasha Romanoff", SpecialtySlug: slugMedicinaGeneral()},
			{ID: "staff-thor", Name: "Dr. Thor Odinson", SpecialtySlug: slugEcografia()},
		},
	}

	env.seed()
	return env
}

// Caller es el router.Caller in-proc que despacha contra los módulos reales
// montados.
func (e *Env) Caller() router.Caller  { return e.caller }
func (e *Env) Broker() events.Broker  { return e.broker }
func (e *Env) IDs() model.IDGenerator { return e.ids }
func (e *Env) TenantID() string       { return TenantID }
func (e *Env) Staff() []StaffOption   { return e.staff }

// Holidays2026 son los feriados de Chile 2026 ("YYYY-MM-DD"), solo lectura. En
// producción vienen de un servicio de feriados; aquí es una constante.
func (e *Env) Holidays2026() []string { return e.holidays }

// seed puebla la DB en memoria vía las ops REALES de appointment_booking cuando
// existe una op, y con db.Create directo cuando el módulo no la expone (las
// excepciones están documentadas en cada método).
func (e *Env) seed() {
	e.upsertCalendarConfigs()
	e.upsertWeekly()
	e.addExceptions()
	e.seedEmployeeServiceConfig()
	e.seedReservations()
}

func (e *Env) upsertCalendarConfigs() {
	for _, s := range e.staff {
		var doneErr error
		e.caller.Call(ab.OpUpsertCalendarConfig, &ab.UpsertCalendarConfigArgs{
			TenantId: TenantID, StaffId: s.ID, Timezone: "America/Santiago", IsActive: true,
		}, nil, func(err error) { doneErr = err })
		if doneErr != nil {
			panic(doneErr)
		}
	}
}

func (e *Env) upsertWeekly() {
	call := func(op string, args model.Encodable) {
		var doneErr error
		e.caller.Call(op, args, nil, func(err error) { doneErr = err })
		if doneErr != nil {
			panic(doneErr)
		}
	}
	// Natasha: Lun–Vie 09:00–18:00 con colación 13:00–14:00.
	for dow := int64(1); dow <= 5; dow++ {
		call(ab.OpUpsertWeeklyCalendar, &ab.UpsertWeeklyCalendarArgs{
			TenantId: TenantID, StaffId: "staff-natasha", DayOfWeek: dow,
			WorkStart: 540, WorkFinish: 1080, BreakStart: 780, BreakFinish: 840, IsActive: true,
		})
	}
	// Tony: Lun/Mié/Vie 08:00–14:00 sin colación.
	for _, dow := range []int64{1, 3, 5} {
		call(ab.OpUpsertWeeklyCalendar, &ab.UpsertWeeklyCalendarArgs{
			TenantId: TenantID, StaffId: "staff-tony", DayOfWeek: dow,
			WorkStart: 480, WorkFinish: 840, BreakStart: 0, BreakFinish: 0, IsActive: true,
		})
	}
	// Thor: sin filas — agenda "recién creada".
}

func (e *Env) addExceptions() {
	call := func(op string, args model.Encodable) {
		var doneErr error
		e.caller.Call(op, args, nil, func(err error) { doneErr = err })
		if doneErr != nil {
			panic(doneErr)
		}
	}
	// Excepciones de Natasha: 18 y 19 de septiembre 2026 cerrados.
	for _, d := range []string{"2026-09-18", "2026-09-19"} {
		call(ab.OpAddCalendarException, &ab.AddCalendarExceptionArgs{
			TenantId: TenantID, StaffId: "staff-natasha",
			SpecificDate: unixDay(d), ExceptionType: ab.ExcHoliday,
		})
	}
}

func (e *Env) seedEmployeeServiceConfig() {
	// employee_service_config NO tiene op en appointment_booking (no hay UI
	// staff↔servicio en este módulo) — se siembra directo con db.Create. Sus
	// ids son estables y los consumen create_reservation / list_availability
	// como config_id. En producción viven en un screen staff↔servicio, fuera
	// de alcance de la demo.
	e.db.Create(&ab.EmployeeServiceConfig{
		Id:          "escNatashaConsulta",
		TenantId:    TenantID,
		StaffId:     "staff-natasha",
		ServiceId:   "svc-consulta",
		DurationMin: 30,
		BufferMin:   0,
		IsActive:    true,
	})
	e.db.Create(&ab.EmployeeServiceConfig{
		Id:          "escTonyCirugia",
		TenantId:    TenantID,
		StaffId:     "staff-tony",
		ServiceId:   "svc-cirugia",
		DurationMin: 30,
		BufferMin:   0,
		IsActive:    true,
	})
	e.db.Create(&ab.EmployeeServiceConfig{
		Id:          "escThorEcografia",
		TenantId:    TenantID,
		StaffId:     "staff-thor",
		ServiceId:   "svc-ecografia",
		DurationMin: 30,
		BufferMin:   0,
		IsActive:    true,
	})
}

// seedReservations siembra reservas CONFIRMED directo con db.Create. La op
// create_reservation valida el slot contra list_availability, lo que hace un
// seed por la op frágil; directo es la forma honesta de sembrar datos
// históricos. Los días caen en la plantilla de Natasha (Lun–Vie).
func (e *Env) seedReservations() {
	seeds := []*ab.Reservation{
		{
			Id: "res-1", TenantId: TenantID, ClientId: "12345678-9",
			CreatorUserId: "demo", EmployeeServiceConfigId: "escNatashaConsulta",
			StaffIdsnapshot: "staff-natasha", ServiceIdsnapshot: "svc-consulta",
			DurationMinSnapshot: 30, Status: ab.StatusConfirmed,
			ReservationDate: unixDay("2026-09-10"), ReservationTime: unixDay("2026-09-10") + 9*3600,
			LocalStringDate: "2026-09-10", LocalStringTime: "09:00",
			Notes: "María Gonzalez", UpdatedAt: unixDay("2026-09-09") * 1000000000,
		},
		{
			Id: "res-2", TenantId: TenantID, ClientId: "98765432-1",
			CreatorUserId: "demo", EmployeeServiceConfigId: "escNatashaConsulta",
			StaffIdsnapshot: "staff-natasha", ServiceIdsnapshot: "svc-consulta",
			DurationMinSnapshot: 30, Status: ab.StatusConfirmed,
			ReservationDate: unixDay("2026-09-10"), ReservationTime: unixDay("2026-09-10") + 10*3600 + 30*60,
			LocalStringDate: "2026-09-10", LocalStringTime: "10:30",
			Notes: "Juan Pérez", UpdatedAt: unixDay("2026-09-09") * 1000000000,
		},
		{
			Id: "res-3", TenantId: TenantID, ClientId: "11223344-5",
			CreatorUserId: "demo", EmployeeServiceConfigId: "escNatashaConsulta",
			StaffIdsnapshot: "staff-natasha", ServiceIdsnapshot: "svc-consulta",
			DurationMinSnapshot: 30, Status: ab.StatusConfirmed,
			ReservationDate: unixDay("2026-09-11"), ReservationTime: unixDay("2026-09-11") + 9*3600,
			LocalStringDate: "2026-09-11", LocalStringTime: "09:00",
			Notes: "Ana Silva", UpdatedAt: unixDay("2026-09-10") * 1000000000,
		},
	}
	for _, r := range seeds {
		e.db.Create(r)
	}
}

// unixDay convierte "YYYY-MM-DD" a medianoche UTC en segundos (la forma que
// work_calendar_exception.specific_date y reservation.reservation_date
// almacenan).
func unixDay(dateStr string) int64 {
	nano, err := tinytime.ParseDate(dateStr)
	if err != nil {
		panic(err)
	}
	return nano / 1000000000
}

// stubStaff/stubCatalog/stubDirectory satisfacen los readers que
// appointment_booking exige: la demo confía en su propio seed, así que todo
// lo que la UI ofrece existe, y las validaciones de fin lo confirman.
type stubStaff struct{}

func (stubStaff) StaffExists(_, _ string) (bool, error) { return true, nil }

type stubCatalog struct{}

func (stubCatalog) ServiceExists(_, _ string) (bool, error) { return true, nil }

type stubDirectory struct{}

func (stubDirectory) ClientExists(_, _ string) (bool, error) { return true, nil }
