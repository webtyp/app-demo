---
PLAN: "fix(agenda): the day is the index, the view has chrome, work_schedule gets its own module"
EXECUTOR: local
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Depends on two GATES**, both of which must ship first:
> [widget/docs/PLAN.md](https://github.com/webtyp/widget/blob/main/docs/PLAN.md)
> → [components/docs/PLAN.md](https://github.com/webtyp/components/blob/main/docs/PLAN.md).
> Bump `webtyp.com/components` in `go.mod` to the tag that ships the new
> `OnWeeklyChange` signature **before** Stage 1. Orchestrator:
> [webtyp/docs/AGENDA_VIEW_FIXES_MASTER_PLAN.md](https://github.com/webtyp/webtyp/blob/main/docs/AGENDA_VIEW_FIXES_MASTER_PLAN.md).

# Plan — `#agenda`: correctness, chrome, and one source of truth

---

## 1. What is wrong (observed live at `https://localhost:8080/#agenda`)

### 1.1 The view corrupts data on save (critical)

[`modules/agenda/agenda.go`](../modules/agenda/agenda.go) → `fallbackWeek`
builds `make([]scheduleeditor.WeeklyRow, 7)` and fills **only** the days
`appointment_booking` returned. Every unfilled row keeps its zero value, so
`DayOfWeek == 0`, and the editor renders four rows as "Domingo":

```
Domingo · Lunes · Domingo · Miércoles · Domingo · Viernes · Domingo
```

That row then travels back through `OnWeeklyChange` →
`ScheduleClient.SaveWeeklyRow` → `upsert_weekly_calendar`, which writes
`row.DayOfWeek` verbatim. **Reproduced live on `staff-tony`:** clicking the
toggle of row index 2 (Tuesday) left row 2 inactive and turned **row 0
(Sunday)** active. Tuesday, Thursday and Saturday all collide on
`day_of_week = 0`.

The tests missed it because they assert only
`strings.Contains(html, "09:00")` — shape, never the day names.

### 1.2 The view has no CSS at all (high)

`agenda.go` writes the classes `agenda`, `agenda__header`, `agenda__body`,
`agenda_schedule`. **Not one stylesheet rule mentions any of them.** Verified in
the running page by walking `document.styleSheets`: zero matches, and every one
of those elements computes `display:block; padding:0; margin:0`, with the
`<h2>` at `16px/400` — indistinguishable from body text.

That is the bulk of "it looks bad": the module is unstyled markup. The
`scheduleeditor` component styles its own `.scheduleeditor*` parts; nothing has
ever styled the module's own chrome. Note also `agenda_schedule` uses a single
underscore where every sibling uses the BEM `agenda__` — a third spelling for
a class nothing reads.

### 1.3 Two sources of truth on one screen (medium)

The "Horario" panel at the bottom is `work_schedule.NewView`, which reads the
**legacy** `staff`/`workcalendar` tables, while the editor above writes
`appointment_booking`'s `work_calendar_*`. In the §1.1 reproduction the panel
still read `Lunes / Miércoles / Viernes` after the write landed. It also
duplicates, unstyled and unlabelled, what the grid immediately above shows.

### 1.4 No page heading, no section labels (medium)

The view opens with a bare `Profesional [select]`. There is no `<h1>`, no
"Plantilla semanal", no "Excepciones" — the 3-month calendar appears with
nothing saying that picking a day adds an exception.

### 1.5 The staff `<select>` is 189×22px (medium)

`browser_audit_mobile` on `.agenda` flags it, below the 44×44 floor.

---

## 2. Stage 1 — migrate to the new component API

**File: [`modules/agenda/agenda.go`](../modules/agenda/agenda.go)**

`components` now declares (see its plan §3):

```go
type WeeklyRow struct {           // NO DayOfWeek field
	Active                  bool
	WorkStart, WorkFinish   int
	BreakStart, BreakFinish int
}

OnWeeklyChange func(dayOfWeek int, row WeeklyRow)
```

### 2.1 `fallbackWeek`

```go
// fallbackWeek returns exactly 7 rows, Sunday..Saturday, indexed by day: the
// rows the op returned land at their own index, the rest stay zero — a day
// with no row is a day that is not scheduled. The index IS the day, so an
// unfilled row can no longer claim to be Sunday.
func fallbackWeek(rows []ab.WorkCalendarWeekly) []scheduleeditor.WeeklyRow {
	week := make([]scheduleeditor.WeeklyRow, 7)
	for _, r := range rows {
		dow := int(r.DayOfWeek)
		if dow < 0 || dow > 6 {
			continue
		}
		week[dow] = scheduleeditor.WeeklyRow{
			Active:      r.IsActive,
			WorkStart:   int(r.WorkStart),
			WorkFinish:  int(r.WorkFinish),
			BreakStart:  int(r.BreakStart),
			BreakFinish: int(r.BreakFinish),
		}
	}
	return week
}
```

`ab.WorkCalendarWeekly.DayOfWeek` still exists — that is the **persistence**
type, where the day is a column and must be. Only the **editor** type drops it.

### 2.2 `toWCWeekly` takes the day

```go
func toWCWeekly(dayOfWeek int, r scheduleeditor.WeeklyRow) ab.WorkCalendarWeekly {
	return ab.WorkCalendarWeekly{
		DayOfWeek:   int64(dayOfWeek),
		IsActive:    r.Active,
		WorkStart:   int64(r.WorkStart),
		WorkFinish:  int64(r.WorkFinish),
		BreakStart:  int64(r.BreakStart),
		BreakFinish: int64(r.BreakFinish),
	}
}
```

### 2.3 The callback in `buildEditor`

```go
		OnWeeklyChange: func(dayOfWeek int, r scheduleeditor.WeeklyRow) {
			client.SaveWeeklyRow(toWCWeekly(dayOfWeek, r), func(err error) {
				s.notifySave(err)
				if err == nil {
					s.reloadEditor()
				}
			})
		},
```

`OnExceptionAdd` and `OnExceptionRemove` are unchanged.

---

## 3. Stage 2 — `work_schedule` gets its own module

### 3.1 Delete the panel from `agenda`

From [`modules/agenda/agenda.go`](../modules/agenda/agenda.go), delete **all**
of:

- the `schedule *SignalNodes` field on `scheduleView`,
- its `if s.schedule == nil { … }` block and the `s.reloadSchedule()` call in
  `Init`,
- the whole `reloadSchedule` method,
- the `agenda_schedule` `Div` (the `H2` + `Ul`) from `Render`,
- the `workschedule "github.com/veltylabs/work_schedule"` import.

Acceptance: `grep -rn "workschedule\|agenda_schedule\|reloadSchedule" modules/agenda/`
→ **empty**.

`env.WorkScheduleStaffID` in [`config/env.go`](../config/env.go) stays — the
new module is its caller. Do **not** delete it, and do not remove
`work_schedule` from `go.mod`.

### 3.2 The shared staff picker

Two modules now need the same plain staff `<select>` (agenda, and the new one),
so it is written once. **New package `modules/staffpick/`**, one file
`staffpick.go`:

```go
// Package staffpick is the demo's one staff <select>. Two modules need to scope
// a view to a professional; the control is written here so it is not spelled
// twice. It is demo chrome, not a framework component — a real app would get
// this from its staff module.
package staffpick

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"

	"webtyp.com/app-demo/config"
)

// Select builds the professional picker. sel holds the chosen StaffOption.ID;
// onChange fires after sel is updated, so the caller can reload.
func Select(staff []config.StaffOption, sel *SignalString, onChange func(id string)) *Element {
	el := NewElement("select").Attr("name", "staff-pick")
	for _, so := range staff {
		if so.ID == sel.Get() {
			el.Child(SelectedOption(so.ID, so.Name))
		} else {
			el.Child(Option(so.ID, so.Name))
		}
	}
	el.On("change", func(ev Event) {
		sel.Set(ev.TargetValue())
		if onChange != nil {
			onChange(ev.TargetValue())
		}
	})
	return el
}
```

`agenda.scheduleView.buildPicker` is **deleted** and replaced by a call to
`staffpick.Select(s.env.Staff(), s.sel, func(string) { s.reloadEditor() })`.

**Anti-footgun.** Do **not** migrate `modules/reservation`'s selects to this.
`reservationContext` holds a *pair* of coupled selects — an area select that
re-scopes the doctor select's options (`rebuildStaffOptions`). It is a different
control that happens to also list staff. Leave it exactly as it is.

### 3.3 The new module

**New package `modules/workschedule/`**, mirroring the thin-wrapper shape of
[`modules/itemcatalogdemo/items.go`](../modules/itemcatalogdemo/items.go):

- `workschedule.go` — a `Module` implementing `platformd.UIModule` with
  `ModelName() "workschedule"`, `Label() "Horario legado"`,
  `Icon() svg.Icon("mod-workschedule")`. Its `View()` returns a component that
  renders `staffpick.Select` plus the list from
  `workschedule.NewView(m.env.Caller(), m.env.WorkScheduleStaffID(id))`,
  rebuilt into a `SignalNodes` when the picker changes — the same
  `SignalNodes` swap `scheduleView` already uses for its editor. Copy that
  pattern; do not invent a second one.
- `svg.go` — `//go:build !wasm`, `func (m *Module) IconSvg() *sprite.Sprite`,
  modelled byte-for-byte on the shape of
  [`modules/agenda/svg.go`](../modules/agenda/svg.go) with a different glyph
  (a document/list, not a clock — the clock is agenda's).

`work_schedule.NewView(caller router.Caller, staffId int64) view.Presenter`
already carries its own title, `"Horario (sistema legado)"`. Read
`veltylabs/modules/work_schedule/view.go` before wiring it.

**File: [`web/client.go`](../web/client.go)** — one line in `p.Modules`, after
`agenda.New(p, env)`:

```go
		workschedule.New(p, env),  // ← work_schedule real, vista read-only sobre tablas legadas
```

plus the import. That is the whole registration — `web/client.go` is the
composition; adding a module is one line there.

---

## 4. Stage 3 — the view gets chrome

### 4.1 Why a widget, and the verification this needs

The `css` package's rule DSL (`rule`, `selector`, `padding`, …) is
**unexported** — an app package cannot build arbitrary rules with it, and
`config/css.go` only overrides tokens through `css.Theme`/`css.SetGradient`.
The one supported way to style a bespoke view is `widget/style`: declare the
view a `widget.Widget` and emit `RenderCSS()` from a `//go:build !wasm`
`css.go`. That is how `layout/crudview` and every `components/*` package do it.

**Risk, and the fallback if it does not hold.** No package under
`app-demo/modules/` declares `RenderCSS()` today, so it is unverified that
`sitec` discovers one there while walking the project. After Stage 3, build and
check:

```
grep -c "\.agenda" web/public/style.css
```

- **Non-zero** → discovery works, done.
- **Zero** → discovery is rooted at the composition root. Then move the sheet's
  construction into [`config/css.go`](../config/css.go), where `Theme` is
  already discovered, and have it call the agenda package's exported sheet
  builder. Do **not** hand-write a CSS string, and do **not** inline styles on
  elements — either would fork the token system.

### 4.2 The widget declaration

**File: `modules/agenda/agenda.go`** — add near the top:

```go
const NameAgenda = widget.Name("agenda")

const (
	PartHeader  = widget.Part("header")
	PartTitle   = widget.Part("title")
	PartSection = widget.Part("section")
	PartBody    = widget.Part("body")
)
```

`scheduleView` gains:

```go
func (s *scheduleView) WidgetName() widget.Name { return NameAgenda }
func (s *scheduleView) WidgetKind() widget.Kind { return widget.Form }
```

Every `Attr("class", "agenda…")` string literal in `Render` is replaced by the
`Name`/`Part`-derived class (`NameAgenda.Root().AsAttr()`,
`NameAgenda.Class(PartHeader).AsAttr()`, …) — the same mechanism
`scheduleeditor.go` uses. The hand-written class strings, `agenda_schedule`'s
odd single underscore included, are gone.

### 4.3 The headings

`Render` gets the structure the view never had:

```go
func (s *scheduleView) Render() *Element {
	return Div().Set(clsRoot.AsAttr()).
		Child(H1().Set(clsTitle.AsAttr()).Text("Agenda")).
		Child(Div().Set(clsHeader.AsAttr()).
			Child(Span().Text("Profesional")).
			Child(staffpick.Select(s.env.Staff(), s.sel, func(string) { s.reloadEditor() }))).
		Child(H2().Set(clsSection.AsAttr()).Text("Plantilla semanal")).
		Child(Div().Set(clsBody.AsAttr()).BindChildren(s.editor))
}
```

The "Excepciones" heading belongs to the exceptions panel, which lives **inside
`ScheduleEditor`** — it is not this module's to add, and this plan does not add
it here. If the heading is still missing after the component ships, that is a
follow-up on `components`, not a local patch. Never style around a library
defect from the consumer.

UI strings stay Spanish; identifiers stay English.

### 4.4 `modules/agenda/css.go` — new file

```go
//go:build !wasm

package agenda

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS is the agenda view's chrome. Before it, the module wrote four class
// names that no stylesheet mentioned: the view rendered as bare display:block
// with zero padding, flush to the viewport edge, and its <h2> was body text.
func (s *scheduleView) RenderCSS() *css.Stylesheet {
	return style.For(s).
		Root(
			style.Stack(style.Space4),
			style.Pad(style.Space4),
		).
		Part(PartTitle,
			style.FontSize(style.Text2xl),
			style.FontWeight(style.WeightBold),
		).
		Part(PartHeader,
			style.Row(style.Space2),
			style.Center(),
			style.ControlBox(),
			style.As(style.Panel),
			style.Round(style.RadiusMd),
		).
		Part(PartSection,
			style.FontSize(style.TextLg),
			style.FontWeight(style.WeightBold),
		).
		Part(PartBody,
			style.Stack(style.Space3),
		).
		Stylesheet()
}
```

`style.ControlBox()` on `PartHeader` is what lifts the 189×22 staff `<select>`
to a real control box and clears the §1.5 tap-target failure.

Every value above is a token. **Never** write a raw `px`/`rem` literal here —
the token scale (`Space1..Space12`, `TextXs..Text2xl`, `RadiusNone..RadiusFull`)
is the only vocabulary, and `style.Validate()` plus the golden CSS tests exist
to enforce it.

---

## 5. Stage 4 — the dictionary

**File: [`config/lang.go`](../config/lang.go)**, in the
`components/scheduleeditor` block.

**Delete** these two entries:

```go
		{EN: "Special", ES: "Especial"},
		{EN: "hours", ES: "horas"},
```

They produced **"Especial horas"**: `lang.Translate("Special", "hours")` is two
lookups joined by a space, and word-by-word translation cannot order a Spanish
adjective phrase. Nothing else uses either key — verify with
`grep -rn '"Special"\|"hours"' .` before deleting.

**Add**, keyed to the component's new single-key call:

```go
		{EN: "Special hours", ES: "Horario especial"},
		{EN: "No exceptions", ES: "Sin excepciones"},

		// components/scheduleeditor's weekly grid column headers.
		{EN: "Day", ES: "Día"},
		{EN: "Work start", ES: "Entrada"},
		{EN: "Work end", ES: "Salida"},
		{EN: "Break start", ES: "Colación desde"},
		{EN: "Break end", ES: "Colación hasta"},
```

Multi-word keys work: `lang.lookupWord` binary-searches the **whole** argument
string against the dictionary's EN column, case-insensitively
(`webtyp.com/fmt/lang/dictionary.go`).

`RegisterWords` sorts on registration, so entry order in the slice does not
matter — keep them grouped by owning component for the human reader.

---

## 6. Stage 5 — tests, README, index

### 6.1 The test that would have caught §1.1

**File: [`modules/agenda/agenda_test.go`](../modules/agenda/agenda_test.go)**

The existing tests assert `strings.Contains(html, "09:00")` — shape only. Add:

```go
// The seven rows carry the seven day names, in order. The bug this replaces:
// fallbackWeek left DayOfWeek at zero on unconfigured days and four rows
// rendered "Domingo" — and saving one of them wrote Sunday.
func TestEditor_RendersSevenDistinctDays(t *testing.T) {
	v, _ := testView(t, "staff-tony") // seeded Mon/Wed/Fri only — 4 rows unfilled
	html := v.buildEditor().Render().String()

	for _, day := range []string{"Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"} {
		if strings.Count(html, day) != 1 {
			t.Errorf("day %q must appear exactly once, got %d:\n%s", day, strings.Count(html, day), html)
		}
	}
}

// fallbackWeek indexes by day: a staff member seeded Mon/Wed/Fri gets active
// rows at 1, 3, 5 and inactive rows everywhere else.
func TestFallbackWeek_IndexesByDay(t *testing.T) {
	week := fallbackWeek([]ab.WorkCalendarWeekly{
		{DayOfWeek: 1, IsActive: true, WorkStart: 480, WorkFinish: 840},
		{DayOfWeek: 3, IsActive: true, WorkStart: 480, WorkFinish: 840},
		{DayOfWeek: 5, IsActive: true, WorkStart: 480, WorkFinish: 840},
	})

	if len(week) != 7 {
		t.Fatalf("week length = %d, want 7", len(week))
	}
	for i, want := range []bool{false, true, false, true, false, true, false} {
		if week[i].Active != want {
			t.Errorf("day %d: Active = %v, want %v", i, week[i].Active, want)
		}
	}
}

// toWCWeekly persists the day it is given, not one carried by the row — the
// row no longer carries one. This is the assertion that makes "enable Tuesday,
// save Sunday" unrepresentable.
func TestToWCWeekly_PersistsTheGivenDay(t *testing.T) {
	got := toWCWeekly(2, scheduleeditor.WeeklyRow{Active: true, WorkStart: 480, WorkFinish: 840})
	if got.DayOfWeek != 2 {
		t.Errorf("DayOfWeek = %d, want 2 (Tuesday)", got.DayOfWeek)
	}
}
```

The day names are Spanish because `config`'s `init()` registers the Spanish
dictionary and activates it — the existing tests already rely on that; do not
add a language switch.

Also **delete** `TestWorkSchedule_ViewsSchedule` from this file (it covers the
panel Stage 2 removes) and re-create the equivalent coverage in
`modules/workschedule/workschedule_test.go` against the new module.

### 6.2 `README.md`

- Line 17 (`modules/agenda`) — drop the "Horario" panel from the description.
- **Delete line 19** entirely (`modules/agenda` also consumes `work_schedule`).
- Add a `modules/workschedule` bullet, and a `modules/staffpick` note under the
  module list.

### 6.3 `docs/PLAN.md`

[`docs/PLAN.md`](PLAN.md) is this repo's index and is never deleted. Add a row
to its stage table pointing at this file, marked in-progress, and update it when
the work lands — the same way every previous stage was recorded.

---

## 7. Constraints — read before writing code

- **No standard library in WASM-compiled code.** `webtyp.com/fmt`, never
  `strconv`/`strings`/`errors`. `_test.go` files may use `strings` and
  `testing` — this repo's tests already do.
- **`dom.Element` embedded by VALUE**, never `*dom.Element`.
- **SSR split by extension:** `css.go` and `svg.go` carry `//go:build !wasm`.
  Never `ssr.go`, never `front.go` — both conventions are eliminated here.
- **Never compose an id string inside `Render()`.** `Key(...)` + `el.Ref()`, and
  `data-*` as the breadcrumb. See `DEMO_AGENDA_MASTER_PLAN.md`, "Fix de ids del
  arnés".
- **No `map`** in code reaching the WASM binary.
- **No `if dev`, no test-only routes.** The `storage/mem` store with a realistic
  seed is the real store here.
- **UI strings Spanish, identifiers English, comments Spanish OK**
  (`AGENTS.md`).
- **Nothing is deprecated.** `buildPicker` and `reloadSchedule` are deleted, not
  kept. Before closing:
  `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' .` — every hit must
  predate this change.
- Run `gotest`, never `go test`.

---

## 8. Acceptance criteria

### 8.1 Static

| # | Check | Expected |
|---|-------|----------|
| 1 | `gotest ./...` | green |
| 2 | `GOOS=js GOARCH=wasm go build ./web/` | compiles |
| 3 | `go list -deps ./web/ \| grep webtyp/svg/sprite` | **empty** (sprite must not reach WASM) |
| 4 | `grep -rn "workschedule\|agenda_schedule\|reloadSchedule" modules/agenda/` | **empty** |
| 5 | `grep -rn "DayOfWeek" modules/agenda/agenda.go` | only inside `fallbackWeek` (reading `ab.WorkCalendarWeekly`) and `toWCWeekly` (writing it) |
| 6 | `grep -rn '"Special"\|"hours"' config/lang.go` | **empty** |
| 7 | `grep -c "\.agenda" web/public/style.css` | **> 0** (see §4.1 fallback if zero) |
| 8 | `grep -n "workschedule.New" web/client.go` | one line |
| 9 | `grep -rn "TODO\|FIXME\|Deprecated" --include='*.go' .` | no hit introduced here |

### 8.2 Live — drive the running app with the `webtyp` MCP browser tools

Do **not** sign this off from reading code. `restart_development`, then
`browser_navigate` to `https://localhost:8080/#agenda` (**https**, not http —
the dev server speaks TLS and http returns a 400).

| # | Check | Expected |
|---|-------|----------|
| 10 | `browser_evaluate_js`: map `.scheduleeditor__day-name` → text | `["Domingo","Lunes","Martes","Miércoles","Jueves","Viernes","Sábado"]` |
| 11 | Click the row-index-2 toggle, re-read every row's `checked` | row 2 flips; rows 0 and 1 unchanged |
| 12 | `getComputedStyle('.scheduleeditor__exc-form').display` with no day picked | `none` |
| 13 | Pick `HOLIDAY`, read `.scheduleeditor__exc-hours` display | `none`; `SPECIAL_HOURS` shows it |
| 14 | The `SPECIAL_HOURS` radio label | **"Horario especial"** |
| 15 | `browser_audit_mobile` on `.agenda` | **0** tap targets under 44×44 |
| 16 | `document.querySelector('.agenda_schedule')` | `null` |
| 17 | Navigate `#workschedule` | renders the read-only legacy list |
| 18 | `getComputedStyle('.agenda').padding` | non-zero |
| 19 | `browser_emulate_device` iPhone 14 Pro, screenshot the type radios | no radio separated from its label by a line break |

---

## 9. Stages

| # | Stage | Files | Done when |
|---|-------|-------|-----------|
| 1 | New component API | `modules/agenda/agenda.go` | `fallbackWeek` indexes by day; `toWCWeekly(dayOfWeek, row)`; compiles against the new tag |
| 2 | `work_schedule` moves out | `modules/agenda/`, **new** `modules/staffpick/`, **new** `modules/workschedule/`, `web/client.go` | check 4 empty; `#workschedule` in the nav |
| 3 | The view gets chrome | `modules/agenda/agenda.go`, **new** `modules/agenda/css.go` | check 7 > 0; `<h1>` and section headings present |
| 4 | Dictionary | `config/lang.go` | check 6 empty; check 14 reads "Horario especial" |
| 5 | Tests, README, index | `modules/agenda/agenda_test.go`, **new** `modules/workschedule/workschedule_test.go`, `README.md`, `docs/PLAN.md` | §8.1 and §8.2 all pass |

## 10. What this deletes

`scheduleView.schedule`, `scheduleView.reloadSchedule`, `scheduleView.buildPicker`,
the `agenda_schedule` markup block, the `work_schedule` import in `agenda`, the
four inert hand-written class strings, `TestWorkSchedule_ViewsSchedule`, and the
`"Special"` / `"hours"` dictionary entries. Nothing is deprecated or kept behind
a flag.
