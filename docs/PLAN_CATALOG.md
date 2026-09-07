# PLAN_CATALOG — módulo demo `item_catalog` real (Etapa H del `DEMO_AGENDA_MASTER_PLAN`)

Orquestador: `webtyp/docs/DEMO_AGENDA_MASTER_PLAN.md` §7 fila H.
Índice del repo: [PLAN.md](PLAN.md).

**Depende de (publicado):** `router/loopback` (B), `crudview.Config.Context` (E)
— este último solo si el catálogo usa el slot para "Área/Especialidad"; si no,
solo B.

## Objetivo

Un módulo demo `modules/item_catalog/` que monta el módulo **real**
`github.com/veltylabs/item_catalog` (`NewView` + `NewSpecialtyView` ya existen)
con datos sembrados. Es la prueba de que el layout `crudview` **se reutiliza sin
bifurcar** entre `item_catalog`, `work_schedule` y `appointment_booking`
(objetivo del usuario: "no un layout por módulo").

## 1. `go.mod`

```
require github.com/veltylabs/item_catalog v0.0.0
replace github.com/veltylabs/item_catalog => ../../veltylabs/modules/item_catalog
```

## 2. `demoenv`

Montar `item_catalog` en el `loopback` junto a los demás. `itemcatalog.New`
devuelve `(*Module, error)` (`if err != nil { panic(err) }`), y `*Module`
implementa `router.OperationModule` (verificado: `var _ router.OperationModule`
en `item_catalog/mcp.go`).

```go
ic, err := itemcatalog.New(db, itemcatalog.Deps{IDs: ids, Publisher: broker})
// panic(err) si err != nil
caller := loopback.New(ab, ws, ic)
```

Seed vía las ops del catálogo (`OpUpsertItem`, `OpUpsertSpecialty`):
- **Especialidades: usar las que el módulo real ya define** —
  `itemcatalog.CanonicalSpecialties` (`Medicina General` `md`, `Dental` `do`,
  `Traumatología` `tr`, `Podología` `po`, `Laboratorio` `la`, `Ecografía` `ec`,
  `Radiología` `ra`, …). Sembrar un subconjunto (5–6) vía `OpUpsertSpecialty`, o
  llamar `MigrateSpecialtiesAndItems("demo")` si es más directo. **No inventar
  slugs/prefijos nuevos** — Etapas D y F deben referenciar estos.
- Ítems de catálogo (servicios): 4–5 con `sku` cuyo prefijo (`sku[:2]`) case con
  una especialidad sembrada — p.ej. `md-consulta` "Consulta médica",
  `ec-abdominal` "Ecografía abdominal", `ra-torax` "Radiografía de tórax",
  `tr-control` "Control traumatología". Tipo `S` (service). Nombres inventados.
- `demoenv` expone `Specialties() []SpecialtyOption` leyendo del catálogo real
  (`OpListSpecialties`) — **fuente única**. Si D/F habían puesto una constante
  local de especialidades, esta etapa la reemplaza por esa lectura.

## 3. `modules/item_catalog/`

```
modules/item_catalog/
  item_catalog.go   # dos structs UIModule (Catalog, Specialties) + New* + View()
  svg.go            # //go:build !wasm
```

**Dos `platformd.UIModule` separados** ("Catálogo" y "Especialidades") — es lo
que hace `mjosefa-cms` de facto y evita inventar sub-navegación en `platformd`:

```go
// item_catalog.go
type catalogMod    struct{ p *platformd.Platform; env *demoenv.Env }
type specialtyMod  struct{ p *platformd.Platform; env *demoenv.Env }
func NewCatalog(p *platformd.Platform, env *demoenv.Env) *catalogMod
func NewSpecialties(p *platformd.Platform, env *demoenv.Env) *specialtyMod

func (m *catalogMod) View() Component {
    v, _ := crudview.New(crudview.Config{
        ParentID:  "catalog_item",
        Presenter: itemcatalog.NewView(m.env.Caller()),  // ← sin Filter/List/Context custom
        IDs:       m.env.IDs(),
    })
    return v
}
// specialtyMod.View() igual con itemcatalog.NewSpecialtyView(...) y ParentID "specialty"
```

La demostración es justamente que ambos usan `crudview.New(Config{})` **sin
config custom** — el mismo layout que `work_schedule` NO usa (porque su
contenido no es una lista) y que `reservation` sí extiende (Context + Filter
calendario). Ese contraste es la lección del §1 del usuario.

## 4. `web/client.go`

```go
itemcatalog.NewCatalog(p, env),      // "Catálogo"
itemcatalog.NewSpecialties(p, env),  // "Especialidades"
```

## 5. `config/lang.go`

`item_catalog` real puede traer chrome traducible (radios de tipo
Service/Product del `itemType()` widget). Agregar EN→ES si aparece en el render.

## Tests (`gotest`)

- `TestCatalog_ListsSeededItems` — `View()` con el catálogo sembrado muestra los
  5 servicios.
- `TestSpecialties_ListsSeeded` — 5 especialidades.
- `TestCrudviewReuse_NoForkNeeded` — assert de que ambos usan
  `crudview.New(Config{})` sin `Filter`/`List`/`Context` custom (grep en el
  código del test o comprobación estructural) — documenta que el layout se
  reutilizó tal cual.
- Crear un ítem nuevo vía el form → aparece en la lista (op real).

## Criterios de aceptación

- `gotest ./...` verde; build WASM OK; sin fuga de sprite.
- En `localhost:8080`: "Catálogo" y "Especialidades" en el rail, CRUD funcional
  sobre el `item_catalog` real.
- `demoenv` tiene UNA fuente de especialidades (el catálogo real), no una
  constante duplicada.
- `README.md` de `app-demo` lista los módulos nuevos y remarca "mismo `crudview`,
  sin bifurcar" como la lección.

## Fuera de alcance

- `employee_service_config` (la relación staff↔servicio con duración) — eso lo
  siembra `demoenv` para `appointment_booking`, no es una vista del catálogo.
- Precios / inventario / cualquier campo de `CatalogItemModel` que el seed no
  necesite.
