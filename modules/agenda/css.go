//go:build !wasm

package agenda

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS es el chrome de la vista agenda. Antes, el módulo escribía cuatro
// nombres de clase que ninguna hoja mencionaba: la vista se renderizaba como
// display:block con cero padding, pegada al borde, y su <h2> era texto de body.
func (s *ScheduleView) RenderCSS() *css.Stylesheet {
	return style.For(s).
		Root(
			style.Stack(style.Space4),
			style.Pad(style.Space4),
		).
		Part(PartTitle,
			style.FontSize(style.Text2xl),
			style.FontWeight(style.WeightBold),
		).
		Part(PartHeader,
			style.Row(style.Space2),
			style.ControlBox(),
			style.As(style.Panel),
			style.Round(style.RadiusMd),
		).
		Part(PartStaff,
			style.ControlBox(),
		).
		Part(PartSection,
			style.FontSize(style.TextLg),
			style.FontWeight(style.WeightBold),
		).
		Part(PartBody,
			style.Stack(style.Space3),
		).
		Stylesheet()
}
