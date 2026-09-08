// Package workschedule — pruebo el módulo demo del horario legado sobre el
// módulo real. Consumer-shaped: uso la vista (Render + items) y verifico contra
// las filas de get_work_schedule vía el entorno demo.
package workschedule

import (
	"strings"
	"testing"

	"webtyp.com/dom"

	"webtyp.com/app-demo/config"
)

type emptyCtx struct{}

func (emptyCtx) OnCleanup(func()) {}

func testList(t *testing.T, staffId string) (*scheduleList, *config.Env) {
	t.Helper()
	env := config.New()
	s := &scheduleList{p: nil, env: env, sel: dom.NewString(staffId)}
	s.Init(&emptyCtx{})
	return s, env
}

// TestWorkschedule_ListNatasha: Natasha (segunda en la lista de staff) muestra
// sus 5 días Lun–Vie 09:00–18:00.
func TestWorkschedule_ListNatasha(t *testing.T) {
	s, _ := testList(t, "staff-natasha")

	nodes := s.items.Get()
	if len(nodes) != 5 {
		t.Fatalf("expected 5 schedule rows for Natasha, got %d", len(nodes))
	}
	joined := ""
	for _, n := range nodes {
		joined += n.String()
	}
	if !strings.Contains(joined, "Lunes: 09:00–18:00") {
		t.Errorf("expected the Monday row text, got:\n%s", joined)
	}
}

// TestWorkschedule_ListThor: Thor (tercer staff) tiene 3 filas sembradas
// (Lun/Mié/Vie 08:00–14:00).
func TestWorkschedule_ListThor(t *testing.T) {
	s, _ := testList(t, "staff-thor")

	nodes := s.items.Get()
	if len(nodes) != 3 {
		t.Errorf("expected 3 schedule rows for Thor, got %d", len(nodes))
	}
}

// TestWorkschedule_RenderHasPickerAndTitle: la vista arma el picker de staff y
// el título del panel.
func TestWorkschedule_RenderHasPickerAndTitle(t *testing.T) {
	s, _ := testList(t, "staff-tony")

	html := s.Render().String()
	if !strings.Contains(html, "staff-pick") || !strings.Contains(html, "Horario (sistema legado)") {
		t.Errorf("workschedule view missing picker or title:\n%s", html)
	}
}
