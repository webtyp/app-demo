---
PLAN: "feat(demo): agenda editor + reserva v2 over real veltylabs modules"
EXECUTOR: local
REVIEWER: none
---

# PLAN — cola de ejecución para `app-demo`

> Si te dijeron "ejecuta el plan de `docs/PLAN.md`", este repo tiene **varios
> planes por etapa** con dependencias sobre OTROS repos. El orden y el estado
> real están en el orquestador:
> `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md` §7.
> (Copia local: `/home/cesar/Dev/Project/webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md`.)
>
> No ejecutar estos planes en secuencia ciega: cada uno depende de que ciertas
> etapas de `components` / `router` / `layout` / `appointment_booking` estén
> **publicadas** (tag nuevo) primero. Ver la columna "Depende de" del master.

| Orden | Plan | Etapa master | Depende de (publicado) |
|-------|------|--------------|------------------------|
| 1 | [PLAN_WORK_SCHEDULE.md](PLAN_WORK_SCHEDULE.md) | D | A, B, C, E (+ C2 opcional) |
| 2 | [PLAN_RESERVATION.md](PLAN_RESERVATION.md) | F | B, C, E |
| 3 | [PLAN_EVENTS.md](PLAN_EVENTS.md) | G | D, F (+ decisión O1 del master) |
| 4 | [PLAN_CATALOG.md](PLAN_CATALOG.md) | H | B, E |
| 5 | [PLAN_DOCS_CLEANUP.md](PLAN_DOCS_CLEANUP.md) | I | G, H |

Cada plan es autocontenido. Al ejecutarlo en local se renombra su contenido a
`docs/LAST_PLAN_EXECUTED.md` (sobrescribiendo el anterior — git guarda el
historial) y se publica con `gopush 'msg'` junto a su implementación.
`docs/PLAN.md` (este índice) NO se borra: es el mapa del repo.

## Reglas que aplican a todos (resumen — completo en el master §6)

- Textos de UI en **español**; código e identificadores en **inglés**;
  comentarios en español OK.
- `web/client.go` es la composición: agregar un módulo = una línea ahí.
- Cada módulo demo = wrapper fino sobre el módulo real de `veltylabs/modules/*`
  (patrón `mjosefa-cms/modules/item_catalog/`): `loopback.Caller` + `orm.DB`
  sembrado + `<mod>.NewView(caller, …)`.
- Sin `if dev`, sin rutas solo-test. Store `storage/mem` real con seed realista
  (nombres inventados).
- `gotest` (nunca `go test`). Antes de publicar:
  `GOOS=js GOARCH=wasm go build ./web/` OK y
  `go list -deps ./web/ | grep webtyp/svg/sprite` vacío.
- Actualizar `README.md` (índice de módulos) y `AGENTS.md` (si cambia la forma
  del módulo) cuando se agrega/quita un módulo.
