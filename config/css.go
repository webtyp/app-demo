//go:build !wasm

package config

import "webtyp.com/css"

// Theme is the visual composition root. sitec discovers RootCSS() while walking
// the project at build time and emits it into web/public/style.css.
type Theme struct{}

// RootCSS uses webtyp's own default palette untouched: ColorPrimary's
// 135° gradient (WebAssembly violet → Go cyan) is now the framework's own
// out-of-box look (see webtyp.com/css's brandRoot/ColorPrimaryGradient), so
// app-demo no longer needs to declare it itself.
func (Theme) RootCSS() *css.Stylesheet {
	return css.Theme()
}
