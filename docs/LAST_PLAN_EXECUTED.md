# PLAN_WORK_SCHEDULE — módulo demo "Agenda" (Etapa D del `DEMO_AGENDA_MASTER_PLAN`)

Orquestador: `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md` §7 fila D.
Índice del repo: [PLAN.md](PLAN.md).

**Depende de (publicado):** `components` con `scheduleeditor` (A), `router` con
`loopback` (B), `appointment_booking` con `ScheduleClient` + list-ops + evento (C).
`work_schedule` con `NewView` (C2) es **opcional** — si no está, se omite el
panel "horario legado".

## Objetivo

Primer corte **vertical y testeable en `localhost:8080`**: un módulo demo nuevo
`modules/agenda/` (Label "Agenda") que deja crear/ajustar la agenda semanal +
excepciones de cualquier profesional, sobre el módulo **real**
`appointment_booking`.

## 1. `go.mod` — importar módulos reales

```
require (
    github.com/veltylabs/appointment_booking v0.0.0
    github.com/veltylabs/work_schedule v0.0.0   // opcional (C2)
    webtyp.com/events v0.0.3                     // para el broker in-proc (mock.Broker)
)
replace github.com/veltylabs/appointment_booking => ../../veltylabs/modules/appointment_booking
replace github.com/veltylabs/work_schedule       => ../../veltylabs/modules/work_schedule
```

- `webtyp.com/events` hoy **no está** en `app-demo/go.mod` (ni directo ni
  indirecto) — agregarlo.
- `webtyp.com/router` hoy es **indirecto**; `demoenv` lo usa directo
  (`router/loopback`) → pasará a directo tras `go mod tidy`.
- Mismo patrón de `replace` que `webtyp.com/components => ../components` (ya
  existe). Quitar cada `replace` cuando el repo esté publicado con el tag.
- Path de los `replace`: `app-demo` está en `webtyp/app-demo`, los módulos en
  `veltylabs/modules/*` → `../../veltylabs/modules/<m>` es correcto.

## 2. `demoenv/` — paquete nuevo (composición compartida)

Un solo lugar que construye, **una vez**, la infra que todos los wrappers de
módulo comparten. No es lógica de dominio (es el seam de composición), así que
no viola "módulos autocontenidos" del AGENTS.md — es el equivalente demo de lo
que `config/client.go` hace en `mjosefa-cms`.

```go
package demoenv

// Env es la composición de la demo: un orm.DB en memoria sembrado, los módulos
// de dominio reales montados, y un router.Caller in-proc que los maneja.
type Env struct { /* privado */ }

func New() *Env                    // construye DB mem + módulos + loopback + seed
func (e *Env) Caller() router.Caller
func (e *Env) Broker() events.Broker      // events/mock (o events/inproc — ver master O1)
func (e *Env) IDs() model.IDGenerator     // unixid
func (e *Env) TenantID() string           // "demo"
func (e *Env) Staff() []StaffOption        // {ID, Name, SpecialtySlug} — para los pickers
func (e *Env) Holidays2026() []string      // feriados CL "YYYY-MM-DD", solo lectura
```

Construcción interna:

1. `db := orm.New(mem.New())`.
2. `ids, _ := unixid.NewUnixID()`; `broker := &mock.Broker{}` (`webtyp.com/events/mock`).
3. Instanciar los módulos reales:
   - `ab, err := appointmentbooking.New(db, appointmentbooking.Deps{Staff: stubStaff{}, Catalog: stubCatalog{}, Directory: stubDirectory{}, IDs: ids, Publisher: broker})`
     — `New` devuelve `(*Module, error)`; en la composición `if err != nil { panic(err) }`
     (como hace `reservation.go` hoy). `stubStaff/stubCatalog/stubDirectory` son
     `struct{}` que devuelven `true, nil` a `StaffExists`/`ServiceExists`/`ClientExists`
     (la demo confía en su propio seed — documentarlo con un comentario que diga
     qué haría un Directory/Staff real).
   - `ws := workschedule.New(db)` (solo si se incluye C2; devuelve `*Module`).
4. `caller := loopback.New(ab, ws)` — `ab` y `ws` implementan
   `router.OperationModule` (verificado: `var _ router.OperationModule` en ambos).
5. **Seed:**
   - **No hay tabla `staff` que sembrar para `appointment_booking`**: valida
     existencia por el `StaffReader` inyectado (el stub devuelve `true`). Basta
     con `[]StaffOption` en memoria con IDs estables (`"staff-tony"`,
     `"staff-natasha"`, `"staff-thor"`) + `SpecialtySlug` — usar slugs de
     `itemcatalog.CanonicalSpecialties` (p.ej. `"traumatologia"`,
     `"medicina-general"`, `"ecografia"`), **no** slugs inventados (ver
     `PLAN_CATALOG.md` §2).
   - **Agenda** (vía las ops reales — ejercita el camino de producción):
     - Por staff: `caller.Call(OpUpsertCalendarConfig, {TenantId:"demo", StaffId, Timezone:"America/Santiago", IsActive:true}, nil, done)`.
     - Natasha: `OpUpsertWeeklyCalendar` Lun–Vie (`day_of_week` 1..5),
       `work_start=540`, `work_finish=1080`, `break_start=780`, `break_finish=840`,
       `is_active=true`. Tony: Lun/Mié/Vie 08:00–14:00 sin colación
       (`break_start=break_finish=0`). Thor: sin filas (agenda "recién creada").
     - 2 excepciones Natasha: `OpAddCalendarException` HOLIDAY 2026-09-18 y
       2026-09-19 (`specific_date` = medianoche UTC vía `webtyp/time`).
   - **`employee_service_config`** — **NO tiene op** en `appointment_booking`
     (revisado: `Repository.InsertEmployeeServiceConfig` existe, ninguna op lo
     expone). Sembrar con `db.Create(&appointmentbooking.EmployeeServiceConfig{Id: "esc-natasha-consulta", TenantId: "demo", StaffId: "staff-natasha", ServiceId: "svc-consulta", DurationMin: 30, BufferMin: 0, IsActive: true})`
     — `Id` preseteado y estable (lo necesitan `create_reservation` y
     `list_availability` como su `config_id`). Documentar: no hay UI para esto en
     la demo (sería una pantalla staff↔servicio, fuera de alcance).
   - **`staff` + `workcalendar` legadas** — solo si **C2** está incluido:
     `work_schedule.GetWorkSchedule` sí lee esas tablas (esquema legado, texto).
     Sembrarlas con `db.Create(&workschedule.Staff{...})` y
     `db.Create(&workschedule.WorkCalendar{...})`. Son tablas **distintas** de las
     de `appointment_booking` (a propósito — ver master §8). En `storage/mem` no
     hace falta migrar: `db.Create` basta.
   - `reservation` sembradas → **diferir a la Etapa F** (`OpCreateReservation`
     valida contra `list_availability`, frágil para seed; F usa `db.Create`).
6. `Holidays2026()` — lista fija CL: `01-01, 09-18, 09-19, 10-12, 10-31, 11-01,
   12-08, 12-25` (prefijo `2026-`). Comentario: en producción viene de un
   servicio de feriados; en la demo es constante.

`demoenv` compila a WASM (el `Caller` vive client-side). `loopback` ya es
`map`-free (ver su plan). El `map` de `mock.Broker` es lo que pesa la decisión
**O1** del master — resolver al llegar a la Etapa G, no bloquea D.

## 3. `modules/agenda/` — el módulo demo

Nombre `agenda` (Label "Agenda"), **no** `work_schedule`, para no chocar con el
import del módulo real `workschedule "github.com/veltylabs/work_schedule"` que
`demoenv` ya usa. Ficheros (patrón AGENTS.md, planos):

```
modules/agenda/
  agenda.go   # Module{p, env}, New, ModelName/Label/Icon, View()
  svg.go      # //go:build !wasm — glifo (reloj/calendario)
```

No lleva `model.go`/`store.go`: los datos y la persistencia son de
`appointment_booking` real vía `env.Caller()`.

`View()` devuelve un componente propio (`type scheduleView struct { dom.Element; ... }`
con `Init` + `Render`) — **no** un `crudview`, el editor de agenda no es una
lista CRUD, y así el `Init` da un único punto para construir signals y la
suscripción a eventos (Etapa G):

- **Picker de profesional** arriba: un `<select>` (o
  `components/selectsearch.SelectSearch`) con `env.Staff()`. Signal
  `sel *dom.SignalString` con el staffId elegido (default: el primero).
- **`scheduleeditor.ScheduleEditor`** debajo. El editor recibe `Week`/`Exceptions`
  por campo (no es fuente de verdad), así que al cambiar `sel` o tras una
  escritura hay que **rehacer el subárbol del editor**: montar el
  `ScheduleEditor` dentro de un contenedor cuyo hijo se bindea a una señal
  (`dom.Show`/`BindChild`-equivalente) o se reemplaza con la API de `dom` para
  swap de nodo — NO llamar métodos del componente ya montado. Mismo modelo que
  `crudview.Reload()`. En cada rebuild:
  - `client := appointmentbooking.NewScheduleClient(env.Caller(), env.TenantID(), sel.Get())`
  - `client.Weekly(func(rows, err){...})` → mapear `[]WorkCalendarWeekly` a
    `[]scheduleeditor.WeeklyRow` (rellenar los 7 días: los que no vengan van
    `Active:false` con horas 0). `client.Exceptions(fromYearStart, toYearEnd, …)`
    → `[]scheduleeditor.Exception` (convertir `specific_date` int → "YYYY-MM-DD").
  - `Holidays: env.Holidays2026()`.
  - `OnWeeklyChange: func(r) { client.SaveWeeklyRow(toWCW(r), func(err){ notify; refetch }) }`
  - `OnExceptionAdd: func(x) { client.AddException(toWCE(x), func(err){ notify; refetch }) }`
  - `OnExceptionRemove: func(id) { client.RemoveException(id, func(err){ notify; refetch }) }`
- Toasts vía `m.p.Notify(Msg.Success/Error, …, platformd.Auto())` como los demás
  módulos.

Helpers de conversión (`toWCW`, `toWCE`, `dateToUnix`, `unixToDate`) locales al
módulo; si aparecen idénticos en la Etapa F, subirlos a `demoenv` (regla DRY).

## 4. `web/client.go`

Una línea nueva en la lista de módulos:

```go
env := demoenv.New()
p.Modules = []platformd.UIModule{
    devices.New(p),
    medicalhistory.New(p),
    reservation.New(p),
    agenda.New(p, env),   // ← nuevo (paquete webtyp.com/app-demo/modules/agenda)
    about.New(),
    hiddenModule{},
}
```

(El `env` se pasará también a `reservation` en la Etapa F.)

## 5. `config/lang.go`

Si `scheduleeditor` renderiza chrome traducible (revisar su doc: etiquetas de
tipo de excepción "Closed"/"Special hours"/"Blocked", días de semana), agregar
sus claves EN→ES al diccionario. No inventar una segunda `RegisterWords`.

## Tests (`gotest`)

- `demoenv_test.go` — `New()` no paniquea; `Caller()` responde
  `OpListWeeklyCalendar` para Natasha con 5 filas; `OpListExceptions` con 2.
- `modules/agenda/agenda_test.go` — `View()` renderiza el picker
  con 3 opciones; con `sel` = Natasha el árbol contiene la grilla de 7 días y
  las horas 09:00/18:00 en la fila del lunes. Un `OnWeeklyChange` simulado llama
  a la op de upsert (verificable releyendo `OpListWeeklyCalendar`).
- Consumer-shaped: manejar el flujo alta-excepción end-to-end contra el
  `appointment_booking` real (add → list muestra la nueva → remove → ya no está).

## Criterios de aceptación

- `gotest ./...` verde en `app-demo`.
- `GOOS=js GOARCH=wasm go build ./web/` OK; `go list -deps ./web/ | grep webtyp/svg/sprite` vacío.
- En `localhost:8080` (daemon `webtyp` corriendo): el módulo "Agenda" aparece en
  el rail; elegir Natasha muestra su semana; activar el sábado con 10:00–13:00 y
  Guardar → toast; recargar el módulo (no la página) → persiste. Agregar
  excepción "Cerrado" un día → aparece en la lista y marcada en el calendario;
  el 18 y 19 de septiembre ya vienen como feriado, no editables.
- `README.md` de `app-demo` lista el módulo nuevo y explica el `demoenv` +
  el `replace` a los módulos reales. `AGENTS.md` si cambia la forma (nuevo tipo
  de módulo: wrapper sobre módulo real + `demoenv`).

## Fuera de alcance

- Selects de área/médico en `reservation` (Etapa F).
- Recalcular `reservation` al cambiar la agenda (Etapa G).
- Panel "horario legado" (`work_schedule.NewView`) si C2 no está listo.
