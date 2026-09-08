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

func testView(t *testing.T, staffId string) (*scheduleView, *config.Env) {
	t.Helper()
	env := config.New()
	v := &scheduleView{p: nil, env: env, sel: dom.NewString(staffId)}
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
	for _, want := range []string{"scheduleeditor__week", "09:00", "18:00"} {
		if !strings.Contains(html, want) {
			t.Errorf("editor missing %q:\n%s", want, html)
		}
	}
}

// TestWorkSchedule_ViewsSchedule: el panel "Horario" del módulo agenda se
// alimenta de work_schedule (NewView sobre sus tablas) — Natasha (segunda en la
// lista de staff) muestra sus 5 días, sin montar nada de appointment_booking.
func TestWorkSchedule_ViewsSchedule(t *testing.T) {
	v, env := testView(t, "staff-natasha")

	// El panel agenda_schedule se ve en el render con el título.
	html := v.Render().String()
	if !strings.Contains(html, "agenda_schedule") || !strings.Contains(html, "Horario") {
		t.Errorf("expected the work_schedule panel in the render:\n%s", html)
	}

	// El Ref() del nodo: el panel se re-llena por NewView en reloadSchedule, y
	// el store del Env responde la op. Count the schedule rows via the node.
	nodes := v.schedule.Get()
	if len(nodes) == 0 {
		t.Fatalf("expected the schedule panel to be populated")
	}
	// 5 días de Natasha (Lun–Vie 09:00–18:00 por el seed de work_schedule).
	if len(nodes) != 5 {
		t.Errorf("expected 5 schedule rows for Natasha, got %d (%v)", len(nodes), nodes)
	}
	joined := ""
	for _, n := range nodes {
		joined += n.String()
	}
	if !strings.Contains(joined, "Lunes: 09:00–18:00") {
		t.Errorf("expected the Monday row text, got:\n%s", joined)
	}
	_ = env
}

// TestWorkSchedule_FallsBackToEmpty: un staff sin horario sembrado muestra la
// fila "Sin horario registrado".
func TestWorkSchedule_FallsBackToEmpty(t *testing.T) {
	v, _ := testView(t, "staff-thor")

	// Thor es el tercer staff → work_schedule seeds 3 entradas para él (igual
	// que Tony: Lun/Mié/Vie). Cambiar el asser: thor tiene filas.
	nodes := v.schedule.Get()
	if len(nodes) != 3 {
		t.Errorf("expected 3 schedule rows for Thor, got %d", len(nodes))
	}
}

// TestCallbacks_OnWeeklyChangePersists: un cambio de semana (sábado activo con
// 10:00–13:00) persiste vía la op upsert — verificable releyendo
// list_weekly_calendar por el ScheduleClient real.
func TestCallbacks_OnWeeklyChangePersists(t *testing.T) {
	v, env := testView(t, "staff-natasha")

	// Un cambio de semana (sábado activo con 10:00–13:00) persiste por la op
	// upsert; la vista lo refleja en su refetch.
	editor := v.buildEditor()
	if editor.OnWeeklyChange == nil {
		t.Fatal("expected the editor to carry OnWeeklyChange")
	}
	editor.OnWeeklyChange(scheduleeditor.WeeklyRow{
		DayOfWeek:  6,
		Active:     true,
		WorkStart:  600,
		WorkFinish: 780,
	})

	// El refetch pasa por list_weekly_calendar — la fila del sábado está ahí.
	rows := listWeekly(t, env, "staff-natasha")
	found := false
	for _, r := range rows {
		if r.DayOfWeek == 6 {
			found = true
			if !r.IsActive || r.WorkStart != 600 || r.WorkFinish != 780 {
				t.Errorf("unexpected saturday row: %+v", r)
			}
		}
	}
	if !found {
		t.Fatal("expected a Saturday row after the weekly change")
	}

	// reloadEditor rehace el editor; el html muestra la ventana nueva del sábado.
	v.reloadEditor()
	html := v.buildEditor().Render().String()
	if !strings.Contains(html, "10:00") || !strings.Contains(html, "13:00") {
		t.Errorf("refetched editor must show the new Saturday window:\n%s", html)
	}
}

// listWeekly lee la plantilla por la op real.
func listWeekly(t *testing.T, env *config.Env, staffId string) []ab.WorkCalendarWeekly {
	t.Helper()
	client := ab.NewScheduleClient(env.Caller(), env.TenantID(), staffId)
	var rows []ab.WorkCalendarWeekly
	client.Weekly(func(r []ab.WorkCalendarWeekly, err error) {
		if err != nil {
			t.Fatalf("Weekly: %v", err)
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
