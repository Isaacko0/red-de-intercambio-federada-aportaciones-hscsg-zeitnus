# Resource Sharing Specification (asimilado de solarpunk_utopia_project)

**Origen:** https://github.com/lizTheDeveloper/solarpunk_utopia_project/specs/platform/resource-sharing.md
**Estado:** Implementado parcialmente en el módulo Cosateca/Objetoteca del nodo federado (red-de-intercambio-federada-aportaciones-hscsg-zeitnus)
**Integración:** La cosateca expone endpoints compatibles con este spec bajo `/api/cosateca/shared-resources/*`

## Resumen

El sistema de intercambio de recursos permite a miembros de la comunidad ofertar y descubrir recursos compartidos: objetos físicos, herramientas, equipos, espacios, y tecnologías de fabricación. Implementa el enfoque de economía de regalo, buy-nothing y bienes comunes para la gestión comunitaria de recursos.

La cosateca de este nodo implementa los requisitos REQ-SHARE-001, REQ-SHARE-002, REQ-SHARE-006, REQ-SHARE-008, REQ-SHARE-011, REQ-SHARE-012, REQ-SHARE-013, REQ-SHARE-014, REQ-SHARE-018 mediante su módulo de objetos, préstamos y mantenimiento.

## Mapeo a la implementación Go

| Spec REQ-SHARE | Implementación en cosateca | Endpoint |
|---|---|---|
| REQ-SHARE-001: Objetos físicos (buy-nothing) | `cosateca_objects` con categoría, condición, ubicación | `GET /api/cosateca/objects` |
| REQ-SHARE-002: Herramientas y equipos | Categoría `herramienta`, `maquina`, `equipo_eventos` | `GET /api/cosateca/objects?category=herramienta` |
| REQ-SHARE-006: Ciclo de vida del recurso | `cosateca_objects.status` + `cosateca_maintenance` | `GET /api/cosateca/maintenance?object_id=...` |
| REQ-SHARE-008: Condición y calidad | Campo `condition` (bueno/regular/malo) + mantenimiento con `outcome` | `PATCH /api/cosateca/objects/{id}` + `POST /api/cosateca/maintenance` |
| REQ-SHARE-011: Proximidad geográfica | Campo `location` + filtrado por nodo | `GET /api/cosateca/objects?location=...` |
| REQ-SHARE-012: Calendarios de disponibilidad | `cosateca_loans` con `due_at` y `state` | `GET /api/cosateca/loans` |
| REQ-SHARE-013: Recogida y entrega | Préstamo con `due_at` acordada | `POST /api/cosateca/loans` + `POST /api/cosateca/loans/{id}/return` |
| REQ-SHARE-014: Multiparty | Múltiples préstamos secuenciales al mismo objeto | `cosateca_loans` historial |
| REQ-SHARE-018: Acceso por habilidad | `design_ref` + documentación en `bill_of_materials` | `GET /api/cosateca/objects/{id}` |

## Diferencias filosóficas

La cosateca de este nodo opera con **TQ** (1 TQ = 1 kWh, crédito de contribución reconocido por la asamblea), mientras que este spec enfatiza una **economía de regalo sin deuda**. Los dos modelos son compatibles: el nodo puede operar en modo "regalo puro" (access_tq_cost = 0) o en modo "contribución reconocida" (access_tq_cost > 0, acreditado en TQ por la asamblea). No se trata como dinero ni se permite intercambio por fiat/cripto.

## Endpoints adicionales (compatibilidad con spec)

Ver `internal/api/cosateca.go` para los endpoints `/api/cosateca/shared-resources/*` que exponen los objetos en formato compatible con Resource Sharing.
