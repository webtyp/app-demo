//go:build !wasm

package reservation

import "github.com/tinywasm/svg/sprite"

func (m *Module) IconSvg() *sprite.Sprite {
	// Optical-size contract: a glyph should fill ~3/4 of its viewBox with
	// even air on all sides — the circle below spans the full 16 units, so
	// the viewBox carries a 2-unit pad on every side (16/20 = 80% fill).
	// Same-size boxes render same-size glyphs only when the art agrees.
	return sprite.NewSprite(
		sprite.Define(
			Icon,
			"-2 -2 20 20",
			sprite.Path("M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm0 14.5a6.5 6.5 0 1 1 0-13 6.5 6.5 0 0 1 0 13zM7.25 4v4.5l3.75 2.25.75-1.23-3-1.77V4h-1.5z"),
		),
	)
}
