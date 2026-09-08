// Package staffpick es el único <select> de profesional de la demo. Dos módulos
// necesitan acotar una vista a un profesional; el control se escribe aquí para
// no deletrearlo dos veces. Es chrome de demo, no un componente de framework —
// una app real lo tomaría de su módulo de staff.
package staffpick

import (
	. "webtyp.com/dom"
	. "webtyp.com/html"

	"webtyp.com/app-demo/config"
)

// Select arma el picker de profesional. sel guarda el StaffOption.ID elegido;
// onChange se dispara después de actualizar sel, para que el caller pueda
// recargar.
func Select(staff []config.StaffOption, sel *SignalString, onChange func(id string)) *Element {
	el := NewElement("select").Attr("name", "staff-pick")
	for _, so := range staff {
		if so.ID == sel.Get() {
			el.Child(SelectedOption(so.ID, so.Name))
		} else {
			el.Child(Option(so.ID, so.Name))
		}
	}
	el.On("change", func(ev Event) {
		sel.Set(ev.TargetValue())
		if onChange != nil {
			onChange(ev.TargetValue())
		}
	})
	return el
}
