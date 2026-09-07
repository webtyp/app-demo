# PLAN_EVENTS — sincronía agenda→reserva + huecos libres (Etapa G del `DEMO_AGENDA_MASTER_PLAN`)

Orquestador: `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md` §4.4, §7 fila G.
Índice del repo: [PLAN.md](PLAN.md).

**Depende de (publicado):** Etapa **A Parte 2** (`targethour.FreeSlots`), módulo
demo `work_schedule` (D) y `reservation` v2 (F). **Además:** resolver la decisión
abierta **O1** del master (broker in-proc para WASM).

## O1 — decidir primero

`demoenv.Broker()` hoy devuelve `*mock.Broker` (`webtyp.com/events/mock`), que
usa `map[string][]events.Handler`. Regla del ecosistema: sin `map` en WASM.
Opciones:

- **(a)** Aceptar el `map` de `events/mock`: está tras su paquete, acotado
  (pocos topics), poblado en el arranque. Cero trabajo. Anotar la excepción con
  comentario en `demoenv`.
- **(b)** Crear `webtyp.com/events/inproc`: `Broker` con
  `[]struct{topic string; h events.Handler}` y scan lineal en `Publish`.
  Mismo contrato (`var _ events.Broker`). Pequeño `PLAN.md` en `events/`.

Recomendación: **(a)** si es el único `map` del árbol WASM de la demo (probable);
**(b)** si el `loopback` registry ya obligó a otro y se quiere el árbol
`map`-free de verdad. Registrar la decisión aquí y en el master O1.

## 1. Cableado del broker (`demoenv` + `web/client.go`)

- `demoenv.New()` ya crea el broker y lo pasa como `Publisher` a
  `appointmentbooking.New(...)` (Etapa D). Confirmar que `Env.Broker()` expone el
  mismo objeto (mismo `events.Broker`, no una copia).
- En `mjosefa-cms` el equivalente será `webtyp/sse` (que ya satisface
  `events.Publisher`). Dejar un comentario en `demoenv` que lo diga: "aquí
  in-proc; en la app real, `sse`".

## 2. `reservation` se suscribe

En `reservation.New(p, env)` / su `View()`:

```go
env.Broker().Subscribe(appointmentbooking.EventScheduleChanged, func(e events.Event) {
    p, ok := e.Payload.(*appointmentbooking.ScheduleChangedPayload)
    if !ok { return }
    if p.StaffId != currentStaffID() { return } // solo si es el médico en foco
    reloadListAndSlots()
})
```

- Suscribir **una vez**. El módulo `reservation` no tiene `Init` propio (los
  módulos demo solo implementan `ModelName/Label/Icon/View`), así que suscribir
  en `View()` con una guarda de re-entrada (`if m.subscribed { ... }` sobre un
  bool del `*Module`), o —mejor— hacer que `reservation.View()` devuelva un
  componente propio con `Init` (como el `scheduleView` de la Etapa D) y suscribir
  ahí. Elegir lo segundo si no cuesta mucho.
- `reloadListAndSlots()` = `crudview.Reload()` + recomputar `FreeSlots` (§3).
- Suscribirse también a los eventos de reserva
  (`appointmentbooking.EventReservationCreated`, `...Cancelled`, `...Completed`)
  para que la lista/ocupación se refresquen al reservar desde la propia vista o
  (a futuro) desde otra pestaña.

## 3. Huecos libres reservables (`targethour.FreeSlots` — lo agrega la Etapa A P2)

- `components/targethour` **no tiene `FreeSlots` hoy**; se lo agrega la **Parte 2
  de la Etapa A** (`FreeSlots []string` + `OnPickFree func(hhmm string)`). Esta
  etapa **depende de que A P2 esté publicada**.
- Al elegir médico + día (o al recibir un evento), llamar
  `env.Caller().Call(OpListAvailability, {TenantId, StaffId, ConfigId: env.ESCForStaff(staffId), From: day, To: day}, &TimeSlotList{}, done)`.
  **`ConfigId` es el id del `employee_service_config`** (revisado en `service.go`:
  `ListAvailability` hace `GetEmployeeServiceConfig(configId)`), **no** el del
  `work_calendar_config`. `demoenv.ESCForStaff` (Etapa F) lo resuelve.
- Mapear `[]TimeSlot{StartUtc,EndUtc}` → `[]string` "HH:MM" (hora local, con
  `webtyp/time`) y setearlos en `targethour.FreeSlots`. `OnPickFree(hhmm)` →
  prellena la hora del form.
- Clic en un hueco libre → prellena la hora del form (y el día ya está del
  calendario) → el usuario completa paciente y Guarda → `create_reservation`.
- Tras crear, el evento `EventReservationCreated` dispara `reloadListAndSlots()`
  → el hueco recién tomado desaparece.

## 4. Verificación de la sincronía (el corazón de la feature)

Con `reservation` abierto en Natasha y un día laboral visible:
1. Los huecos libres de ese día se ven en `targethour`.
2. Ir al módulo "Agenda", quitar (o desactivar) ese día de la semana de Natasha,
   Guardar.
3. Volver a "Reserva Hora" **sin recargar la página**: los huecos de ese día
   desaparecieron (el `schedule.changed` gatilló `reloadListAndSlots`).
4. Reactivar el día en "Agenda" → los huecos vuelven.

## Tests (`gotest`)

- `TestScheduleChanged_ReloadsReservation` — con un doble de `crudview`/lister
  que cuenta `Reload()`: publicar `EventScheduleChanged` con el `staffId` en
  foco → `Reload` llamado una vez; con otro `staffId` → cero.
- `TestFreeSlots_DerivedFromAvailability` — sembrar agenda + 1 reserva; los
  `FreeSlots` calculados excluyen el slot reservado y los de la colación.
- `TestReservationCreated_RemovesFreeSlot` — crear en un hueco → recomputar →
  ese "HH:MM" ya no está en `FreeSlots`.

## Criterios de aceptación

- `gotest ./...` verde; build WASM OK; sin fuga de sprite.
- La verificación §4 pasa a mano en `localhost:8080`.
- Decisión O1 registrada (en este plan y en el master).
- `README.md` de `app-demo`: sección "Sincronía por eventos" explicando el
  broker in-proc y el swap a `sse` en producción.

## Fuera de alcance

- Construir `targethour.FreeSlots` — eso es la Etapa A Parte 2, no esta etapa.
- Push entre pestañas/dispositivos (eso es `sse`, en `mjosefa-cms`).
