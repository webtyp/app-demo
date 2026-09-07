# PLAN_RESERVATION — reserva de hora v2 (Etapa F del `DEMO_AGENDA_MASTER_PLAN`)

Orquestador: `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md` §7 fila F.
Índice del repo: [PLAN.md](PLAN.md).

**Depende de (publicado):** `router/loopback` (B), `appointment_booking` list-ops
+ `ScheduleClient` + evento (C), `crudview.Config.Context` (E).

## Objetivo

El módulo demo `modules/reservation/` gana:

1. Un selector de **área** (especialidad) y **médico** en el slot `Context` del
   `crudview`. Elegir área filtra los médicos; elegir médico **acota** la lista y
   el calendario a ese profesional.
2. Su **lista** y su **calendario de ocupación** se leen de `appointment_booking`
   real (`list_reservations_by_staff` vía `env.Caller()`), y **crear** una
   reserva llama `create_reservation`. Los huecos libres son la Etapa G.

## Restricción clave — NO se usa `appointmentbooking.NewView`

`appointmentbooking.NewView(caller, tenantId, staffId)` devuelve un `view.Presenter`
sobre `appointmentbooking.Reservation`, cuyos campos son todos `model.Text()` /
`model.Int()` — **sin widgets de form**. `crudview.New` hace
`form.New(cfg.Presenter.Record(), ...)` y "a record with no widgets fails HERE,
loudly" (crud.go). Además ese Presenter es list/select-only (sin `Saver`).

Por eso `reservation` **mantiene su record local** (`reservation.Reservation`,
con `input.X()` en los campos de paciente + día + hora — la forma del
**formulario**) y un **lister/saver a medida** traduce a/desde las ops reales.
El `crudview` (calendario `Filter` + `targethour` `List`) no cambia de forma.

## Brecha conocida — sin módulo Directory

`appointment_booking.Reservation` referencia al paciente por `client_id` (lo
valida `DirectoryReader`). La demo **no tiene** módulo Directory. Igual que
`mjosefa-cms/modules/clinical_encounter` documenta su brecha de picker de
paciente: aquí el formulario mantiene los campos de paciente como **texto libre**
(run, nombre, contacto), y al crear se usa el `run` como `client_id` sintético;
`demoenv` usa un `stubDirectory` que devuelve siempre `true`. Documentarlo en el
`reservation.go` con un comentario que explique el touch-point y qué haría un
Directory real. No es deuda oculta: es el trade-off explícito de una demo sin ese
módulo.

## 1. Datos de contexto (en `demoenv`, ampliando la Etapa D)

- **Especialidades = las canónicas de `item_catalog`** (`itemcatalog.CanonicalSpecialties`:
  `medicina-general`, `dental`, `traumatologia`, `ecografia`, `radiologia`, …).
  Si la Etapa H ya está, `demoenv.Specialties()` las lee del catálogo real
  (`OpListSpecialties`); si no, `demoenv` las declara como constante con **esos
  mismos slugs** (H luego la reemplaza por la lectura). No inventar slugs.
- Cada `StaffOption` lleva `SpecialtySlug` (uno de esos slugs) — fijado en el
  seed de la Etapa D.
- `employee_service_config` seed por (staff, servicio) con `duration_min` real
  (Natasha consulta 30 min, Tony cirugía 20 min, etc.) — vía `db.Create` con
  `Id` estable (**no tiene op** — ver Etapa D §5). Este `Id` es el que
  `create_reservation` recibe como `employee_service_config_id` y el que
  `list_availability` recibe como su arg `config_id` (revisado en `service.go`:
  `ListAvailability` hace `GetEmployeeServiceConfig(configId)` — el `config_id`
  **NO** es el del `work_calendar_config`, es el del `employee_service_config`).
- Algunas `reservation` sembradas con **`db.Create(&appointmentbooking.Reservation{...})`
  directo** (documentado): `OpCreateReservation` valida el slot contra
  `list_availability`, lo que hace el seed frágil. Campos mínimos: `Id`,
  `TenantId`, `StaffIdsnapshot`, `EmployeeServiceConfigId`, `ReservationDate`
  (medianoche UTC), `ReservationTime`, `LocalStringDate`, `LocalStringTime`,
  `DurationMinSnapshot`, `Status: appointmentbooking.StatusConfirmed`, `ClientId`.

## 2. Slot `Context` — selects área + médico

Un componente `reservationContext` (local al módulo, `Render()` + signals):

```go
type reservationContext struct {
    dom.Element
    area   *dom.SignalString   // slug de especialidad, "" = todas
    staff  *dom.SignalString   // staffId elegido, "" = ninguno
    staffOptions []demoenv.StaffOption
    onStaffChange func(staffId string)   // el módulo lo cablea a re-scope + Reload
}
```

- `<select>` de área: opciones = especialidades + "Todas". `onchange` → set
  `area`, y recomputar las opciones del segundo select.
- `<select>` de médico: opciones = `staffOptions` filtradas por `area`
  (`SpecialtySlug == area` o todas si `area == ""`). `onchange` → set `staff` +
  `onStaffChange(staff.Get())`.
- CSS-first donde se pueda; el filtrado del segundo select sí necesita re-render
  de sus `<option>` → bindear con `BindChildrenFunc`/equivalente sobre `area`.

Se pasa a `crudview.Config.Context`. **No** es `widget.Filterable` (ver Etapa E):
el módulo cablea `onStaffChange` a mano.

## 3. Lister/saver a medida (sobre el record local)

- `Config.Presenter` = `view.New(<lister a medida>, &reservation.Reservation{}, view.WithTitle(...))`,
  envuelto en el `byDay` que ya existe (o su equivalente). El `crudview`
  (calendar `Filter`, `targethour` `List`) no cambia.
- **Un solo servicio por médico en la demo:** no hay picker de servicio (el
  `Context` es área+médico). `demoenv` expone `ESCForStaff(staffId) string` → el
  `employee_service_config.id` primario de ese médico (del seed). Documentarlo;
  un picker de servicio sería una ampliación.
- El lister:
  - `List()` → `env.Caller().Call(OpListReservationsByStaff, {TenantId, StaffId: ctx.staff.Get(), From, To}, &ReservationList{}, done)`
    (rango = un rango amplio fijo, p.ej. el año). `ctx.staff == ""` → lista vacía
    (como `medicalhistory` sin paciente). Mapea cada `appointmentbooking.Reservation`
    → `reservation.Reservation` local (`Hour = LocalStringTime`,
    `Day = LocalStringDate`, `PatientName`/`PatientRun` desde `Notes` o el sidecar,
    `Status` traducido).
  - `Filter(term)` (term = "YYYY-MM-DD" del calendario) → filtra en cliente por
    fecha, como hoy hace `byDay`.
  - `Save(rec)` → resuelve `employee_service_config_id = env.ESCForStaff(ctx.staff.Get())`
    + `slot_start_utc` desde (día del calendario, `rec.Hour`) con `webtyp/time`, y
    `env.Caller().Call(OpCreateReservation, {TenantId, ClientId: rec.PatientRun,
    CreatorUserId: "demo", EmployeeServiceConfigId, SlotStartUtc, Notes: <paciente serializado>}, nil, done)`.
    `ErrSlotTaken`/`ErrConflict` (status 409) → toast de error, sin crash ni
    duplicado.
- `onStaffChange` del `Context` → `crudview.Reload()` + limpiar selección.
- `view.Item` para `targethour`: `LeadMain = Hour`, `Label = PatientName`,
  `Description` = estado ES (`PENDING`→"", `CONFIRMED`→"Confirmada",
  `COMPLETED`→"Atendida"). `StatusOf` según ese estado.

## 4. `web/client.go`

`reservation.New(p)` pasa a `reservation.New(p, env)` (mismo `env` que
`work_schedule`).

## 5. Limpieza

- `modules/reservation/model.go`: el `Reservation` local **se conserva** — es la
  forma del formulario (widgets de paciente/día/hora). Se le puede agregar un
  campo `Area`/`Doctor` informativo si ayuda, pero el scope real lo lleva el
  `Context`, no el record.
- `modules/reservation/store.go`: el `reservationStore` local y su `orm.DB`
  sembrado **se eliminan** — la lista/creación van por `env.Caller()`. Los
  helpers de seed que sigan siendo útiles migran a `demoenv`.
- `reservation_test.go` / `store_*_test.go`: reescribir para el nuevo flujo
  (crear vía op real, listar por staff, filtrar por día).

## Tests (`gotest`)

- `TestContext_AreaFiltersStaff` — `area="ME"` → el select de médico solo lista
  Natasha (y quien más sea Medicina).
- `TestList_ScopedByStaff` — con staff = Natasha, la lista trae solo sus
  reservas; staff = "" → vacía.
- `TestCreate_GoesThroughRealOp` — un save con día+hora válidos crea una
  reserva que luego aparece en `OpListReservationsByStaff`.
- `TestCreate_SlotTaken` — crear sobre un slot ocupado → toast de error, sin
  duplicado.

## Criterios de aceptación

- `gotest ./...` verde; build WASM OK; sin fuga de sprite.
- En `localhost:8080`: elegir "Medicina" acota el select de médicos; elegir
  Natasha + un día muestra sus reservas de ese día en `targethour`; crear una
  reserva la agrega a la lista.
- `README.md` de `app-demo` actualizado (reservation ahora sobre módulo real +
  brecha Directory documentada).

## Fuera de alcance

- Huecos libres reservables + reacción al evento `schedule.changed` (Etapa G).
- Un módulo Directory real.
