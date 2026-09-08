---
PLAN: "feat(demo): agenda editor + reserva v2 over real veltylabs modules"
EXECUTOR: local
REVIEWER: none
---

# PLAN — registro de ejecución para `app-demo`

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