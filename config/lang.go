// Package config — this file is the demo's ONE translation dictionary.
//
// Every tinywasm library renders its own chrome text (crudview's confirm
// dialog, calendarslider's month/weekday names, …) through
// github.com/tinywasm/fmt/lang, always keyed in English — a library never
// hardcodes a human language, never calls lang.RegisterWords itself (see
// layout/AGENTS.md's "Translatable messages" section and
// layout/docs/DICTIONARY.md). The demo reads in Spanish because THIS file
// says so, not because any library is Spanish. Any other tinywasm-framework
// app registers its own words here the same way — config/lang.go is the
// convention (see the project-layout skill), not something specific to
// this demo.
package config

import "github.com/tinywasm/fmt/lang"

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
