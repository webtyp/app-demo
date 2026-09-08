//go:build !wasm

package workschedule

import "webtyp.com/svg/sprite"

func (m *Module) IconSvg() *sprite.Sprite {
	// Un documento con líneas — la marca del "horario legado" (lista read-only).
	return sprite.NewSprite(
		sprite.Define(
			Icon,
			"-2 -2 20 20",
			sprite.Path("M4 2h8l4 4v8a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2zm8 1.2V6h2.8L12 3.2zM6 8.5h8M6 11.5h8M6 14.5h5"),
		),
	)
}
