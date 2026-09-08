---
PLAN: "chore(demo): drop every local replace, all deps to latest published versions"
EXECUTOR: jules
REVIEWER: none
---

> Dispatched via the CodeJob workflow. See skill: **agents-workflow**.
> Queued from [PLAN.md](PLAN.md) (row "Deps"). Read `AGENTS.md` in this repo
> root first.
>
> **Scope is `go.mod` / `go.sum` hygiene only.** This plan does NOT wire the
> real backend or the agenda domain — those stay in
> [PLAN_REAL_BACKEND.md](PLAN_REAL_BACKEND.md) and
> [PLAN_AGENDA_DOMAIN.md](PLAN_AGENDA_DOMAIN.md). The one job here: the module
> graph resolves entirely from published versions, with nothing served from a
> sibling working tree.

# Plan — remove the local `replace` directives, pin everything to latest

## 1. Why

`app-demo/go.mod` carries four `replace` lines that point at sibling working
trees:

```
replace webtyp.com/icons                        => ../icons
replace github.com/veltylabs/appointment_booking => ../../veltylabs/modules/appointment_booking
replace github.com/veltylabs/item_catalog        => ../../veltylabs/modules/item_catalog
replace github.com/veltylabs/work_schedule       => ../../veltylabs/modules/work_schedule
```

Every one of those targets is now **published and tagged**, and each sibling
working tree is **clean and in sync with `origin/main` at that tag** (verified
2026-09-08):

| Module | Replace target HEAD | Published tag to require |
|---|---|---|
| `webtyp.com/icons` | `../icons` | **`v0.0.7`** |
| `github.com/veltylabs/appointment_booking` | `../../veltylabs/modules/appointment_booking` | **`v0.1.5`** (`feat!`: blocks replace the single window) |
| `github.com/veltylabs/item_catalog` | `../../veltylabs/modules/item_catalog` | **`v0.3.4`** |
| `github.com/veltylabs/work_schedule` | `../../veltylabs/modules/work_schedule` | **`v0.1.4`** |

The replaces are therefore dead weight: they resolve to the same code the
version would. Keeping them hides real version drift and blocks a clean
`go mod tidy`. Remove all four; require the published versions instead.

The three `github.com/veltylabs/*` modules are currently pinned to the
placeholder `v0.0.0` in `require` — that only ever resolved through the
replace. Those `v0.0.0` lines must become the real tags above.

## 2. Latest published versions for the `webtyp.com/*` dependencies

Bring every `webtyp.com/*` dependency (direct and indirect) to its latest
published version. Known deltas at plan time — confirm with
`go list -m -u all` and take whatever is newest, do not hard-code these if a
higher one exists:

| Module | Current | Latest |
|---|---|---|
| `webtyp.com/components` | `v0.6.22` | `v0.6.23` |
| `webtyp.com/events` | `v0.0.4` | `v0.0.5` — brings `events.Fanout` |
| `webtyp.com/html` | `v0.0.21` | `v0.0.23` |
| `webtyp.com/layout` | `v0.2.26` | `v0.2.31` |
| `webtyp.com/time` | `v0.5.5` | `v0.5.6` |
| `webtyp.com/view` | `v0.5.2` | `v0.5.7` |

All other `webtyp.com/*` modules are already at their latest and must not
regress.

## 3. Stages

### Stage 1 — strip the replaces

**File: `go.mod`.**

- Delete all four `replace` lines **and** the three comment blocks that
  introduce them (`// Local replaces for unreleased work: …`, `// Los módulos de
  veltylabs …`, `// item_catalog: …`, `// work_schedule: …`). No `replace`
  directive and no orphan comment about replaces may remain.
- In `require`, change the three `github.com/veltylabs/*` entries from `v0.0.0`
  to the tags in §1 (`v0.1.5`, `v0.3.4`, `v0.1.4`).

### Stage 2 — bring the rest to latest

Run, from the repo root:

```
go get -u webtyp.com/...@latest
go get github.com/veltylabs/appointment_booking@v0.1.5 \
       github.com/veltylabs/item_catalog@v0.3.4 \
       github.com/veltylabs/work_schedule@v0.1.4
go mod tidy
```

If a higher `github.com/veltylabs/*` tag exists at execution time, take it and
note the bump in the PR body.

### Stage 3 — fix the fallout

Removing the replaces changes no code, but the version bumps in §2 can. Build
and vet the whole module and the WASM entrypoint:

```
go build ./...
go vet ./...
GOOS=js GOARCH=wasm go build ./web/
```

For every compile or vet error:

- **Mechanical rename / signature shift** (a symbol moved, an argument added,
  a constant renamed) → apply the migration across every call site. Match the
  surrounding code's style; `app-demo` code is held to a readability bar
  (`AGENTS.md`).
- **A breaking change that needs a domain or design decision** (behaviour
  differs, a capability was removed with no drop-in replacement) → **STOP**.
  Do not invent behaviour or stub it. Leave that dependency at the highest
  version that still builds, list the blocker explicitly in the PR body, and
  finish the rest of the plan. A follow-up plan resolves it.
- Do **not** rewrite `config/`, `modules/agenda/` or `modules/workschedule/`
  beyond what the compiler forces. The real-backend and agenda-domain rewrites
  are other plans.

### Stage 4 — verify

```
gotest ./...
```

Green, including `conformance`-style tests. Then the repo's existing publish
guard:

```
GOOS=js GOARCH=wasm go build ./web/                       # compiles
go list -deps ./web/ | grep webtyp.com/svg/sprite         # empty
```

## 4. Acceptance criteria

| # | Check | Expected |
|---|-------|----------|
| 1 | `grep -c '^replace ' go.mod` | **0** |
| 2 | `grep -n 'v0.0.0' go.mod` | **empty** — no placeholder versions |
| 3 | `grep -n '\.\./\.\./veltylabs\|=> \.\./' go.mod` | **empty** |
| 4 | `go mod tidy` then `git diff --exit-code go.mod go.sum` | clean after the tidy is committed |
| 5 | `go list -m all \| grep -E 'veltylabs/(appointment_booking\|item_catalog\|work_schedule)'` | `v0.1.5`, `v0.3.4`, `v0.1.4` (or higher) |
| 6 | `go build ./... && go vet ./...` | green |
| 7 | `GOOS=js GOARCH=wasm go build ./web/` | compiles |
| 8 | `go list -deps ./web/ \| grep webtyp.com/svg/sprite` | empty |
| 9 | `gotest ./...` | green |
| 10 | `grep -rn 'replace' docs/PLAN.md` | only the historical Etapa-6 note; no open "pendiente quitar" item left for a replace that this plan removed |

## 5. Docs to update in the same PR

- **`docs/PLAN.md`** — mark the "Deps" row done; in the Etapa 6 note, strike the
  "pendiente quitar esas … líneas … cuando esos repos publiquen la corrección"
  clause (the repos published; the lines are gone).
- **`README.md`** — if it describes the module graph as "served from local
  working trees via `replace`", correct it to "every dependency resolves from
  its published version".
- **`AGENTS.md`** — only if it references the `replace` pattern as current
  practice.

## 6. Constraints

- `gotest`, never `go test`. No `gopush`/`codejob` from inside the plan.
- UI strings in Spanish; code and identifiers in English; comments may be
  Spanish.
- No new `replace`, no `TODO`/`FIXME`, nothing deprecated left behind.
- Touch source only where the compiler or `go vet` forces it — this is a
  dependency-hygiene plan, not a refactor.

## 7. Stages table

| # | Stage | Files | Done when |
|---|-------|-------|-----------|
| 1 | Strip replaces | `go.mod` | 0 `replace`, veltylabs deps at real tags |
| 2 | Deps to latest | `go.mod`, `go.sum` | `go list -m -u all` shows nothing newer for `webtyp.com/*` |
| 3 | Fix fallout | wherever the compiler points | `go build ./... && go vet ./...` green |
| 4 | Verify | — | `gotest ./...` + WASM build + sprite guard green |
| 5 | Docs | `docs/PLAN.md`, `README.md`, `AGENTS.md` | replace pattern no longer described as current |
