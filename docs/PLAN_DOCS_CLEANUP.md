# PLAN_DOCS_CLEANUP — verificación de docs multi-repo (Etapa I del `DEMO_AGENDA_MASTER_PLAN`)

Orquestador: `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md` §7 fila I.
Índice del repo: [PLAN.md](PLAN.md).

**Depende de:** todas las etapas anteriores publicadas (A, B, C, C2, D, E, F, G, H).

## Objetivo

Cerrar la feature: verificar que la documentación permanente de cada repo tocado
coincide con lo implementado, y limpiar marcadores temporales. **Solo `.md`** —
si algo del código está mal, es un `PLAN.md` nuevo en ese repo, no un fix aquí.

## Checklist por repo

### `webtyp/components`
- [ ] `docs/CATALOG.md` tiene la entrada de `scheduleeditor` con su API real.
- [ ] `README.md` lo indexa.
- [ ] `docs/ARCHITECTURE.md` (si lista componentes) lo incluye.
- [ ] `docs/PLAN.md` → renombrado a `LAST_PLAN_EXECUTED.md` al ejecutar (flujo
      local); confirmar que no quedó un `PLAN.md` huérfano.

### `webtyp/router`
- [ ] `docs/ARCHITECTURE.md` o `INTROSPECTION.md` describe `loopback` como el
      `Caller` in-proc de referencia frente a `mcp.NewCaller`.
- [ ] `README.md` indexa el sub-paquete.

### `webtyp/layout`
- [ ] `crudview` README / `docs/ARCHITECTURE.md` documentan `Config.Context` y la
      decisión "no auto-cableo al filtro".
- [ ] `docs/DICTIONARY.md` — si `scheduleeditor` aportó claves de traducción que
      pasan por `fmt/lang`, listarlas aquí (es el índice de chrome traducible).

### `veltylabs/modules/appointment_booking`
- [ ] `docs/ARCHITECTURE.md §7` reescrito: ya NO dice "no second view for
      calendar configuration"; describe `list_weekly_calendar`, `list_exceptions`,
      `ScheduleClient`.
- [ ] `docs/ARCHITECTURE.md §6` (tabla de eventos) incluye
      `appointment.schedule.changed`.
- [ ] `README.md`: lista de ops 11 → 13; evento nuevo documentado.
- [ ] Sin referencias vivas a `docs/PLAN_*.md` desde docs permanentes.

### `veltylabs/modules/work_schedule`
- [ ] `README.md` menciona `NewView` (si C2 se hizo).
- [ ] `docs/ARCHITECTURE.md` nota: vista solo-lectura por diseño; la edición de
      agenda vive en `appointment_booking`.

### `webtyp/app-demo`
- [ ] `README.md` — índice de módulos al día: `agenda` ("Agenda", editor sobre
      `appointment_booking`), `item_catalog` ("Catálogo"/"Especialidades"),
      `reservation` v2. Sección
      "demoenv" (loopback + módulos reales + `replace`). Sección "Sincronía por
      eventos" (broker in-proc → `sse` en producción). Brecha "sin Directory"
      documentada.
- [ ] `AGENTS.md` — si la forma del módulo cambió (wrapper sobre módulo real +
      `demoenv` en vez de `model.go`/`store.go` locales), actualizar "What a
      module looks like". Si `devices`/`medicalhistory` siguen con el patrón
      viejo, documentar que conviven dos patrones y por qué (los reales tienen
      módulo `veltylabs/` publicado; `devices`/`medicalhistory` no).
- [ ] `docs/PLAN.md` (índice) actualizado con el estado final; los `PLAN_*.md`
      de cada etapa → su contenido rotado a `LAST_PLAN_EXECUTED.md` en el último
      `gopush`, o dejados como registro (decidir: para un multi-etapa, dejar los
      `PLAN_*.md` como historial es aceptable — anotarlo en `docs/PLAN.md`).

### `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md`
- [ ] §7 — todas las filas en 🟢 con el tag publicado de cada repo.
- [ ] §3 O1/O2 — O1 resuelta (decir cómo). O2 sigue abierta (anotar que no se
      abordó y por qué).
- [ ] Nota final: "feature completa — este doc queda como registro histórico".

## Marcadores temporales a barrer

`grep -rn "STATUS (remove this note" <cada repo>/docs` — cada nota self-deleting
que un plan haya dejado debe estar quitada (su remoción era tarea del propio
plan; esta etapa solo verifica).

## Criterios de aceptación

- Los checklists de arriba, todos marcados.
- `grep -rn "docs/PLAN" <repo>/README.md <repo>/docs/ARCHITECTURE.md` en cada
  repo → sin referencias (los docs permanentes nunca citan un `PLAN.md`).
- `gotest ./...` verde en todos los repos (regresión final).
