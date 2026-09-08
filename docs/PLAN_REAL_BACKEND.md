---
PLAN: "feat!(demo): real backend — routes/routes.go, HTTP transport and SSE, all in memory"
EXECUTOR: local
REVIEWER: none
---

> This plan is dispatched via the CodeJob workflow. See skill: agents-workflow.
>
> **Depends on a GATE**: `webtyp/events` must publish `events.Fanout` first
> ([events/docs/PLAN.md](https://github.com/webtyp/events/blob/main/docs/PLAN.md)).
> **Runs BEFORE** [PLAN_AGENDA_DOMAIN.md](PLAN_AGENDA_DOMAIN.md), which wires
> the agenda domain into the structure this plan creates. Orchestrator:
> [webtyp/docs/AGENDA_DOMAIN_MASTER_PLAN.md](https://github.com/webtyp/webtyp/blob/main/docs/AGENDA_DOMAIN_MASTER_PLAN.md).
>
> Read `AGENTS.md` in this repo root first.

# Plan — un backend real para la demo, íntegramente en memoria

## 1. What is wrong today

`app-demo` has **no backend at all**. Everything runs inside the WASM binary:

- `config.Env` builds an in-WASM `orm.DB` over `storage/mem`, mounts the
  veltylabs modules on a **`router/loopback` caller**, and seeds them — all in
  the browser.
- `web/` holds only `client.go`. There is no `server.go`, no `config/server.go`,
  no `routes/routes.go`. The dev daemon says so on every boot:
  `SERVER Internal mode: no routes/routes.go and no server main, checked
  .../app-demo/routes/routes.go`.
- Events use `events/mock.Broker` in-process, inside the browser.

So the demo proves the **views** work, but never exercises the transport a real
application uses: HTTP ops, a server-side registry, and SSE pushing domain
events to the client. `app-demo/AGENTS.md` calls this repo the canonical
"how you build with WebTyp" example; right now it is canonical for half the
stack.

It also diverges from the layout every other app uses (skill
**project-layout**): `config/` is missing `server.go` and `client.go`, `modules/`
has no `init.go`/`backend.go`/`view.go` split, and there is no `tests/`.

## 2. What this plan builds

A **real client/server split that needs no database**. The store stays
`storage/mem`, seeded at server start; stopping `webtyp dev` resets everything.
That is the explicit goal: the **complete flow** must be exercisable, not the
data preserved.

```
browser (WASM)                    server (Go)
──────────────                    ───────────
mcp.NewCaller ──── HTTP /mcp ───► router registry ──► veltylabs modules
                                                          │  orm.DB (storage/mem)
sse.SSEClient ◄─── GET /events ──── sse.Publisher ◄───────┘  via events.Fanout
```

## 3. Design decisions

| # | Decision | Why |
|---|---|---|
| **B1** | Store stays `storage/mem`, seeded at server start | The owner's requirement: no DB, data resets with the daemon, the flow is what matters. `storage/mem` is a real `orm` backend, not a test double — no `if dev` anywhere. |
| **B2** | `loopback.Caller` is **deleted** from the client | Keeping it beside the HTTP caller would leave two ways to reach an op, and the loopback path would silently keep working when the transport broke. |
| **B3** | One `events.Fanout` on the server | Modules take a single `Publisher` that reaches both the in-proc broker (module→module, e.g. `appointment_booking` subscribing to `business_calendar`) and `sse.Publisher` (server→browser). See the `events` gate plan. |
| **B4** | `Subscriber` is always the broker, never SSE | SSE is one-way server→browser and cannot deliver to an in-process handler. |
| **B5** | Adopt the canonical layout | `config/server.go`, `config/client.go`, `routes/routes.go`, `tests/`. **No `web/server.go`** — declaring `routes/routes.go` makes `webtyp.com/server` generate the main (§4). The demo teaches the layout; it must have the current one. |

## 4. Stage 1 — `routes/routes.go` (and NO `web/server.go`)

**Declaring `routes/routes.go` is the whole requirement.** `webtyp.com/server`
generates the server main for you. Verified in `server/main_decision.go`:

```go
usesGeneratedMain() = HasRoutes(rootDir) && !hasHandWrittenMain()
```

| Startup log | Condition |
|---|---|
| `External mode: routes/routes.go present, generated main from …` | **the target state** |
| `External mode: user-written server main at …` | a `web/server.go` exists — the escape hatch |
| `Internal mode: no routes/routes.go and no server main, checked …` | today's state |

The generated main lands in **`.build/server/main.go`** (`.build/` auto-added to
`.gitignore`), is marked `DO NOT EDIT`, and already does all of this:

```go
s := httpd.New(httpd.Config{Port: …, PublicDir: …, Gzip: true, Health: true,
    TLS: httpd.TLSConfig{DevTLS: …}})
routes.Register(s.Router())
s.Router().PublicAsset(httpd.CAPath, …)   // dev CA download
s.ListenAndServe()
```

**Do NOT write `web/server.go`.** `httpd.New`, the port, the public dir, gzip,
health, dev TLS and the listener are not this repo's to write. `mjosefa-cms`
has one only because it must open Postgres and gate schema sync before
listening; the demo has neither.

### 4.1 `Register` takes exactly ONE parameter

This is the constraint that decides the whole composition. `detectRegisterArgs`
(`server/maingen.go`) reads `Register`'s arity from the AST and renders `nil`
for **every parameter after the router**:

```go
func Register(r router.Router)                        // → routes.Register(s.Router())       ✅
func Register(r router.Router, m ...router.APIModule) // → routes.Register(s.Router(), nil)  ❌
```

The `mjosefa-cms` variadic shape would boot the demo **with no modules
mounted**. So `routes.Register` calls into `config/` itself:

```go
// Package routes is the single place app-demo declares what its server
// answers. It is the server's entry point: webtyp.com/server generates the
// main from its presence and calls Register with the running router.
//
// Register takes ONLY the router: the generated main renders nil for every
// parameter after it, so a dependency taken here would arrive nil. The
// composition lives in config/, which this file calls.
package routes

import (
	"webtyp.com/router"

	"webtyp.com/app-demo/config"
)

func Register(r router.Router) {
	for _, m := range config.ServerModules() {
		if m != nil {
			m.MountAPI(r)
		}
	}
}
```

`config.ServerModules()` is the composition root's export: it builds the store,
the broker, SSE and the domain modules **once** and returns them as
`[]router.APIModule`. `config/` composes; `routes/` mounts. The rule survives,
only the direction of the call changes.

## 5. Stage 2 — `config/server.go`

**New file, `//go:build !wasm`.** Model it on
[`mjosefa-cms/config/server.go`](https://github.com/veltylabs/mjosefa-cms/blob/main/config/server.go),
minus everything about Postgres, auth and RBAC — the demo has none.

**There is no `BuildServer`/`RunServer` here.** The generated main owns
`httpd.New` and `ListenAndServe` (§4). This file supplies **dependencies**, not
a listener:

```go
// ServerModules builds the demo's server-side runtime ONCE and returns what
// routes.Register mounts. Idempotent: repeated calls return the same modules,
// so the seed runs a single time.
func ServerModules() []router.APIModule
```

`ServerModules` does, in order:

1. `db := orm.New(mem.New())` — the in-memory store (B1).
2. `broker := &mock.Broker{}` — in-process delivery.
3. `sseSrv := sse.New(...)` and `ssePub := sse.Publisher{Server: sseSrv}`.
4. `pub := events.Fanout{broker, ssePub}` — the single `Publisher` every module
   receives (B3).
5. `ids := unixid.NewUnixID()`.
6. Build the veltylabs modules with `db`, `ids`, `pub` and — where the module
   needs it — `broker` as `Subscriber` (B4).
7. Seed, through the **real ops**, exactly as `config.Env` does today.
8. Return the modules as `[]router.APIModule` for `routes.Register` to mount.

It does **not** create an `httpd.Server`, choose a port, serve `PublicDir`, or
listen — all of that is in the generated main. Guard the construction with a
`sync.Once` so the seed cannot run twice.

`ServerConfig`/`ServerDeps` still exist for injecting doubles in `tests/` (a
fake clock, a recording publisher), but no field of theirs is a port or a
listener.

**`config/env.go` is deleted.** Its two jobs split: building the runtime moves
to `config/server.go` (server side), and the demo's staff list and helpers move
to `config/config.go` as neutral shared data. Nothing that touches `orm` may
remain in a neutral file — the client no longer has a database.

## 6. Stage 3 — `config/client.go` and the HTTP caller

**New file, neutral (no build tag)** so the view rail stays testable under
stdlib, per the canonical layout.

```go
type ClientConfig struct {
	Origin string // where the server lives; the browser's own origin
}

// BuildClient assembles the platformd chassis with every module's view, all
// reaching the server over HTTP. Returns the chassis; web/client.go only
// resolves the origin and calls this.
func BuildClient(cfg ClientConfig) *platformd.Platform
```

The caller changes from in-process to real:

```go
caller := mcp.NewCaller(mcp.NewClient(cfg.Origin, ""))
```

**Delete `router/loopback` from `go.mod` and from every import** (B2). Acceptance:
`grep -rn "loopback" .` → empty.

The auth token is `""` — the demo has no authentication, and that is stated in
`README.md` rather than faked.

## 7. Stage 4 — `web/` mains

**`web/client.go`** shrinks to a thin main (it currently holds the whole
composition — brand, identity, module list):

```go
//go:build wasm

func main() {
	p := config.BuildClient(config.ClientConfig{Origin: dom.Origin()})
	Append("body", p)
	select {}
}
```

Everything it holds today — `demoBrand`, `demoIdentity`, `hiddenModule`, the
`p.Modules` slice — moves into `config/client.go`. Check what `dom` exposes for
the current origin; if there is no helper, read `window.location.origin` through
the existing `dom` API rather than adding one.

**`web/server.go` is NOT created.** The generated main at `.build/server/main.go`
is the server entry point (§4). Creating one here would silently switch the
daemon to the `user-written server main` branch and put this repo on the hook
for `httpd.New`, the port, gzip, health and dev TLS — none of which it needs to
decide.

No `DATABASE_URL`, no `ENABLE_SCHEMA_SYNC`, no port flag: the demo has no
database, and the daemon already owns the port.

## 8. Stage 5 — SSE on the client

The browser subscribes to the stream and republishes into a client-side broker,
so views keep using the `events.Subscriber` contract they already know.

```go
// In config/client.go: the SSE stream is the client's Publisher-side; views
// subscribe on the local broker exactly as they did in-process.
clientBroker := &mock.Broker{}
sseClient := sse.NewClient(...)   // read sse/docs/USAGE.md for the exact API
// on message: decode into events.Event and clientBroker.Publish(e)
```

Read [`sse/docs/USAGE.md`](https://github.com/webtyp/sse/blob/main/docs/USAGE.md)
and `ClientConfig` before writing — the client half is TinyGo-compatible by
design and its API is documented there. `sse.Publisher` already encodes the
payload with `webtyp/json` and uses the topic as **both** the SSE event name and
the channel, so the client dispatches by topic with no extra convention.

`modules/reservation` already subscribes to `schedule.changed` through a broker
(`README.md` → "Sincronía por eventos"). **It must not change.** Swapping the
broker it receives from an in-process one to the SSE-fed one is the whole point:
the view code is untouched and the transport moved underneath it. That is the
proof this plan exists to produce.

## 9. Stage 6 — `tests/`

The canonical layout puts **every** test in `tests/`, package `tests`, importing
`config`. Today they sit beside the code (`modules/agenda/agenda_test.go`,
`config/env_test.go`). Move them and add the two shapes the split makes possible:

- `tests/setup_test.go` — `TestMain`, plus a `buildUnderTest()` helper that
  assembles a server the same way the generated main does — `httpd.New(...)`,
  `routes.Register(s.Router())`, `s.Handler()` — and wraps it in `httptest`.
  Tests own that assembly precisely because production's copy is generated and
  gitignored; do **not** add a `BuildServer` to `config/` just to share it, or
  the escape-hatch branch becomes tempting.
- `tests/<m>_contract_test.go` — **real caller ↔ httptest**: `mcp.NewCaller`
  against the real handler, proving an op works over the transport rather than
  over loopback.
- `tests/<m>_view_test.go` — the view with a mock caller, neutral.

This is what makes "the complete flow" testable without a browser.

## 10. Stage 7 — docs

- `README.md` — replace the "everything runs in WASM via loopback" description
  with the real topology of §2. State plainly: **no database, `storage/mem`,
  data resets when the daemon stops**, and **no authentication**.
- `AGENTS.md` — the module shape changed (`init.go`/`backend.go`/`view.go`); it
  must say so.
- [`docs/PLAN.md`](PLAN.md) — add this stage's row.

## 11. Constraints

- **No stdlib in WASM code**: `webtyp.com/fmt`. `config/server.go` and
  `routes/routes.go` are server-side and may use `log`.
- **`dom.Element` embedded by VALUE.**
- **Build tags by layer**: `config/server.go` and `config/css.go` are `!wasm`;
  `config/config.go`, `config/client.go` and `config/lang.go` are **neutral**;
  `web/client.go` is `wasm`. `routes/routes.go` carries **no** build tag —
  the generated main imports it under `!wasm` and nothing else does.
- **`config/client.go` must stay neutral** — no `orm`, no `storage`, no
  `loopback`. If it needs a build tag, the composition leaked a server concern
  into the client.
- **No `if dev`, no test-only routes.** `storage/mem` is a real backend.
- **Nothing deprecated.** `config/env.go` and `router/loopback` are **deleted**,
  not left beside the new path.
- `gotest`, never `go test`.

## 12. Acceptance criteria

### 12.1 Static

| # | Check | Expected |
|---|-------|----------|
| 1 | `gotest ./...` | green |
| 2 | `GOOS=js GOARCH=wasm go build ./web/` | compiles |
| 3 | `go build ./routes/ ./config/` | the server-side composition compiles |
| 4 | `go list -deps ./web/ \| grep webtyp/svg/sprite` | **empty** |
| 5 | `grep -rn "loopback" .` | **empty** — B2 |
| 6 | `ls config/env.go` | **no such file** |
| 7 | `ls routes/routes.go config/server.go config/client.go` | all present |
| 7b | `ls web/server.go` | **no such file** — the main is generated (§4) |
| 7c | `grep -c "func Register" routes/routes.go` and its arity | exactly one param, `router.Router` (§4.1) |
| 7d | `grep -n "^\.build/" .gitignore` | present — the daemon adds it |
| 8 | `grep -n "events.Fanout" config/server.go` | present — B3 |
| 9 | `grep -rn "orm\.\|storage/" config/client.go` | **empty** — the client has no database |
| 10 | `ls tests/setup_test.go` | present |

### 12.2 Live — drive the running app with the `webtyp` MCP browser tools

`restart_development`, then **https**://localhost:8080 (http returns 400).

| # | Check | Expected |
|---|-------|----------|
| 11 | `app_get_logs BUILD` | says `External mode: routes/routes.go present, generated main from …` — **not** `Internal mode`, and **not** `user-written server main` |
| 12 | `browser_get_network_logs` after loading `#reservation` | at least one request to the `/mcp` endpoint — ops travel over HTTP |
| 13 | `browser_get_network_logs` | an open `text/event-stream` connection |
| 14 | Every module still renders and its list loads | no regression from the transport swap |
| 15 | Editing an agenda in `#agenda` updates `#reservation`'s slots | `schedule.changed` crossed the process boundary over SSE |
| 16 | `browser_get_errors` | empty |
| 17 | Restart the daemon and reload | the seed is back to its initial state — B1, and it is the documented behaviour |

Check 15 is the one that proves the whole plan: the view code that already
subscribed to `schedule.changed` in-process now receives it over the wire, with
no change to the view.

## 13. Stages

| # | Stage | Files | Done when |
|---|-------|-------|-----------|
| 1 | Routes | **new** `routes/routes.go` (one param) | checks 7b, 7c, 11 |
| 2 | Server composition | **new** `config/server.go`, **delete** `config/env.go` | checks 3, 6, 7, 8 |
| 3 | Client composition | **new** `config/client.go` | checks 5, 9 |
| 4 | Thin client main | `web/client.go` only — **no `web/server.go`** | checks 2, 7b |
| 5 | SSE client | `config/client.go` | checks 13, 15 |
| 6 | Tests move | **new** `tests/`, delete the in-place ones | checks 1, 10 |
| 7 | Docs | `README.md`, `AGENTS.md`, `docs/PLAN.md` | topology, no-DB and no-auth stated |

## 14. What this plan does NOT do

- **No database.** `storage/mem` only; data resets with the daemon (B1).
- **No authentication.** The MCP client sends an empty token; stated in
  `README.md`, not faked.
- **No agenda-domain changes.** Blocks, marked days, holidays and conflicts are
  [PLAN_AGENDA_DOMAIN.md](PLAN_AGENDA_DOMAIN.md), which runs after this and
  wires into `config/server.go` rather than the deleted `config/env.go`.
