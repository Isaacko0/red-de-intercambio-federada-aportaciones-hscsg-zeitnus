# Cosateca y Objetoteca: bienes comunes fisicos para ecoaldeas federadas

**Modulo:** `internal/api/cosateca.go` + `internal/db/migrations/171_cosateca_schema.sql` + `web/src/pages/Cosateca.tsx`
**Estado:** implementado en el nodo (local-first, listo para federar)
**Repo:** `https://github.com/Isaacko0/red-de-intercambio-federada-aportaciones-hscsg-zeitnus`

## 1. La idea

Partimos de la propuesta de cosatecas y objetotecas dentro del OAD de Integral (documento adjunto, platica entre amigos):

> Una objetoteca funciona como una biblioteca, pero en lugar de prestar libros, presta objetos.
> Una cosateca organiza cosas utiles de la vida cotidiana: objetos reutilizables, materiales, piezas recuperadas y recursos que circulan entre personas y proyectos comunitarios.

El ciclo que cierra el circuito es el de Integral:

1. **OAD (Open Access Design)** — sabemos como hacerlo: planos, lista de materiales, instrucciones de reparacion y mejoras compartidas.
2. **Produccion comunitaria (COS)** — la comunidad fabrica o adapta el objeto con sus recursos.
3. **Objetoteca** — el objeto se registra y se gestiona para que cualquiera lo use cuando lo necesita.
4. **Aprendemos y mejoramos (FRS)** — si el objeto falla o consume demasiado, documentamos la experiencia y retroalimentamos el diseno en el OAD.

Aplicado a las Ecoaldeas Federadas (Cergio Monasterio): una red de comunidades autónomas que intercambian productos y conocimientos y construyen soberania material. Ejemplo del documento adjunto: una ecoaldea consulta el catalogo federado de disenos de muebles, adapta los planos a la madera de su region, pide prestadas las herramientas compartidas de la cosateca, fabrica con trabajo comunitario, registra que materiales funcionaron mejor y comparte las mejoras con las otras ecoaldeas.

## 2. Mapeo al modelo Integral (OAD / COS / ITC / FRS)

El devguide de Integral define cinco subsistemas: CDS (gobernanza), OAD (diseños), COS (produccion), ITC (creditos de contribucion) y FRS (retroalimentacion). El loop celular es **CDS → OAD → COS → ITC → FRS** (mismo loop que HSCSG v15 OS mide con la metrica ICS).

| Subsistema Integral | Que hace | Donde vive en este nodo |
|---|---|---|
| **OAD** | Repositorio de diseños certificados: planos, lista de materiales, coeficientes ecologicos, protocolo de mantenimiento. | Referenciado por `cosateca_objects.design_ref`. El modulo OAD completo vive en HSCSG v15 OS; aqui el objeto solo apunta al diseno certificado (nada se produce sin diseno, igual que COS no organiza trabajo sin blueprint). |
| **COS** | Organiza el trabajo a partir de un diseno certificado: tareas, materiales, controles de calidad, distribucion. | Este modulo: catalogo de objetos, prestamos, mantenimiento, estado de cada pieza. |
| **ITC** | Creditos de contribucion: no son dinero, no son transferibles, se extinguen al acceder a bienes. | **No se usa ITC aqui.** Este nodo lleva **TQ** (ver seccion 3). ITC y TQ son sistemas distintos; no los tratamos como equivalentes sin una integración disenada explicitamente. |
| **FRS** | Senales operativas: uso real, averias, desviaciones del diseno; recomendaciones a CDS/OAD. | Prestamos vencidos, objetos en mantenimiento, objetos retirados = senales crudas para FRS (pendiente de cablear al modulo FRS del nodo). |

## 3. TQ (Ecoaldeas Federadas) vs ITC (Integral): una cautela

El Libro Ecoaldeas Federadas propone el **TQ**: 1 TQ = 1 kWh = 3.6 MJ, una "cinta metrica contable" para medir reciprocidad. No es dinero, no genera intereses, no se acumula infinitamente (limite simetrico -500/+500 TQ, escalable por asamblea), y esta **prohibido intercambiar TQ directamente por fiat o cripto** (regla de la aldea). El valor entra a la red transformando dinero en bienes tangibles (herramientas, paneles solares) que se truecan al colectivo por los TQ correspondientes a su energia incorporada.

La documentacion de Integral usa **ITC** (creditos de contribucion y acceso), un sistema distinto: creditos no transferibles que se extinguen al acceder a bienes y decaen con el tiempo.

**Decisión de diseño en este modulo:** el nodo ya opera con TQ (ledger de `users.balance`, tabla `transactions`, `ledger_entries`). Las aportaciones a la cosateca se reconocen en TQ (credito al balance del contribuyente, contrapartida al puente nodal `node_bridge_global`, mismo patron de escritura directa que usa `internal/federation`). No mezclamos ITC con TQ: si una ecoaldea quiere operar ITC en vez de TQ, es una decision de asamblea (CDS) y requiere un modulo de contabilidad aparte.

## 4. Tres capas, nunca confundidas

El documento adjunto distingue deliberadamente:

1. **El objeto**: la herramienta fisica que existe y debe conservarse (`cosateca_objects`: nombre, categoria, estado, ubicacion, diseno_ref de OAD, valor incorporado en TQ).
2. **El registro**: disponibilidad, prestamos, estado, historial de mantenimiento (`cosateca_loans`, `cosateca_maintenance`).
3. **La reciprocidad**: acuerdos de la comunidad para distribuir el acceso y reconocer aportaciones, integrados con TQ segun las reglas aprobadas por la ecoaldea (`cosateca_contributions` + credito TQ al contribuyente).

Cada ecoaldea (nodo) acuerda que herramientas son personales, cuáles se prestan y cuáles pertenecen al fondo colectivo.

## 5. Implementación en este repo

### Base de datos
Migración `171_cosateca_schema.sql` (aplicada automaticamente al arrancar el nodo via `RunMigrations`):

- `cosateca_objects` — catalogo de objetos compartidos (`design_ref` apunta al diseno certificado del OAD; `tq_value` energia incorporada; `federated_visible` para compartir con nodos federados).
- `cosateca_loans` — prestamos (estado activo/devuelto/vencido, fecha de devolucion).
- `cosateca_maintenance` — reparaciones, limpiezas, calibraciones, revisiones.
- `cosateca_contributions` — aportaciones al fondo colectivo con `tq_value` reconocido por la asamblea y `credited` para controlar el abono en el ledger TQ.

### API (autenticada, prefijo `/api/cosateca`)

| Endpoint | Metodo | Que hace |
|---|---|---|
| `/api/cosateca/objects` | GET | Lista objetos (filtros `status`, `category`, `search`, `limit`, `offset`) |
| `/api/cosateca/objects` | POST | Aporta/registra un objeto a la cosateca |
| `/api/cosateca/objects/{id}` | GET | Detalle de un objeto (incluye prestamo activo si lo tiene) |
| `/api/cosateca/objects/{id}` | PATCH | Edita (solo quien lo aporto) |
| `/api/cosateca/objects/{id}` | DELETE | Elimina (si no tiene prestamos activos) |
| `/api/cosateca/loans` | GET | Mis prestamos (`mine=1`) o todos |
| `/api/cosateca/loans` | POST | Solicita prestamo de un objeto disponible |
| `/api/cosateca/loans/{id}/return` | POST | Devuelve el objeto |
| `/api/cosateca/maintenance` | GET | Historial de mantenimiento (filtro `object_id`) |
| `/api/cosateca/maintenance` | POST | Registra mantenimiento |
| `/api/cosateca/contributions` | GET | Aportaciones reconocidas |
| `/api/cosateca/contributions` | POST | Registra una aportación |
| `/api/cosateca/contributions/{id}/credit` | POST | Acredita el valor TQ al contribuyente en el ledger del nodo |

Reglas de negocio: un objeto solo se presta si esta `disponible`; quien lo aporto no puede pedírselo prestado a si mismo; la devolucion pasa el objeto a `disponible`; el credito TQ se abona solo una vez (`credited`) y requiere que la asamblea haya valorado el aporte (`tq_value > 0`).

### Frontend
Ruta `/app/cosateca` (protegida por `PendingAdmissionGuard`, igual que el resto de `/app/*`), con cuatro pestañas:

- **Catalogo**: grid de objetos, estados, filtro por categoria/estado, botón "Prestar" que abre un dialogo para acordar la fecha de devolucion.
- **Mis prestamos**: tabla de prestamos activos del usuario con botón de devolución.
- **Aportar objeto**: formulario para registrar un objeto nuevo (nombre, categoria, estado, ubicacion, referencia al diseno OAD, valor incorporado en TQ).
- **Principios**: recordatorio de las tres capas (objeto / registro / reciprocidad).

Locales: `web/src/locales/es/cosateca.json` y `web/src/locales/en/cosateca.json`; item de navegación `nav.cosateca` en `common.json` (es/en).

## 6. Flujo ejemplo (según el documento adjunto)

1. Ecoaldea B necesita muebles para su escuela comunitaria.
2. Consulta el catalogo (local, y futuro federado) y encuentra diseños de muebles que otra ecoaldea ya desarrollo (`design_ref` en el OAD).
3. Revisa planos y lista de materiales; adapta a la madera de su region.
4. Consulta que herramientas compartidas necesita y organiza el prestamo desde la cosateca.
5. Fabrica los muebles con trabajo comunitario (COS).
6. Registra que materiales funcionaron mejor y cuanto mantenimiento requieren (`cosateca_maintenance`).
7. Comparte las mejoras al repositorio OAD para que otras ecoaldeas las aprovechen.
8. Quien aporto herramientas al fondo colectivo recibe reconocimiento en TQ según las reglas de la asamblea.

## 7. Federación (pendiente de cablear)

Los objetos llevan `node_domain` y `federated_visible`. Hoy el modulo es **local-first**: funciona dentro del nodo sin depender de la red. La sincronización de catálogos de objetos entre nodos federados (gossip, igual que productos y saldos) es el siguiente paso; el esquema ya prevé `federated_visible` para marcar qué objetos se comparten.

## 8. Como probar

1. Arrancar el nodo (Docker o `go run ./cmd/node`): la migración `171` se aplica automáticamente.
2. Iniciar sesión como miembro admitido.
3. Ir a `/app/cosateca` en el panel.
4. Aportar un objeto (ej: "Taladro de columna", categoria herramienta, tq_value 8).
5. Otro miembro solicita prestamo desde el catalogo; el prestamista lo devuelve al terminar.
6. Registrar un mantenimiento y acreditar la aportación en TQ.
