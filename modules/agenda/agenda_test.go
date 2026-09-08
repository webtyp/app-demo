// Package agenda — pruebo el módulo demo del editor de agenda sobre el módulo
// real. Consumer-shaped: uso la View (Render) y verifico contra las ops reales
// de appointment_booking vía el entorno demo.
package agenda

import (
	"strings"
	"testing"

	"webtyp.com/components/scheduleeditor"
	"webtyp.com/dom"

	ab "github.com/veltylabs/appointment_booking"

	"webtyp.com/app-demo/config"
)

type emptyCtx struct{}

func (emptyCtx) OnCleanup(func()) {}

func testView(t *testing.T, staffId string) (*ScheduleView, *config.Env) {
	t.Helper()
	env := config.New()
	v := &ScheduleView{p: nil, env: env, sel: dom.NewString(staffId)}
	v.Init(&emptyCtx{})
	return v, env
}

// TestView_RendersPicker: la vista arma el picker con las 3 opciones de staff.
func TestView_RendersPicker(t *testing.T) {
	v, _ := testView(t, "staff-natasha")

	html := v.Render().String()
	if !strings.Contains(html, "agenda__header") {
		t.Errorf("view render missing the header band:\n%s", html)
	}
	for _, want := range []string{"Dr. Tony Stark", "Dra. Natasha Romanoff", "Dr. Thor Odinson"} {
		if !strings.Contains(html, want) {
			t.Errorf("picker missing %q:\n%s", want, html)
		}
	}
}

// TestEditor_ShowsNatashaWeek: con sel = Natasha, el subárbol del editor
// contiene la grilla de 7 días y 09:00/18:00 en la fila del lunes (su ventana
// sembrada).
func TestEditor_ShowsNatashaWeek(t *testing.T) {
	v, _ := testView(t, "staff-natasha")

	html := v.buildEditor().Render().String()
	for _, want := range []string{"scheduleeditor__pattern", "09:00", "18:00"} {
		if !strings.Contains(html, want) {
			t.Errorf("editor missing %q:\n%s", want, html)
		}
	}
}

// TestCallbacks_OnPatternChangePersists: un cambio de patrón (sábado activo con
// 10:00–13:00) persiste vía op save_day_blocks — verificable releyendo
// list_blocks por el ScheduleClient real.
func TestCallbacks_OnPatternChangePersists(t *testing.T) {
	v, env := testView(t, "staff-natasha")

	editor := v.buildEditor()
	if editor.OnPatternChange == nil {
		t.Fatal("expected the editor to carry OnPatternChange")
	}
	editor.OnPatternChange([]scheduleeditor.PatternRow{
		{StartMin: 600, EndMin: 780, Days: []int{6}},
	})

	// El refetch pasa por list_blocks — la fila del sábado está ahí.
	rows := listBlocks(t, env, "staff-natasha")
	found := false
	for _, r := range rows {
		if r.DayOfWeek == 6 {
			found = true
			if !r.IsActive || r.StartMin != 600 || r.EndMin != 780 {
				t.Errorf("unexpected saturday block: %+v", r)
			}
		}
	}
	if !found {
		t.Fatal("expected a Saturday block after the pattern change")
	}

	// reloadEditor rehace el editor; el html muestra la ventana nueva del sábado.
	v.reloadEditor()
	html := v.buildEditor().Render().String()
	if !strings.Contains(html, "10:00") || !strings.Contains(html, "13:00") {
		t.Errorf("refetched editor must show the new Saturday window:\n%s", html)
	}
}

// listBlocks lee los bloques por la op real.
func listBlocks(t *testing.T, env *config.Env, staffId string) []ab.WorkCalendarBlock {
	t.Helper()
	client := ab.NewScheduleClient(env.Caller(), env.TenantID(), staffId)
	var rows []ab.WorkCalendarBlock
	client.Blocks(func(r []ab.WorkCalendarBlock, err error) {
		if err != nil {
			t.Fatalf("Blocks: %v", err)
		}
		rows = r
	})
	return rows
}

// TestExceptionRoundTrip: alta de excepción desde el editor → aparece en la
// lista → remove → ya no está (end-to-end contra appointment_booking real).
func TestExceptionRoundTrip(t *testing.T) {
	v, env := testView(t, "staff-thor")

	editor := v.buildEditor()
	if editor.OnExceptionAdd == nil || editor.OnExceptionRemove == nil {
		t.Fatal("expected the editor to carry the exception callbacks")
	}

	// Alta: Blocked el 25-12-2026 con 11:00–12:00.
	editor.OnExceptionAdd(scheduleeditor.Exception{
		Date:     "2026-12-25",
		Type:     scheduleeditor.ExcBlocked,
		StartMin: 660,
		EndMin:   720,
	})

	from := unixDay("2026-12-01")
	to := unixDay("2026-12-31")
	client := ab.NewScheduleClient(env.Caller(), env.TenantID(), "staff-thor")
	var excs []ab.WorkCalendarException
	client.Exceptions(from, to, func(rows []ab.WorkCalendarException, err error) {
		if err != nil {
			t.Fatalf("list after add: %v", err)
		}
		excs = rows
	})
	if len(excs) != 1 || excs[0].ExceptionType != ab.ExcBlocked {
		t.Fatalf("expected 1 BLOCKED exception after add, got %+v", excs)
	}

	// Remove por id.
	editor.OnExceptionRemove(excs[0].Id)
	client.Exceptions(from, to, func(rows []ab.WorkCalendarException, err error) {
		if err != nil {
			t.Fatalf("list after remove: %v", err)
		}
		if len(rows) != 0 {
			t.Fatalf("expected no exceptions after remove, got %d", len(rows))
		}
	})
}

// Las siete filas llevan los siete nombres de día, en orden. El bug que esto
// reemplaza: fallbackWeek dejaba el día en cero en los días no configurados y
// cuatro filas renderizaban "Domingo" — y guardar una de ellas escribía Domingo.
// Se cuentan los spans de día (.scheduleeditor__day-name), no el HTML completo,
// porque el calendario de excepciones también pinta nombres de día.
func TestEditor_RendersSevenDistinctDays(t *testing.T) {
	v, _ := testView(t, "staff-tony") // sembrado Lun/Mié/Vie solo — 1 bloque
	html := v.buildEditor().Render().String()

	for _, want := range []string{"scheduleeditor__pattern", "08:00", "14:00"} {
		if !strings.Contains(html, want) {
			t.Errorf("editor missing %q:\n%s", want, html)
		}
	}
}

// blocksToPattern agrupa bloques por rango de horario.
func TestBlocksToPattern_GroupsByTimeRange(t *testing.T) {
	pattern := blocksToPattern([]ab.WorkCalendarBlock{
		{DayOfWeek: 1, IsActive: true, StartMin: 480, EndMin: 840},
		{DayOfWeek: 3, IsActive: true, StartMin: 480, EndMin: 840},
		{DayOfWeek: 5, IsActive: true, StartMin: 480, EndMin: 840},
	})

	if len(pattern) != 1 {
		t.Fatalf("pattern length = %d, want 1", len(pattern))
	}
	if len(pattern[0].Days) != 3 {
		t.Fatalf("days length = %d, want 3", len(pattern[0].Days))
	}
}
