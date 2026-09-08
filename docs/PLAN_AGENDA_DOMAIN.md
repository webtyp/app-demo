---
PLAN: "feat!(demo): business_calendar wired, blocks + marked days, conflicts surfaced"
EXECUTOR: local
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Terminal stage — depends on four GATES**, in order:
> `webtyp/events` (`Fanout`) → `veltylabs/business_calendar` →
> `veltylabs/appointment_booking` → `webtyp/components`. Bump all in `go.mod`
> before starting.
>
> **It also depends on [PLAN_REAL_BACKEND.md](PLAN_REAL_BACKEND.md) landing
> first**, in this same repo: that plan deletes `config/env.go` and replaces the
> in-WASM `loopback` composition with a real server (`config/server.go`,
> `routes/routes.go`, HTTP + SSE). Every reference below to "the composition
> root" means **`config/server.go`**, not the deleted `config/env.go`.
> Orchestrator:
> [webtyp/docs/AGENDA_DOMAIN_MASTER_PLAN.md](https://github.com/webtyp/webtyp/blob/main/docs/AGENDA_DOMAIN_MASTER_PLAN.md).
>
> Queued in [PLAN.md](PLAN.md). Read `AGENTS.md` in this repo root first.

# Plan — cablear el dominio de agenda en la demo

## 1. What this stage proves

`app-demo` is the canonical "this is how you build with WebTyp" application. It
must show the whole domain working over the **real** veltylabs modules, with
nothing reimplemented locally:

- an administrator editing the establishment's hours, holidays and closures;
- a professional editing a weekly pattern **and** marking concrete days;
- the professional's hour selects bounded by what the administrator allows, and
  widening the moment the administrator widens them;
- a schedule change surfacing the reservations it put in conflict.

## 2. Use cases

CU-03, CU-17, CU-21, CU-22, CU-23, CU-24 from the master plan §4, plus the
end-to-end wiring that makes CU-01…CU-20 reachable from the running app.

## 3. Stage 1 — `business_calendar` in the composition root

**File: `config/server.go`** (created by
[PLAN_REAL_BACKEND.md](PLAN_REAL_BACKEND.md) §5; `config/env.go` no longer
exists).

`config.ServerModules()` already assembles the shared runtime: an in-memory
`orm.DB` (`storage/mem`), a `mock.Broker`, an `sse.Publisher`, the
`events.Fanout` that joins them, and the veltylabs modules it returns for
`routes.Register` to mount. Add `business_calendar` the same way the other real
modules are added, **before** `appointment_booking` (which consumes it).

**This app is where the two domain modules meet.** Neither imports the other
(master plan §3-bis): `appointment_booking` declares its own `BoundsReader` and
the value crosses as `tinytime.DayBounds`. `business_calendar` satisfies that
port **structurally**, so there is no adapter to write — but the *composition*
and the *event wiring* are this file's job, because only the app legitimately
knows both.

```go
pub := events.Fanout{broker, sse.Publisher{Server: sseSrv}}

cal, _  := businesscalendar.New(db, businesscalendar.Deps{IDs: ids, Publisher: pub})
book, _ := appointmentbooking.New(db, appointmentbooking.Deps{
    IDs:       ids,
    Publisher: pub,  // reaches in-proc subscribers AND the browser
    Bounds:    cal,  // satisfies BoundsReader structurally — no import either way
})

// appointment_booking does NOT subscribe: that would need business_calendar's
// topic constant, i.e. the import we just avoided. The app owns the trigger.
broker.Subscribe(businesscalendar.EventCalendarChanged, func(ev events.Event) {
    var p businesscalendar.CalendarChangedPayload
    // decode ev.Payload into p …
    if !p.Closed {
        return // opening time can never invalidate a booking (CU-29)
    }
    _, _ = book.RecomputeConflicts(config.TenantID, p.FromDate, p.ToDate)
})
```

Three things to get right:

- **Subscribe on the broker, never on the `Fanout` or `sse.Publisher`.** SSE is
  one-way server→browser and cannot deliver to an in-process handler; `Fanout`
  is a `Publisher` only. Decision B4 of the backend plan.
- **One broker instance** shared by both modules — two brokers means the handler
  never fires and CU-25 silently does nothing.
- **Construction order**: `business_calendar` first (it supplies `Bounds`), then
  `appointment_booking`, then the subscription (it closes over `book`).

**`config/client.go` gets none of this.** The browser has no database, no broker
of its own beyond the SSE-fed one, and no module instances; it reaches every op
through `mcp.NewCaller`.

**File: [`config/holidays.go`](../config/holidays.go) — DELETE THE FILE.**
(If the backend plan has not already removed it while collapsing `config/`, it
goes here.)

`holidaysCL2026()` is a compiled-in Go function. It is exactly the defect this
whole plan removes: a holiday appearing by law must never require a recompile.
Its twelve dates move into the seed, written through the **real**
`add_holiday` op, so the demo exercises the same path a real administrator does.

Acceptance: `grep -rn "holidaysCL2026\|Holidays2026" .` → empty.

## 4. Stage 2 — seed

**File: `config/server.go`.** Extend the seed step (§5.7 of the backend plan),
still written through the **real ops** so the demo exercises the same path a
real administrator does:

- **Business hours** via `upsert_business_hours`: Mon–Fri 08:00–20:00, Sat
  09:00–14:00, Sun closed. In **minutes int** (480/1200, 540/840) — the new
  module's encoding, not `"HH:MM"`.
- **Holidays** via `add_holiday`: the twelve Chilean 2026 dates that
  `holidays.go` held, each with its name ("Año Nuevo", "Fiestas Patrias", …).
- **One closure** via `add_closure`, so CU-06 is visible in the running demo and
  distinguishable from a holiday.
- **Blocks** replacing `upsertWeekly()`:
  - *Natasha* — pattern Mon–Fri, **two blocks a day**: 09:00–13:00 and
    14:00–18:00. This is the CU-09 case, and it is what proves the lunch break
    is a gap and not a field.
  - *Tony* — pattern Mon/Wed/Fri 08:00–14:00, one block.
  - *Thor* — **no weekly pattern at all**, and ~8 marked days spread over the
    next two months via `mark_working_days`. This is the CU-10 case: the
    irregular professional who was impossible to express before.
- **One reservation on a day Tony works**, positioned so that a later narrowing
  of his schedule puts it in conflict — CU-16's fixture, and the same one CU-25
  uses when that date is declared a holiday.
- **One late reservation** (e.g. 19:00 on a Monday, inside the seeded 08:00–20:00
  establishment window) so that narrowing Monday to 18:00 puts it in conflict
  without touching any professional's blocks — CU-27's fixture.

Existing seed rules hold: real ops where an op exists, `db.Create` only where
the module exposes none, and every exception documented in its method.

## 5. Stage 3 — `agenda` consumes `rightpanel` (CU-23)

**File: [`modules/agenda/agenda.go`](../modules/agenda/agenda.go).**

Today `agenda` hand-writes an `<h1>` and carries its own `css.go`, while every
`crudview` module gets its title from `rightpanel.Title`. Two mechanisms, one of
which produced the inconsistency the owner reported.

`ScheduleView.Render()` returns a `rightpanel.RightPanel`:

```go
&rightpanel.RightPanel{
    Title:        "Agenda",
    HeadControls: staffpick.Select(...),   // the professional picker
    Article:      editorNodes,             // the ScheduleEditor
}
```

Read [`rightpanel.go`](https://github.com/webtyp/layout/blob/main/rightpanel/rightpanel.go)
first: `Title` renders as `<h1>`, `HeadControls` sits below the title row, and
`Article` is the main content area. Every slot must embed `dom.Element` **by
value**.

Then **delete** `modules/agenda/css.go` and the `NameAgenda`/`Part*`/`cls*`
declarations it fed. The chrome comes from `rightpanel` now; keeping a local
stylesheet beside it is the second mechanism all over again.

Do the same for [`modules/workschedule`](../modules/workschedule/workschedule.go)
— it received the same hand-rolled treatment and must not diverge.

Acceptance: `grep -rln "H1()" modules/` → empty.

## 6. Stage 4 — the professional's editor

**File: `modules/agenda/agenda.go`.** Migrate to `scheduleeditor` v2:

- `Week []WeeklyRow` → `Pattern []PatternRow` + `Marked []MarkedDay`, read from
  `list_blocks`.
- `OnWeeklyChange` → `OnPatternChange` (→ `save_day_blocks`), `OnDaysMarked`
  (→ `mark_working_days`), `OnDaysUnmarked` (→ `unmark_working_days`),
  `OnMarkedDayEdit` (→ `save_date_blocks`).
- **`Bounds`** from the new `get_day_bounds` op — this is what makes CU-02 real
  in the running app.
- `Holidays` and `Closures` from `list_holidays` / `list_closures`, no longer
  from a Go constant.

**CU-03's live wiring.** The professional's editor must reflect widened hours
without a manual reload. `business_calendar` publishes
`business.calendar.changed`; `agenda` subscribes through the demo's broker and
re-reads its bounds, exactly as `reservation` already subscribes to
`schedule.changed` (`README.md` → "Sincronía por eventos"). Copy that pattern;
do not invent a second one.

## 7. Stage 5 — the administrator's module

**New package `modules/businesscalendar/`.** Three `crudview` mounts over the
real presenters, following the `itemcatalogdemo` pattern (a thin wrapper, no
custom config):

- `NewHours(p, env)` — `business_calendar.NewBusinessHoursView`, label "Horario
  del local".
- `NewHolidays(p, env)` — `NewHolidaysView`, label "Feriados".
- `NewClosures(p, env)` — `NewClosuresView`, label "Cierres".

Plus `svg.go` (`//go:build !wasm`) with one icon per module, modelled on
[`modules/agenda/svg.go`](../modules/agenda/svg.go).

Each is one line in [`web/client.go`](../web/client.go)'s `p.Modules`.

## 8. Stage 6 — conflicts surfaced (CU-17)

**New package `modules/conflicts/`.** A `crudview` over
`list_conflicting_reservations`, label "Reservas en conflicto", showing the
patient, the date and the reason. It subscribes to `schedule.changed` and
reloads when `ConflictCount > 0`.

This is the administrator's worklist — the "¿dónde está eso?" the owner asked.
Notification **transport** (email/SMS to the patient, CU-18) is out of scope for
the demo: `webtyp/sse` already satisfies `events.Publisher`, and the demo's
broker is in-process. State that explicitly in `README.md` rather than faking a
delivery that does not happen.

## 9. Stage 7 — dictionary

**File: [`config/lang.go`](../config/lang.go).** Register the English keys the
two new modules introduce:

- from `business_calendar`: `Business hours`, `Holidays`, `Closures`, `Open`,
  `Closed`;
- from `scheduleeditor` v2: `Add row`, `Remove row`, `Marked days`, `Weekly
  pattern`, `Hours for marked days`, and the seven weekday short names.

Delete the keys of the removed `scheduleeditor` chrome (`Work start`, `Work
end`, `Break start`, `Break end`, `No break`) — dead entries are debt. Verify
each is truly unused before deleting: `grep -rn '"Break start"' .`

## 10. Stage 8 — docs

- `README.md` — the module index gains `businesscalendar` and `conflicts`; the
  `agenda` entry says blocks + marked days; a note that notification transport
  is out of scope and why.
- `AGENTS.md` — only if the module shape changed.
- [`docs/PLAN.md`](PLAN.md) — add this stage's row and keep the index current.

## 11. Constraints

- **No stdlib in WASM code**: `webtyp.com/fmt`. `_test.go` may use `strings`.
- **`dom.Element` embedded by VALUE**, never a pointer — `rightpanel` slots
  included.
- **SSR split by extension**: `svg.go` with `//go:build !wasm`. Never `ssr.go`,
  never `front.go`.
- **Never compose an id string in `Render()`** — `Key(...)` + `el.Ref()`.
- **No `map`** in code reaching WASM.
- **No `if dev`, no test-only routes.** `storage/mem` with a realistic seed is
  the real store.
- **UI strings Spanish, identifiers English.**
- **Nothing deprecated.** `holidays.go` and `agenda/css.go` are **deleted**.
- `gotest`, never `go test`.

## 12. Acceptance criteria

### 12.1 Static

| # | Check | Expected |
|---|-------|----------|
| 1 | `gotest ./...` | green |
| 2 | `GOOS=js GOARCH=wasm go build ./web/` | compiles |
| 3 | `go list -deps ./web/ \| grep webtyp/svg/sprite` | **empty** |
| 4 | `grep -rn "holidaysCL2026\|Holidays2026" .` | **empty** |
| 5 | `ls modules/agenda/css.go` | **no such file** |
| 6 | `grep -rln "H1()" modules/` | **empty** — titles come from `rightpanel` |
| 7 | `grep -rn "WeeklyRow\|OnWeeklyChange" modules/` | **empty** |
| 8 | `grep -rn '"Break start"\|"No break"' config/lang.go` | **empty** |

### 12.2 Live — drive the running app with the `webtyp` MCP browser tools

`restart_development`, then **https**://localhost:8080 (http returns 400 — the
dev server speaks TLS).

| # | Check | CU |
|---|-------|-----|
| 9 | `#agenda` shows a weekly pattern **and** a day marker in one view | CU-24 |
| 10 | Natasha renders **two blocks** on a weekday, with the gap between them | CU-09 |
| 11 | Thor has **no pattern** and his marked days still offer slots in `#reservation` | **CU-10** |
| 12 | An hour select offers nothing before 08:00 or after 20:00 | **CU-02** |
| 13 | Widening Monday to 22:00 in `#businesscalendar-hours`, then returning to `#agenda`, offers hours up to 22:00 **with no reload** | **CU-03** |
| 14 | Adding a holiday makes that day unselectable in the marker and dropped from `#reservation` | CU-04 |
| 15 | Narrowing Tony's schedule surfaces the seeded reservation in `#conflicts` | **CU-16/17** |
| 16 | Narrowing a schedule with no affected reservations leaves `#conflicts` empty | **CU-19** |
| 16b | Declaring a **holiday** on the day of the seeded reservation surfaces it in `#conflicts`, with reason "día cerrado" | **CU-25** |
| 16c | **Removing** that holiday empties `#conflicts` again, with no manual step | **CU-28** |
| 16d | Narrowing the establishment's Monday hours under a late reservation surfaces it | **CU-27** |
| 17 | `#agenda`'s title is the same `rightpanel__title` element every crudview module uses | **CU-23** |
| 18 | `browser_audit_mobile` on the editor reports 0 real tap targets under 44×44 | CU-24 |
| 19 | On iPhone 14 Pro emulation the editor fits without a 900px exceptions panel | CU-24 |

## 13. Stages

| # | Stage | Files | Done when |
|---|-------|-------|-----------|
| 1 | Composition root | `config/server.go`, **delete** `config/holidays.go` | checks 4 |
| 2 | Seed | `config/server.go` | Natasha 2 blocks, Thor marked-only, 2 conflict fixtures |
| 3 | `rightpanel` | `modules/agenda/`, `modules/workschedule/`, **delete** `agenda/css.go` | checks 5, 6, 17 |
| 4 | Editor v2 | `modules/agenda/agenda.go` | checks 7, 9–13 |
| 5 | Admin module | **new** `modules/businesscalendar/`, `web/client.go` | checks 13, 14 |
| 6 | Conflicts | **new** `modules/conflicts/`, `web/client.go` | checks 15, 16 |
| 7 | Dictionary | `config/lang.go` | check 8 |
| 8 | Docs | `README.md`, `docs/PLAN.md` | index current |

## 14. What this plan does NOT do

- **No notification transport.** The demo surfaces conflicts in a module and
  publishes events; email/SMS to the patient (CU-18) needs a real channel and is
  stated as out of scope in `README.md`.
- **No permissions UI.** The ops already declare `Requires(resource, action)`;
  the demo runs as a single administrator identity.
- **No migration of existing data.** The store is in-memory and reseeded on
  every load.
