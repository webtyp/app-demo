//go:build !wasm

package workschedule

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS es el chrome de la vista workschedule. Nació sin ninguna regla
// (mismo defecto que agenda tenía antes de su fix): display:block por
// defecto, cero padding, <h2> a peso de body. Mismo tratamiento que
// modules/agenda/css.go.
func (s *ScheduleList) RenderCSS() *css.Stylesheet {
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
		Part(PartList,
			style.Stack(style.Space1),
		).
		Stylesheet()
}
