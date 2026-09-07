//go:build !wasm

package agenda

import "webtyp.com/svg/sprite"

func (m *Module) IconSvg() *sprite.Sprite {
	// Un reloj con manecillas — la marca del "editor de agenda".
	return sprite.NewSprite(
		sprite.Define(
			Icon,
			"-2 -2 20 20",
			sprite.Path("M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm0 14.5a6.5 6.5 0 1 1 0-13 6.5 6.5 0 0 1 0 13zM7.25 4v4.5l3.5 2 .75-1.25-3-1.72V4h-1.25z"),
		),
	)
}
