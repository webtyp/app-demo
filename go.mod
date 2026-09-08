module webtyp.com/app-demo

go 1.25.2

require (
	github.com/veltylabs/appointment_booking v0.0.0
	github.com/veltylabs/item_catalog v0.0.0
	github.com/veltylabs/work_schedule v0.0.0
	webtyp.com/components v0.6.20
	webtyp.com/css v0.4.22
	webtyp.com/dom v0.13.12
	webtyp.com/events v0.0.4
	webtyp.com/fmt v1.0.0
	webtyp.com/html v0.0.21
	webtyp.com/input v0.0.6
	webtyp.com/layout v0.2.26
	webtyp.com/model v0.1.8
	webtyp.com/orm v0.12.1
	webtyp.com/router v0.1.36
	webtyp.com/storage v0.0.7
	webtyp.com/svg v0.3.9
	webtyp.com/time v0.5.5
	webtyp.com/unixid v0.2.28
	webtyp.com/view v0.5.2
	webtyp.com/widget v0.6.25
)

require (
	webtyp.com/color v0.1.2 // indirect
	webtyp.com/date v0.0.6 // indirect
	webtyp.com/ddl v0.0.15 // indirect
	webtyp.com/font v0.0.5 // indirect
	webtyp.com/form v0.4.8 // indirect
	webtyp.com/icons v0.0.3 // indirect
	webtyp.com/json v0.5.25 // indirect
)

// Local replaces for unreleased work: the daemon serves these live, so no
// publish is needed to verify in the running demo. Drop each line once its
// repo is published past the change.

replace webtyp.com/icons => ../icons

// Local replace for unreleased work: components carries the new
// scheduleeditor OnWeeklyChange signature and the RevealedBy(widget.Open)
// reveal (the AGENDA_VIEW_FIXES gates). Drop once components publishes the tag.
replace webtyp.com/components => ../components

// widget's Form-holds-Open widening, consumed transitively by the local
// components. Drop once widget publishes the tag.

// Local replace for unreleased work: widget/style's ControlBox now emits
// --control-width, which lives in the local css catalog. Drop once css publishes.

// Los módulos de veltylabs aún no publican los changes de la Etapa C — sirve
// el working tree local (mismo patrón que webtyp.com/components => ../components).
replace github.com/veltylabs/appointment_booking => ../../veltylabs/modules/appointment_booking

// item_catalog: solo se consume para los slugs canónicos de especialidad
// (D6 del master) — la Etapa H lo monta como módulo demo completo.
replace github.com/veltylabs/item_catalog => ../../veltylabs/modules/item_catalog

// work_schedule: vista read-only "horario legado" (Etapa C2) — tag v0.1.4 local.
replace github.com/veltylabs/work_schedule => ../../veltylabs/modules/work_schedule
