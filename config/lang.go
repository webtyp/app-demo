// Package config — this file is the demo's ONE translation dictionary.
//
// Every webtyp library renders its own chrome text (crudview's confirm
// dialog, calendarslider's month/weekday names, …) through
// webtyp.com/fmt/lang, always keyed in English — a library never
// hardcodes a human language, never calls lang.RegisterWords itself (see
// layout/AGENTS.md's "Translatable messages" section and
// layout/docs/DICTIONARY.md). The demo reads in Spanish because THIS file
// says so, not because any library is Spanish. Any other webtyp-framework
// app registers its own words here the same way — config/lang.go is the
// convention (see the project-layout skill), not something specific to
// this demo.
package config

import "webtyp.com/fmt/lang"

func init() {
	lang.RegisterWords([]lang.DictEntry{
		// layout/crudview's delete-confirmation dialog — see
		// layout/docs/DICTIONARY.md for the full key list this mirrors.
		{EN: "Confirm", ES: "Confirmar"},
		{EN: "Cancel", ES: "Cancelar"},
		{EN: "Delete", ES: "Eliminar"},
		{EN: "This", ES: "Esta"},
		{EN: "action", ES: "acción"},
		{EN: "cannot", ES: "no"},
		{EN: "be", ES: "se"},
		{EN: "undone.", ES: "puede deshacer."},
		{EN: "records", ES: "registros"},

		// components/scheduleeditor's exception type + chrome labels (see
		// components/scheduleeditor/README.md — "Translation keys").
		{EN: "Closed", ES: "Cerrado"},
		{EN: "Special hours", ES: "Horario especial"},
		{EN: "No exceptions", ES: "Sin excepciones"},
		{EN: "Blocked", ES: "Bloqueado"},
		{EN: "Add", ES: "Agregar"},
		{EN: "Remove", ES: "Quitar"},
		{EN: "Type", ES: "Tipo"},
		{EN: "Date", ES: "Fecha"},
		{EN: "Notes", ES: "Notas"},

		// components/scheduleeditor's pattern editor: one row per time range
		// plus its day chips, and the marked-days block. The old weekly-grid
		// keys (Day / Work start / Break start / …) died with v0.6.23 — the
		// model is blocks now, and the break is the gap between two rows, not a
		// field of its own.
		{EN: "Weekly pattern", ES: "Patrón semanal"},
		{EN: "Add row", ES: "Agregar fila"},
		{EN: "Remove row", ES: "Quitar fila"},
		{EN: "Marked days", ES: "Días marcados"},
		{EN: "Hours for marked days", ES: "Horario de días marcados"},

		// scheduleeditor's day chips (shortWeekdayKeys) — the short form;
		// calendarslider renders the long names registered below.
		{EN: "Sun", ES: "Dom"}, {EN: "Mon", ES: "Lun"},
		{EN: "Tue", ES: "Mar"}, {EN: "Wed", ES: "Mié"},
		{EN: "Thu", ES: "Jue"}, {EN: "Fri", ES: "Vie"},
		{EN: "Sat", ES: "Sáb"},

		// components/calendarslider's month/weekday names (date.MonthName /
		// date.WeekdayName return these English keys; not yet called by any
		// module in this demo, registered ahead of time so nothing needs to
		// come back here once that lands).
		{EN: "January", ES: "Enero"}, {EN: "February", ES: "Febrero"},
		{EN: "March", ES: "Marzo"}, {EN: "April", ES: "Abril"},
		{EN: "May", ES: "Mayo"}, {EN: "June", ES: "Junio"},
		{EN: "July", ES: "Julio"}, {EN: "August", ES: "Agosto"},
		{EN: "September", ES: "Septiembre"}, {EN: "October", ES: "Octubre"},
		{EN: "November", ES: "Noviembre"}, {EN: "December", ES: "Diciembre"},
		{EN: "Sunday", ES: "Domingo"}, {EN: "Monday", ES: "Lunes"},
		{EN: "Tuesday", ES: "Martes"}, {EN: "Wednesday", ES: "Miércoles"},
		{EN: "Thursday", ES: "Jueves"}, {EN: "Friday", ES: "Viernes"},
		{EN: "Saturday", ES: "Sábado"},
	})

	// The activation, not just the dictionary — without this, fmt/lang's own
	// default (auto-detect, falling back to English) decides, and the demo
	// would render in English on any machine/browser that isn't set to
	// Spanish. This is what makes the demo's language a deliberate choice
	// instead of an environment accident.
	lang.OutLang(lang.ES)
}
