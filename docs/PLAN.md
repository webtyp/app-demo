---
PLAN: "fix(agenda): Spanish chrome for the pattern editor + buttons on the shared recipe"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 7195889616396498878
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> Its gates are shipped: **`webtyp.com/widget v0.6.26`** and
> **`webtyp.com/components v0.6.24`**. Orchestrator:
> [webtyp/docs/BUTTON_SYSTEM_MASTER_PLAN.md](https://github.com/webtyp/webtyp/blob/main/docs/BUTTON_SYSTEM_MASTER_PLAN.md).

# Plan — close the `#agenda` view: Spanish chrome, then the button bump

## 1. Context (zero assumptions)

`app-demo` is the WebTyp demo application. Its `#agenda` view wraps
`components/scheduleeditor` through [`modules/agenda/`](../modules/agenda/),
seeded by [`config/env.go`](../config/env.go) against the real
`github.com/veltylabs/appointment_booking` module.

Two things went wrong in that view and this plan closes both.

### 1.1 The chrome rendered in English (already fixed in the working tree)

`components/scheduleeditor` renders **all** of its chrome through
`webtyp.com/fmt/lang` and registers **no dictionary of its own** — that is the
framework rule: a library never hardcodes a human language, and the consuming
app supplies the words. This app's one dictionary is
[`config/lang.go`](../config/lang.go), and `AGENTS.md` §Language states the
obligation:

> Adding a module that pulls in a NEW framework component with its own
> translatable chrome? Add its English→Spanish words to `config/lang.go`'s
> dictionary — do not invent a second registration site.

When the editor migrated from the weekly grid to the block/pattern model, the
dictionary was not updated. It still carried the dead grid keys (`Day`,
`Work start`, `Work end`, `Break start`, `Break end`, `No break` — no code
anywhere translates those any more) and lacked every key the pattern editor
introduced. The running demo showed *"Weekly pattern / Sun Mon Tue Wed Thu Fri
Sat / Add row / Remove row / Marked days / Hours for marked days"* in English
inside an otherwise Spanish app.

**This is already done in the working tree** and verified live. The dead keys
are gone and these are registered:

| Key | ES |
|---|---|
| `Weekly pattern` | Patrón semanal |
| `Add row` | Agregar fila |
| `Remove row` | Quitar fila |
| `Marked days` | Días marcados |
| `Hours for marked days` | Horario de días marcados |
| `Sun`…`Sat` | Dom, Lun, Mar, Mié, Jue, Vie, Sáb |

Also already done: `modules/agenda/agenda.go` no longer renders its own
`<h2>Plantilla semanal</h2>`, which sat directly above the component's own
"Patrón semanal" label and said the same thing twice. Its now-orphaned
`PartSection` constant, `clsSection` var and CSS rule were removed with it.

**Your job for §1.1 is to verify, not to redo it.** If the checks in §5 pass,
change nothing here.

### 1.2 The buttons (fixed upstream; verify here)

`.scheduleeditor__row-add` ("Agregar fila") rendered as a **799.16 px** gradient
bar across the panel, and `.scheduleeditor__row-remove` overflowed its own box.
Neither was fixable from this repo: the cause was in `widget/style`, which had
no way to stop a button stretching across a `Stack`'s cross axis. Fixed upstream
by `widget v0.6.26`'s `style.Button`, consumed by `components v0.6.24`.

Measured after the bump, same viewport (888×588): `.scheduleeditor__row-add` is
now **105.72 px** wide with `padding-inline: 12px`. This repo only confirms it
stays that way.

## 2. Stage 1 — take the published tags

From the repo root:

```
go get -u webtyp.com/...@latest
go mod tidy
```

The `go.mod` in the tree may already carry both tags (the publish of
`components` cascades a deps-only bump into this repo). If `go get` changes
nothing, that is correct — verify, do not force a change. Confirm, and state
the resolved versions in the PR body:

- `webtyp.com/components` — **`v0.6.24`** or higher
- `webtyp.com/widget` — **`v0.6.26`** or higher

Do **not** add any `replace` directive. This repo resolves every dependency
from its published version and a recent plan removed the last local replace;
re-introducing one is a regression.

`webtyp.com/image` will still show as upgradable in `go list -m -u all`
(`v0.1.3` → `v0.1.6`). **Leave it.** No package in this module imports it — it
enters the graph only through `components`/`layout`'s own `go.mod`, and
`go mod tidy` removes any pin you add. It is theirs to raise, not this repo's.

## 3. Stage 2 — fix whatever the bump breaks

The bump changes no API this app calls: Phase B touched only `css.go` files in
`components`. So `go build ./...` should pass untouched.

If it does not:

- **Mechanical shift** (a symbol moved, an argument added) → apply it across
  every call site, matching the surrounding style.
- **A change that needs a design decision** → **STOP.** Do not invent
  behaviour, do not stub it. Leave the dependency at the highest version that
  builds, state the blocker explicitly in the PR body, and finish the rest.

Do **not** rewrite `modules/agenda/`, `modules/workschedule/` or `config/`
beyond what the compiler forces. `docs/PLAN_REAL_BACKEND.md` and
`docs/PLAN_AGENDA_DOMAIN.md` own those rewrites.

## 4. Stage 3 — no dictionary key may be left behind

Re-read `components/scheduleeditor`'s `README.md` section **"Translation keys"**
in the newly resolved module version, and make sure every key it lists exists in
[`config/lang.go`](../config/lang.go). If Phase B introduced a new one, add it
with a Spanish value; if it retired one, delete that entry.

This is the check whose absence caused §1.1. A key the component translates and
the dictionary does not carry renders in English with no error anywhere — a
silent failure, which is why it gets its own stage rather than a footnote.

## 5. Stage 4 — verify

```
gotest ./...
go vet ./... && gofmt -l .
GOOS=js GOARCH=wasm go build -o /tmp/demo.wasm ./web/
GOOS=js GOARCH=wasm go list -deps ./web/ | grep webtyp.com/svg/sprite
```

`gotest` green; `gofmt` clean; the WASM build compiles; the sprite grep is
empty. Note the `GOOS=js GOARCH=wasm` prefix on the sprite guard: without it
`go list` cannot build `./web/` at all and the guard passes vacuously.

Then the dictionary and heading checks, which must all hold:

| # | Check | Expected |
|---|-------|----------|
| 1 | `grep -c 'Work start\|Break start\|No break' config/lang.go` | **0** — the dead grid keys are gone |
| 2 | `grep -c 'Weekly pattern\|Add row\|Remove row\|Marked days\|Hours for marked days' config/lang.go` | **5** |
| 3 | `grep -c '{EN: "Sun"\|{EN: "Mon"\|{EN: "Tue"\|{EN: "Wed"\|{EN: "Thu"\|{EN: "Fri"\|{EN: "Sat"' config/lang.go` | **7** |
| 4 | `grep -n 'Plantilla semanal' modules/agenda/agenda.go` | empty — the duplicate heading is gone |
| 5 | `grep -rn 'PartSection\|clsSection' modules/agenda/` | empty — no orphan constant |
| 6 | `grep -c '^replace ' go.mod` | **0** |
| 7 | `grep -n 'components v0.6.2\|widget v0.6.2' go.mod` | `components ≥ v0.6.24`, `widget ≥ v0.6.26` |

## 6. Constraints

- `gotest`, never `go test`. Never run `gopush` or `codejob`.
- UI strings in Spanish; code and identifiers in English; Spanish comments are
  fine in this repo.
- `config/lang.go` is the **only** translation registration site. Never call
  `lang.RegisterWords` anywhere else, and never hardcode a Spanish string into a
  library.
- No `replace`, no `TODO`, no `FIXME`, nothing deprecated left behind.
- Do not touch `docs/PLAN_REAL_BACKEND.md`, `docs/PLAN_AGENDA_DOMAIN.md` or the
  work they describe.

## 7. Stages table

| # | Stage | Files | Done when |
|---|-------|-------|-----------|
| 1 | Take the tags | `go.mod`, `go.sum` | components at the Phase B tag, widget ≥ `v0.6.26`, tidy clean |
| 2 | Fix fallout | wherever the compiler points | `go build ./... && go vet ./...` green |
| 3 | Dictionary parity | `config/lang.go` | every key in the component's README is registered |
| 4 | Verify | — | §5's four commands and six checks all pass |
