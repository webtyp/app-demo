# PLAN — El CRUD de app-demo no guarda (historial médico) y corrompe registros (todos los módulos)

Estado: **sin despachar**. Fecha del diagnóstico: 2026-09-24.
Versiones en uso por app-demo al diagnosticar: `layout v0.2.47`, `form v0.4.13`
(`form` HEAD = v0.4.16, sin cambios en `sync.go`/`validate.go`/`load.go` desde v0.4.13),
`view v0.6.0`, `input v0.0.9`.

---

## 1. Síntoma y reproducción

| Módulo | Acción | Resultado observado |
|---|---|---|
| Historial Médico | elegir "Juan Pérez" → fila "Control" → editar *diagnosis* → blur | nada: sin aviso, sin error visible; al re-seleccionar la fila vuelve "Sin hallazgos" |
| Computadores | "+" → Pc Prueba / 192.168.122.99 → blur | se crea (parece funcionar) |
| Computadores | después de lo anterior, seleccionar "Pc Calidad" → editar nombre → blur | **"Pc Prueba" desaparece**: la edición sobrescribió el último registro creado; "Pc Calidad" queda intacto |

Es decir: Historial Médico no guarda **nada**, y Computadores "guarda" pero escribe en el
registro equivocado. La inconsistencia entre ambos no es que uno funcione y el otro no:
son **tres defectos distintos** que se enmascaran entre sí.

## 2. Causa raíz — tres defectos

### D1 · app-demo — el esquema rechaza sus propios datos (por qué Historial Médico no guarda)

`modules/medicalhistory/model.go` declara `date` como `input.Text()`. `input.Text` permite
letras, tildes, números, espacios y `. , ( )` — **no `-`**. Todos los `date` son
`"YYYY-MM-DD"`, así que `form.Validate()` falla en **toda** ficha, nueva o existente.
`crudview.saveAction` aborta antes de llamar a `Save`. Computadores no lo sufre porque
`ip` usa `input.IP()` y sus nombres no traen `-`.

Mismo patrón en `modules/reservation/model.go`: `day`/`patient_birthday` (fechas),
`hour` (`HH:MM`), `patient_run` (RUT con `-`) declarados como `input.Text()`.

### D2 · form — un fallo de validación en `Validate()` es silencioso

`form.Validate()` (validate.go) devuelve el primer error pero **no pinta** el
`errorSignals[i]` del campo culpable. La validación en vivo (`fieldComponent.validate`,
render_input.go) sólo corre sobre el campo que el usuario teclea; un valor **cargado**
inválido (la fecha de la ficha) nunca se pinta. `crudview.reportSave` hace `Log` +
`OnSaved(err)`, y el módulo sólo notifica el éxito → el usuario no ve nada.
Viola la regla del harness: *compile error → diagnóstico ruidoso → nunca silencio*.

### D3 · form — el PK oculto no viaja con el registro cargado (corrupción de datos, todos los módulos)

`crudview.saveAction` sincroniza el formulario sobre `Presenter.Record()`, que es **el
mismo registro plantilla** (`&Device{}` pasado a `view.New`) durante toda la vida de la
vista. `form.New` oculta el PK (no tiene Input), y:

- `LoadValues(rec)` copia a los signals sólo los campos con Input → **el Id del registro
  seleccionado nunca se guarda en ningún lado**.
- `SyncValues(target)` sólo asigna Id si `target.Id == ""`; si no, deja el que tenga.
  Su comentario asume *"An EXISTING record already carries its real id here
  (Presenter.Select loaded it…)"* — **falso**: `Select` devuelve la fila indexada, no
  toca `Record()`.

Consecuencias (verificadas en el navegador en Computadores):

1. Página recién cargada, editar un registro existente → plantilla con Id `""` → se
   acuña un Id nuevo → `Create` → **duplicado**; el original queda sin cambios.
2. Tras un "+" que acuñó el Id X en la plantilla, editar cualquier otra fila → `Update`
   sobre X → **se sobrescribe el último registro creado** (lo observado con "Pc Prueba").
3. Un segundo "+" reutiliza X → sobrescribe el primer registro nuevo en vez de crear otro.

`storage/mem` copia valores por columna (no guarda el puntero), así que no hay aliasing:
el defecto es sólo el Id pegado/ausente en la plantilla. En Historial Médico D1 lo tapa;
al corregir D1 sin D3, las ediciones empezarían a duplicar/pisar fichas.

**D3 se corrige en `form`, no en `crudview` ni en `view`**: el formulario es quien oculta
el PK, así que es quien debe recordar qué PK cargó. `crudview` y `view` no cambian de API.

---

## 2bis. Auditoría de tests — por qué nada de esto falló

Existen tests para exactamente estos flujos y todos pasan (`gotest ./crudview` en layout
v0.2.54: verde). Cada uno tiene un hueco concreto:

| Test | Dónde | Hueco |
|---|---|---|
| `conformance.Run` → cláusula `select_fills_form` | `view/conformance/conformance.go` | **Es el test correcto** (Select "2" → editar → exige `ID == "2"`), pero `MockRecord.Schema()` declara `id` **sin** `DB: &model.FieldDB{PK: true}`. Sin PK, `form.New` lo pinta como input visible, `LoadValues` lo carga y el Id viaja. Ningún modelo real tiene un `id` sin PK → la suite certifica una forma de modelo que nadie usa. |
| `TestSyncValuesPreservesExistingHiddenPK` | `form/tests/new_tests_test.go` | **Fija el contrato equivocado**: pone `id` a mano en el mismo objeto que después recibe `SyncValues` y verifica que sobreviva. Nunca hace `LoadValues(A)` → `SyncValues(otro)`, que es lo que hace crudview. Su doc ("an EXISTING record's id must survive") es la suposición falsa de sync.go. |
| `TestLoadValues` (round-trip) | `form/tests/load_test.go` | `WidgetsModel` no tiene PK → el round-trip no cubre el campo que se pierde. |
| `TestConsumer_SaveWithFormData` | `layout/crudview/consumer_test.go` | Hace `f.SetValues("id", "12")`, pero `id` es PK oculto y no tiene Input → **la línea es un no-op silencioso** y el test nunca verifica `dev.Id`. Cree que prueba "guardar con Id", prueba sólo Name/Ip. |
| `TestConsumer_SelectPopulatesForm` | `layout/crudview/consumer_test.go` | Su comentario reconoce el hueco y lo descarta: *"there is nothing id-specific left to assert here"*. Verifica que el form se llene, nunca qué Id se guarda después. |
| Driver `SetField` de la conformance | `layout/crudview/conformance_test.go:35` | `v.form.SetValues(name, value)` sobre un campo inexistente no falla → un test puede "setear" un campo que no existe sin enterarse. |
| `TestConsumer_SaveInvalidForm` | `layout/crudview/consumer_test.go` | Verifica que `OnSaved` reciba error, no que el usuario lo **vea** (error signal del campo) → D2 invisible. |
| `store_save_test.go` / `store_update_test.go` | `app-demo/modules/{devices,medicalhistory}` | Llaman `Saver.Save` directo con un `&Visit{Id: "v1", …}` armado a mano: saltan form, validación y crudview. Prueban el store (bien), pero se presentan como *"regression net for the silent save breakage"* y no pueden ver D1, D2 ni D3. |
| — (no existe) | app-demo | Ningún test valida las semillas contra su propio esquema → D1 (fecha con `-` en `input.Text`) nunca pudo fallar. |

**Confirmado empíricamente** (sonda temporal vía `go test -overlay`, sin tocar el repo):
el mismo flujo de `select_fills_form` pero con el `Device` de `consumer_test.go` (PK
oculto) → `selectAction("23")` → editar → `saveAction` → **guarda con `Id="test-id-1"`**
(recién acuñado), mientras `TestViewConformance/select_fills_form` pasa en la misma corrida.

**Lección de diseño** (va en la etapa de `view`): una suite de conformance que usa un
modelo de juguete más permisivo que los reales no prueba sustituibilidad. El `MockRecord`
debe tener la forma mínima de un modelo real: PK oculto con `DB: &model.FieldDB{PK: true}`.

---

## 3. Gate de diseño (api-design)

### 3.1 Prior art

| Framework | Cómo viaja la identidad del registro editado |
|---|---|
| Django `ModelForm(instance=obj)` | el form queda ligado a la instancia cargada; `save()` actualiza ese `pk`; sin instancia → crea |
| Rails `form_with(model: @record)` | la identidad sale del modelo cargado (`persisted?` + `id` en la ruta/hidden); nunca de un objeto aparte |
| React Hook Form `reset(values)` / Formik `initialValues` | cargar un registro reemplaza **todos** los valores por defecto, id incluido; resetear los borra |

Los tres: **lo que se cargó define la identidad; resetear la borra**. Nuestro `form`
oculta el PK pero no lo trata como estado propio — ése es el hueco. Coincidimos con
Django/RHF: el form recuerda el PK oculto cargado.

Para D2: Django (`form.errors` por campo tras `is_valid()`), RHF (`handleSubmit` pinta
`errors[field]`), Formik (`validateForm` → `touched`+`errors`). Ninguno valida en bloque
sin pintar el campo culpable.

### 3.2 Nombres

**No se agrega ni renombra ningún símbolo público.** Cambia el comportamiento de
`LoadValues`, `Reset`, `SyncValues` y `Validate`, que ya dicen lo que ahora sí harán.

### 3.3 Ledger de complejidad

```
Conceptos que el desarrollador aprende      +0 / −1  (ya no "hay que setear el Id en el record antes de SyncValues")
Archivos que toca para un CRUD correcto     +0 / −0
Líneas en el sitio de llamada               +0 / −0
Formas de hacer lo mismo                    +0 / −1  (el Id ya no puede venir "del target o del form": sólo del form)
```

### 3.4 Dónde vive

- D1 → `app-demo` (esquema del módulo). Es dato de la app, no defecto de librería.
- D2, D3 → `webtyp/form` (dueño del PK oculto y de la validación). `crudview` sólo gana
  un test consumer-shaped que prueba el arreglo a través del stack real.
- La prueba de sustituibilidad (L) → `webtyp/view/conformance`, dueña del contrato
  `Presenter`: ahí se endurece el modelo y se agregan las cláusulas de identidad.

### 3.5 Qué borra este cambio

- La rama de `SyncValues` que lee el PK del target y lo respeta si no está vacío, y su
  comentario falso (sync.go, bloque `for _, idx := range f.hiddenPKIndices`).
- El fallo silencioso de `Validate`.
- `input.Text()` en campos de fecha/hora/RUT de medicalhistory y reservation.
- `TestSyncValuesPreservesExistingHiddenPK` (fija el contrato equivocado) y la línea
  no-op `f.SetValues("id", "12")` de `TestConsumer_SaveWithFormData`.
- El `id` sin PK del `MockRecord` de la conformance.

---

## 4. Etapas

Orden obligatorio: **E0 (view) → E1 (form) → publicar tags → E2 (layout) → publicar tag → E3 (app-demo)**.
E0 va primero a propósito: endurece la suite para que **falle** contra el crudview actual
(prueba roja), y E1 la pone verde. E3 no se despacha hasta tener los tags.

### E0 · `webtyp/view` — la conformance usa un modelo con forma real

Leer `view/AGENTS.md` antes.

1. `conformance/conformance.go`, `MockRecord.Schema()`: `id` →
   `{Name: "id", Type: input.Text(), NotNull: true, DB: &model.FieldDB{PK: true}}`.
   Comentario: la suite debe usar la forma mínima de un modelo real; un `id` sin PK se
   renderiza visible y oculta los defectos del PK oculto (ver PLAN de app-demo §2bis).
2. Cláusulas nuevas en `Run` (mismo estilo que `select_fills_form`, usando sólo el driver):
   - `new_after_new_gets_distinct_ids`: `New` → `SetField("name","A")` → `Save` →
     `New` → `SetField("name","B")` → `Save` → dos `SavedRecords` con `ID` no vacíos y distintos.
   - `edit_after_create_keeps_selected_id`: filas `{1,Alice},{2,Bob}` → `New` →
     `SetField("name","C")` → `Save` → `Select("2")` → `SetField("name","Bob2")` → `Save`
     → el último `SavedRecords` tiene `ID == "2"`.
   - `first_edit_after_mount_keeps_selected_id`: igual que `select_fills_form` pero
     `Select("1")` sin ningún `New` previo → `ID == "1"` (atrapa el caso "duplicado").
   `Driver.New` ya existe (conformance.go:135); no se cambia la API del driver.
3. `view/mock.Renderer` (el segundo implementador) debe pasar las cláusulas nuevas; si no,
   se corrige en esta etapa.
4. `gotest` en view (el mock debe quedar verde) → `gopush`. Se espera que
   `layout/crudview` quede **rojo** contra este tag hasta E1+E2: es la prueba del defecto.

### E1 · `webtyp/form` — PK oculto como estado del form + `Validate` ruidoso

Leer `form/AGENTS.md` antes. WASM/TinyGo: sin maps, sin stdlib, `webtyp.com/fmt`.

**E1.a — PK oculto (D3)**

1. `form.go`, struct `Form`: agregar `loadedPK []string` — paralelo a `hiddenPKIndices`
   (mismo largo, mismo orden). Inicializarlo en `New` con `make([]string, len(f.hiddenPKIndices))`
   justo después de poblar `hiddenPKIndices`.
2. `load.go`, `LoadValues(data)`: tras el loop de Inputs, para cada `k, idx := range f.hiddenPKIndices`:
   `f.loadedPK[k] = fmt.Convert(values[idx]).String()` (guardar `idx < len(values)`).
   La rama `model.IsNil(data)` → `f.reset()` ya cubre el borrado (paso 3).
3. `form.go`, `reset()`: agregar `for k := range f.loadedPK { f.loadedPK[k] = "" }`.
4. `sync.go`, reemplazar el bloque final de `SyncValues` por:
   ```go
   // The hidden PK is form state: LoadValues remembered it, Reset cleared it.
   // The target's own PK is never read — the target may be a scratch record
   // shared across edits (crudview syncs into Presenter.Record()), and trusting
   // it wrote edits onto whichever record was synced last.
   for k, idx := range f.hiddenPKIndices {
   	ptr, st := pointers[idx], schema[idx].Type.Storage()
   	if f.loadedPK[k] == "" && st == model.FieldText {
   		f.loadedPK[k] = f.idGen.NewID() // a new record keeps this id across retries until Reset
   	}
   	if f.loadedPK[k] == "" {
   		zeroField(ptr, st) // int PK of a new record: auto-increment is the DB's job
   		continue
   	}
   	writeField(ptr, st, []string{f.loadedPK[k]})
   }
   ```
   (Si `ShowField` volvió a mostrar el PK, no está en `hiddenPKIndices` y lo maneja el
   loop de Inputs — sin cambio.)
5. Tests nuevos en `form/tests/hidden_pk_test.go` (dual, stdlib asserts, `gotest`),
   con un modelo de prueba de PK texto oculto y un `IDGenerator` fake que devuelve
   `"n1"`, `"n2"`, …:
   - `LoadValues(A{Id:"a"})` → `SyncValues(scratch{Id:"zzz"})` → `scratch.Id == "a"`.
   - `Reset()` → `SyncValues(scratch{Id:"a"})` → `scratch.Id == "n1"`.
   - Dos `SyncValues` seguidos sin `Reset` → mismo `"n1"` (reintento idempotente).
   - `Reset` → `SyncValues` → `Reset` → `SyncValues` → `"n1"` luego `"n2"`.
   - `LoadValues(nil)` → equivale a `Reset`.
   - PK int oculto: `LoadValues(B{Id:7})` → `SyncValues(scratch{Id:99})` → `7`; tras `Reset` → `0`.
6. **Borrar** `TestSyncValuesPreservesExistingHiddenPK` (new_tests_test.go): fija el
   contrato equivocado (§2bis). Lo reemplaza el primer caso del punto 5.
   `TestSyncValuesAssignsHiddenPK` se conserva (sigue siendo cierto con un form recién creado).
7. `form/tests/load_test.go`: agregar al round-trip de `TestLoadValues` un modelo **con PK
   oculto**, verificando que el PK cargado llega al target de `SyncValues`.

**E1.b — `Validate` pinta el campo culpable (D2)**

1. `validate.go`: recorrer **todos** los Inputs (mismo skip de `GetSkipValidation`);
   por cada uno, si falla y `val != ""` → `f.errorSignals[i].Set(err.Error())`; si pasa →
   `f.errorSignals[i].Set("")`. Devolver el **primer** error como hoy.
   Un campo vacío que falla no se pinta: en un borrador el usuario aún no llegó a él, y
   cuando lo toque la validación en vivo (`fieldComponent.validate`) lo pinta. Un valor
   presente e inválido — el caso de D1 — siempre se ve.
2. Test en `form/tests/validate_paints_test.go`: cargar un registro con un valor inválido
   en un campo no tocado → `Validate()` devuelve error **y** el error signal de ese campo
   no está vacío; un campo vacío requerido no queda pintado; un valor corregido limpia el signal.
3. Docs: actualizar `form/README.md` / `docs/` donde se describan `LoadValues`,
   `SyncValues` y `Validate` (una línea de contrato cada uno). Publicar con `gopush`.

### E2 · `webtyp/layout` — tests de crudview que sí miran el Id

Sin cambio de código productivo. Actualizar `form` (E1) y `view` (E0) a sus tags: la
conformance endurecida debe pasar sin tocar crudview.go — ésa es la prueba de que D3 era
de `form`.

Arreglar los tests huecos (§2bis):

1. `consumer_test.go`, `TestConsumer_SaveWithFormData`: borrar `f.SetValues("id", "12")`
   (no-op). Verificar `dev.Id != ""` (Id acuñado para un registro nuevo).
2. `consumer_test.go`, `TestConsumer_SelectPopulatesForm`: borrar el comentario que
   descarta el Id; agregar editar + `saveAction` y verificar que se guarda `Id == "12"`.
3. `conformance_test.go`, driver `SetField`: si `v.form.Input(name) == nil` →
   `t.Fatalf("SetField: form has no input %q (hidden PK?)", name)`. Un test no puede
   volver a "setear" un campo que no existe.
4. `consumer_test.go`, `TestConsumer_SaveInvalidForm`: además del error en `OnSaved`,
   verificar que el error signal del campo culpable no esté vacío (D2).

Y agregar `crudview/record_identity_test.go` que use **el stack real** (`crudview.New`
+ `view.New` + `form` real + el `Device` de consumer_test.go, que tiene PK oculto, + un
`Lister` fake en memoria con `Save`/`Delete` que registra lo recibido):

- seleccionar fila A, cambiar un campo, `autoSaveAction()` → el fake recibe Id `A`, y
  el total de filas no cambia.
- "+", llenar, guardar → Id nuevo N1; "+" otra vez, llenar, guardar → Id N2 ≠ N1.
- crear N1, luego seleccionar A y editar → el fake recibe `A`, **nunca** `N1`.
- registro con valor inválido cargado → `autoSaveAction()` no llama a `Save` y
  `OnSaved` recibe el error.

`gotest` → `gopush`.

### E3 · `app-demo` — esquemas correctos + guardia de semillas

1. `go get` de los tags de `form` (E1) y `layout` (E2); `go mod tidy`.
2. `modules/medicalhistory/model.go`: `date` → `input.Date()`.
3. `modules/reservation/model.go`: `day` y `patient_birthday` → `input.Date()`;
   `hour` → `input.Hour()`; `patient_run` → `input.Rut()`; `patient_contact` →
   `input.Phone()` si el dato semilla lo permite. **Regla**: el tipo de input se elige por
   el formato real del dato; si un dato semilla no pasa (p. ej. RUT con puntos:
   `input.Rut` sólo permite dígitos, `-`, `k`), se normaliza la semilla al formato del
   input, no se relaja el input a `Text`.
4. Guardia contra la regresión (D1 no vuelve): en cada módulo local con semilla
   (`devices`, `medicalhistory`, `reservation`) un test `seed_validates_test.go` que
   construye `form.New(...)` con el modelo del módulo, hace `LoadValues(row)` y
   `Validate()` por **cada** fila semilla, y falla nombrando fila y campo.
5. `store_save_test.go` de `devices` y `medicalhistory`: corregir su doc — prueban el
   store, no el flujo de guardado de la UI; quitar "regression net for the silent save
   breakage". El flujo de UI lo prueban E0/E2 en las librerías y el test de semillas aquí.
6. `cv.OnSaved` en `devices` y `medicalhistory`: con `err != nil` notificar
   `m.p.Notify(Msg.Error, err.Error(), platformd.Auto())` — el campo ya se pinta (E1.b),
   el toast confirma que el guardado no ocurrió.

---

## 5. Verificación end-to-end (app corriendo, MCP `webtyp`)

1. `gotest` verde en `form`, `layout/crudview`, `app-demo`.
2. Historial Médico: elegir "Juan Pérez" → fila "Control" → editar diagnosis → blur →
   toast "Guardado"; seleccionar otra fila y volver → el cambio persiste; la lista sigue
   con 2 fichas (sin duplicado).
3. Historial Médico: "+" con patient "Juan Pérez" y fecha `2026-09-24` → aparece como
   tercera ficha; segundo "+" → cuarta ficha (no pisa la tercera).
4. Computadores: "+" Pc Prueba → luego editar "Pc Calidad" → **ambos** existen, Pc
   Calidad con el nombre nuevo; total = 16.
5. Fecha inválida tecleada (`2026-13-40`) → el campo se pinta en rojo, toast de error,
   nada se guarda.
6. Borrado individual y masivo, y ✏ masivo en Computadores siguen funcionando.

## 6. Fuera de alcance (observado, no se toca aquí)

- Tras recargar la lista en Computadores quedó una fila huérfana duplicada al final
  ("Servidor Monitoreo", nodo `id=17` de la primera pintura) — defecto de
  reconciliación de `targetlist.SetItems`; va en su propio plan en `components`.
