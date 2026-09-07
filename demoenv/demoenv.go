// Package demoenv conecta los módulos de dominio REALES de veltylabs a la demo:
// un *orm.DB en memoria sembrado, los módulos montados en un router.Caller
// in-proc (router/loopback) y un events.Broker para la comunicación módulo→
// módulo. Es el seam de composición de la demo — el equivalente que
// config/client.go hace en mjosefa-cms — no lógica de dominio.
//
// Se compila a WASM: el Caller vive client-side y maneja los módulos sin
// servidor. Todo lo que hay aquí es WASM-safe (loopback es map-free; el único
// map del árbol es el de events/mock, acotado y auditado — ver la nota O1 del
// DEMO_AGENDA_MASTER_PLAN, a resolver en la Etapa G).
package demoenv

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

// slugMedicinaGeneral / slugTraumatologia / slugEcografia son los slugs
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

// Env es la composición completa de la demo. Se construye una sola vez en
// web/client.go y se comparte entre módulos.
type Env struct {
	caller   router.Caller
	broker   events.Broker
	ids      model.IDGenerator
	ab       *ab.Module
	db       *orm.DB
	staff    []StaffOption
	holidays []string
}

// New construye el entorno de la demo: DB en memoria, módulos reales montados
// en un loopback.Caller, y seed realista.
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

// Caller es el router.Caller in-proc que despacha contra el appointment_booking
// real montado.
func (e *Env) Caller() router.Caller  { return e.caller }
func (e *Env) Broker() events.Broker  { return e.broker }
func (e *Env) IDs() model.IDGenerator { return e.ids }
func (e *Env) TenantID() string       { return TenantID }
func (e *Env) Staff() []StaffOption   { return e.staff }

// Holidays2026 son los feriados de Chile 2026 ("YYYY-MM-DD"), solo lectura. En
// producción vienen de un servicio de feriados; aquí es una constante.
func (e *Env) Holidays2026() []string { return e.holidays }

// seed puebla la DB en memoria vía las ops REALES de appointment_booking cuando
// existe una op, y con db.Create directo cuando el módulo no la expone (es la
// única excepción, documentada abajo).
func (e *Env) seed() {
	e.upsertCalendarConfigs()
	e.upsertWeekly()
	e.addExceptions()
	e.seedEmployeeServiceConfig()
}

func (e *Env) upsertCalendarConfigs() {
	for _, s := range e.staff {
		call := func(op string, args model.Encodable) {
			var doneErr error
			e.caller.Call(op, args, nil, func(err error) { doneErr = err })
			if doneErr != nil {
				panic(doneErr)
			}
		}
		call(ab.OpUpsertCalendarConfig, &ab.UpsertCalendarConfigArgs{
			TenantId: TenantID, StaffId: s.ID, Timezone: "America/Santiago", IsActive: true,
		})
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
	// staff↔servicio en este módulo) — se siembra directo con db.Create. Su id
	// es estable y lo consumen list_availability / create_reservation como
	// config_id. En producción vive en un screen staff↔servicio, fuera de
	// alcance de la demo.
	e.db.Create(&ab.EmployeeServiceConfig{
		Id:          "esc-natasha-consulta",
		TenantId:    TenantID,
		StaffId:     "staff-natasha",
		ServiceId:   "svc-consulta",
		DurationMin: 30,
		BufferMin:   0,
		IsActive:    true,
	})
}

// unixDay convierte "YYYY-MM-DD" a medianoche UTC en segundos (la forma que
// work_calendar_exception.specific_date almacena).
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
