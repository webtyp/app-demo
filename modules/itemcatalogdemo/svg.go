//go:build !wasm

package itemcatalogdemo

import "webtyp.com/svg/sprite"

func (m *catalogMod) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		sprite.Define(
			IconCatalog,
			"-2 -2 20 20",
			sprite.Path("M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm0 14.5a6.5 6.5 0 1 1 0-13 6.5 6.5 0 0 1 0 13zM5 7h2v2H5zm6 0h2v2h-2zM7 4h2v2H7z"),
		),
	)
}

func (m *specialtyMod) IconSvg() *sprite.Sprite {
	return sprite.NewSprite(
		sprite.Define(
			IconSpecial,
			"-2 -2 20 20",
			sprite.Path("M8 0a8 8 0 1 0 0 16A8 8 0 0 0 8 0zm0 14.5a6.5 6.5 0 1 1 0-13 6.5 6.5 0 0 1 0 13zM7.5 4.5h1v3h3v1h-3v3h-1v-3h-3v-1h3v-3z"),
		),
	)
}
