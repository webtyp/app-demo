---
PLAN: "chore(demo): drop every local replace, all deps to latest published versions"
EXECUTOR: jules
REVIEWER: none
STATUS: running
SESSION: 11756890902070943952
---

# PLAN — registro de ejecución para `app-demo`

> **PRÓXIMO DISPATCH.** Si te dijeron "execute the plan described in
> docs/PLAN.md", ejecutá **[PLAN_DROP_LOCAL_REPLACES.md](PLAN_DROP_LOCAL_REPLACES.md)**
> (fila "Deps" de la tabla). El resto de este archivo es el mapa/historial del
> repo — contexto, no la tarea — y **no se borra**.

> **ESTADO FINAL (Etapa I): feature completa.** El orquestador
> `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md` §7 tiene el estado de todos los
> repos. (Copia local: `/home/cesar/Dev/Project/webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md`.)
>
> Cada etapa se ejecutó en local: su `PLAN_*.md` rotó a `docs/LAST_PLAN_EXECUTED.md`
> (el contenido se sobrescribe en cada ejecución — git guarda el historial) y
> se publicó con `gopush` junto a su implementación. Este índice NO se borra:
> es el mapa del repo. Los `PLAN_*.md` de etapas ya cerradas se re-numneran/rotan
> como historial (git conserva cada versión).

| Orden | Etapa | Plan (contenido en LAST_PLAN_EXECUTED | Estado |
|-------|-------|---------------------------------------|--------|
| 1 | D — agenda E2E | [PLAN_WORK_SCHEDULE](LAST_PLAN_EXECUTED.md) | ✅ hecho (`v0.0.14`, config root) |
| 2 | F — reservation v2 | [PLAN_RESERVATION](LAST_PLAN_EXECUTED.md) | ✅ hecho (`v0.0.15`) |
| 3 | G — eventos + FreeSlots | [PLAN_EVENTS.md](PLAN_EVENTS.md) | ✅ hecho (`v0.0.18`) |
| 4 | H — item_catalog real | [PLAN_CATALOG](LAST_PLAN_EXECUTED.md) | ✅ hecho (`v0.0.16`) |
| 5 | I — docs + limpieza | [PLAN_DOCS_CLEANUP.md](PLAN_DOCS_CLEANUP.md) | 🟡 en curso (estado en el master) |
| 6 | Fixes de la vista `#agenda` | [PLAN_AGENDA_VIEW.md](PLAN_AGENDA_VIEW.md) | ✅ hecho completo (`6490daf` + residual §10) |
| 7 | Backend real: `routes/`, HTTP y SSE, todo en memoria | [PLAN_REAL_BACKEND.md](PLAN_REAL_BACKEND.md) | ⬜ pendiente — bloqueado por `events.Fanout` |
| 8 | Dominio de agenda (bloques, días marcados, feriados, conflictos) | [PLAN_AGENDA_DOMAIN.md](PLAN_AGENDA_DOMAIN.md) | ⬜ pendiente — bloqueado por 4 gates + etapa 7 |
| Deps | Quitar los 4 `replace` locales; todo dep a su última versión publicada | [PLAN_DROP_LOCAL_REPLACES.md](PLAN_DROP_LOCAL_REPLACES.md) | ✅ hecho (sin replaces locales, todo dep a su versión publicada) |

> **Etapa 6 (2026-09-08) — ✅ EJECUTADA COMPLETA.** Manejando `#agenda` en la
> demo corriendo aparecieron una corrupción de datos al guardar (activar el
> martes escribía el domingo), un formulario de excepciones permanentemente
> abierto y un módulo sin **ninguna** regla CSS. Se resolvió en tres repos:
> `widget v0.6.25` → `components v0.6.22` → `app-demo 6490daf`. Verificado en
> vivo con el MCP: nombres de día correctos, activar Martes activa Martes,
> formulario oculto sin día elegido, "Horario especial", cabeceras de columna,
> 0 tap targets reales bajo 44×44. Orquestador con la tabla completa:
> [`webtyp/docs/AGENDA_VIEW_FIXES_MASTER_PLAN.md`](https://github.com/webtyp/webtyp/blob/main/docs/AGENDA_VIEW_FIXES_MASTER_PLAN.md).
>
> Los 3 residuales de §10 (`style.Center()` en `PartDay`/`PartHeader`, "sin
> colación" mostrado como 06:00, `modules/workschedule` sin chrome) se
> corrigieron directo en fuente el mismo día. `gotest` verde en `components`, `widget` y `app-demo`.

> **Etapa 7 (2026-09-08) — ⬜ PENDIENTE.** Hoy la demo corre **entera en WASM**
> con `router/loopback` y sin backend: el daemon lo dice en cada arranque
> (`Internal mode: no routes/routes.go and no server main`). Prueba las vistas
> pero nunca el transporte que usa una app real. Esta etapa mete
> `routes/routes.go`, `config/server.go`, `web/server.go`, el caller HTTP
> (`mcp.NewCaller`) y **`webtyp/sse`** para los eventos, y adopta el layout
> canónico (skill `project-layout`). **Sin base de datos**: `storage/mem`
> sembrado al arrancar; parar el daemon resetea todo. `router/loopback` se borra.

> **Etapa 8 (2026-09-08) — ⬜ PENDIENTE.** La Etapa 6 arregló la *vista* sobre el
> modelo viejo; esta cambia el **modelo**. Bloques por día (la colación pasa a
> ser el hueco entre bloques, no un campo), días marcados para profesionales
> irregulares, feriados y horario del local como **dato editable** en el módulo
> nuevo `veltylabs/business_calendar`, y detección de reservas en conflicto al
> reducir una agenda. Sin vigencia y sin RRULE — ambas descartadas por decisión
> del dueño. Los 24 casos de uso y el grafo de gates están en
> [`webtyp/docs/AGENDA_DOMAIN_MASTER_PLAN.md`](https://github.com/webtyp/webtyp/blob/main/docs/AGENDA_DOMAIN_MASTER_PLAN.md).
| 6 | J — agenda view fixes | [PLAN_AGENDA_VIEW.md](PLAN_AGENDA_VIEW.md) | 🟡 en curso (orquestador `AGENDA_VIEW_FIXES_MASTER_PLAN.md`) |

## Decisiones de la ejecución (registradas en el master §3/§7)

- **`demoenv` → `config`**: el paquete `demoenv/` nuevo se consolidó en
  `config/` (la raíz de composición ya existía con `lang.go`/`css.go`); el
  `Env` vive en [`config/env.go`](../config/env.go).
- **O1 resuelta**: `events/mock.Broker` pasa a `map`-free (plan a `events`).
- **Ids del arnés**: `dom` será la única fuente de ids (plan a `dom`).

## Reglas (resumen — completo en el master §6)

- Textos de UI en **español**; código e identificadores en **inglés**;
  comentarios en español OK.
- `web/client.go` es la composición: agregar un módulo = una línea ahí.
- Cada módulo demo = wrapper fino sobre el módulo real de `veltylabs/modules/*`:
  `loopback.Caller` (via `config.Env`) + `orm.DB` sembrado + el client/NewView
  del módulo real. Los módulos locales (devices/medicalhistory) conservan su
  patrón de `model.go`/`store.go` (no tienen módulo `veltylabs/`).
- Sin `if dev`, sin rutas solo-test. Store `storage/mem` real con seed realista.
- `gotest` (nunca `go test`). Antes de publicar:
  `GOOS=js GOARCH=wasm go build ./web/` OK y
  `go list -deps ./web/ | grep webtyp/svg/sprite` vacío.
- Actualizar `README.md` (índice de módulos) y `AGENTS.md` (si cambia la forma
  del módulo) cuando se agrega/quita un módulo.