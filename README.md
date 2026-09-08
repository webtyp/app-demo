# app-demo
<img src="docs/img/badges.svg">

WebTyp demo app — the `platformd` shell wired with CRUD modules plus real
domain modules from `github.com/veltylabs/modules/*`, served from this
dedicated repo so the demo never weighs on library `go.mod`s.

The reference demo used by [github.com/webtyp/layout](https://github.com/webtyp/layout)
(`platformd`): a playground to import ANY webtyp package freely and test it
in a running app.

## Modules

- `modules/devices` — full CRUD (form + list) over an in-memory `storage/mem` store via `layout/crudview`.
- `modules/medicalhistory` — CRUD with a `selectsearch`/`targetdate` list.
- `modules/reservation` — reservation over the REAL `appointment_booking` module: `calendarslider` filter + `targethour` list, `area/doctor` Context selects, and `schedule.changed` sync (Etapa F/G).
- `modules/agenda` — schedule editor over the REAL `appointment_booking` module (`ScheduleClient` + `ScheduleEditor`): weekly template + per-date exceptions, via `loopback.New` (Etapa D).
- `modules/itemcatalog` — TWO modules over the REAL `item_catalog`: "Catálogo" (`NewView`) and "Especialidades" (`NewSpecialtyView`), both mounted on `crudview.New(Config{})` with NO custom config — proof the layout is reused as-is (Etapa H).
- `modules/about` — static module.
- `web/client.go` — the composition root: `platformd.Platform` + `hiddenModule`
  (exercises `CanView`), mock brand/identity, `themetoggle` user action, and
  the `fieldset` global form skin.

## `config` — the demo composition root for real modules

`config.New()` builds, once, the infrastructure the real-module wrappers
share: an in-memory `orm.DB` (`storage/mem`), the `veltylabs` domain modules
mounted on a `router/loopback` `Caller` (in-process, no server), an
`events/mock` broker, a stable `unixid` generator, the list of demo staff +
national holidays, and a realistic seed. Modules like `agenda` call
`config.Caller()` to talk to `appointment_booking`. It is the demo analogue
of what `config/client.go` does in `mjosefa-cms` — a composition seam, not
domain logic (it also carries the Spanish dictionary, `config/lang.go`, and
the theme, `config/css.go`).

## Releases

The real modules are pinned by version and pointed at the local working tree
while they are in development (same pattern as `webtyp.com/components`):

```
replace github.com/veltylabs/appointment_booking => ../../veltylabs/modules/appointment_booking
replace github.com/veltylabs/item_catalog       => ../../veltylabs/modules/item_catalog
```

Drop each `replace` once the repo publishes the needed tag (a `go.mod` clean
of `replace`s means every dependency is released).

## Translations

`config/lang.go` is the demo's one dictionary: it registers the
English→Spanish words the framework's own chrome needs (`layout/crudview`'s
confirm dialog, `components/calendarslider`'s month/weekday names,
`components/scheduleeditor`'s exception labels, …) via
`webtyp.com/fmt/lang`, and activates Spanish
(`lang.OutLang(lang.ES)`). Libraries never hardcode a language themselves —
they render an English key through `lang.Translate(...)`; this file is what
turns that into Spanish for the demo. `web/client.go` blank-imports
`config` so the registration actually runs. Adding a module whose framework
component has its own translatable chrome? Add its words to this same file.

## Sincronía por eventos (agenda → reserva)

`reservation` se suscribe al broker in-proc de la demo (`config/env.go` →
`appointment_booking` publica `schedule.changed` al mutar la agenda) y a los
eventos de reserva. Al cambiar la agenda del médico en foco (desde `agenda`),
la lista y los huecos libres se recalculan **sin recargar la página**. El
broker in-proc es el `events/mock` (map-free — decisión O1 del master); en una
app real (`mjosefa-cms`) el mismo contrato lo cumple `webtyp/sse`, que ya
satisfice `events.Publisher`.

## Brecha conocida — el módulo Directory

`appointment_booking` referencia al paciente por `client_id` (validado por un
`DirectoryReader`). La demo no tiene un módulo Directory: el formulario de
`reservation` mantiene los campos de paciente como **texto libre** (run, nombre,
contacto) y al crear usa el run como `client_id` sintético; `config` usa un
`stubDirectory` que responde `true` y deja un comentario de qué haría el módulo
real. Es el mismo trade-off explícito que `mjosefa-cms/modules/clinical_encounter`
documenta para su picker de paciente — deuda visible, no oculta.

## Run

    webtyp            # from this repo; dev server :8080, MCP :6060

## Layout dependency

The demo consumes the shell via a local replace while layout is in monorepo
development:

    replace webtyp.com/layout => ../layout

With no replace, resolution falls back to the published `webtyp.com/layout`
module. (A clone outside this workspace needs either the published version or
the `../layout` checkout next to it.)