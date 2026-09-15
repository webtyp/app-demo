//go:build !wasm

package agenda

import "webtyp.com/svg/sprite"

func (m *Module) IconSvg() *sprite.Sprite {
	// Una cuadrícula de semana, NO un reloj. Reserva Hora ya es el reloj, y su
	// glifo y este eran el mismo círculo con manecillas: en un rail de ocho
	// íconos, dos idénticos obligan a leer la etiqueta, que es exactamente lo
	// que un ícono existe para evitar.
	//
	// La distinción es de significado, no de adorno: Reserva Hora agenda UN
	// momento — un reloj —; Agenda define QUÉ DÍAS se trabaja — una semana. La
	// cuadrícula con la columna del día libre vacía es lo que se edita en esa
	// pantalla, dibujado.
	//
	// Mismo viewBox con 2 unidades de aire por lado que reservation/svg.go,
	// para que los dos glifos se rendericen del mismo tamaño óptico.
	return sprite.NewSprite(
		sprite.Define(
			Icon,
			"-2 -2 20 20",
			sprite.Path("M1 1h14a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H1a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1zm.5 4.5v9h13v-9h-13zM3 7h2.5v2H3V7zm4 0h2.5v2H7V7zm4 0h2.5v2H11V7zM3 10.5h2.5v2H3v-2zm4 0h2.5v2H7v-2z"),
		),
	)
}
