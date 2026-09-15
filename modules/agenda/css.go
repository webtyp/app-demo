//go:build !wasm

package agenda

import (
	"webtyp.com/css"
	"webtyp.com/widget/style"
)

// RenderCSS es el chrome de la vista agenda — lo poco que le queda propio.
// El título, el marco y la región con scroll los pone rightpanel; esta hoja ya
// no los reescribe. Antes esta vista fabricaba su propio chasis (un <h1>
// Text2xl, un Stack con Pad) y por eso no tenía panel ni scroll: el editor se
// cortaba en el borde inferior de la pantalla.
func (s *ScheduleView) RenderCSS() *css.Stylesheet {
	return style.For(s).
		Root(
			style.Fill(),
		).
		// La fila de controles va en el HeadControls de rightpanel, que ya le
		// da su sitio: aquí solo se alinea la etiqueta con el select. Sin
		// As(Panel) ni Round — pintaría una tarjeta dentro de la cabecera que
		// ya es una.
		Part(PartHeader,
			style.Row(style.Space2),
			style.CenterContent(),
		).
		Part(PartStaff,
			style.ControlBox(),
		).
		// El editor respira dentro del article, que es quien hace scroll.
		Part(PartBody,
			style.Stack(style.Space3),
			style.Pad(style.Space3),
		).
		Stylesheet()
}
