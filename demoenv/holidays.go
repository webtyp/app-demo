// Package demoenv — feriados para las excepciones del seed. En producción
// estos vienen de un servicio de feriados; en la demo es una constante local,
// documentada como tal para que nadie la confunda con datos de un cliente.
package demoenv

// holidaysCL2026 son los feriados de Chile del 2026 ("YYYY-MM-DD"), los mismos
// que un servicio de feriados entregaría. Solo lectura en la UI (ScheduleEditor
// los muestra sin "Quitar").
func holidaysCL2026() []string {
	return []string{
		"2026-01-01", // Año Nuevo
		"2026-05-01", // Día del Trabajador
		"2026-06-29", // San Pedro y San Pablo
		"2026-07-16", // Virgen del Carmen
		"2026-08-15", // Asunción de la Virgen
		"2026-09-18", // Fiestas Patrias
		"2026-09-19", // Glorias del Ejército
		"2026-10-12", // Encuentro de Dos Mundos
		"2026-10-31", // Día de las Iglesias Evangélicas
		"2026-11-01", // Día de Todos los Santos
		"2026-12-08", // Inmaculada Concepción
		"2026-12-25", // Navidad
	}
}
