# Integración: Cosateca/Objetoteca con solarpunk_utopia y solarpunk_utopia_project

**Repositorio integrador:** https://github.com/Isaacko0/red-de-intercambio-federada-aportaciones-hscsg-zeitnus
**Repositorios integrados:**
- https://github.com/lizTheDeveloper/solarpunk_utopia (Python FastAPI, gift economy mesh network)
- https://github.com/lizTheDeveloper/solarpunk_utopia_project (TypeScript, OpenSpec specs)

## Alcance de la integución

Las pilas tecnológicas son distintas (Go + PostgreSQL vs Python FastAPI + IndexedDB), por lo que **no se fusiona código fuente directamente**. La integración se da a nivel de:

1. **Specs OpenSpec asimilados** (docs/openspec/): resource-sharing, time-bank, community-care de solarpunk_utopia_project
2. **Endpoints compatibles** en la API Go de cosateca (/api/cosateca/shared-resources/*)
3. **Documentación de mapeo** entre TQ y gift economy

## Endpoints de compatibilidad con Resource Sharing

La cosateca expone los objetos como "shared resources" bajo `/api/cosateca/shared-resources/*`:

| Método | Endpoint | Mapea a spec |
|---|---|---|
| GET | `/api/cosateca/shared-resources` | REQ-SHARE-001, REQ-SHARE-002 (listar recursos) |
| GET | `/api/cosateca/shared-resources/{id}` | REQ-SHARE-001 (detalle de recurso) |
| POST | `/api/cosateca/shared-resources/{id}/reserve` | REQ-SHARE-013 (reservar para recogida) |
| POST | `/api/cosateca/shared-resources/{id}/return` | REQ-SHARE-013 (devolver) |
| GET | `/api/cosateca/shared-resources/{id}/maintenance` | REQ-SHARE-006, REQ-SHARE-008 (ciclo de vida) |

Los endpoints existentes de cosateca (`/api/cosateca/objects`, `/api/cosateca/loans`, `/api/cosateca/maintenance`) ya cubren estos flujos; los endpoints `/shared-resources/*` son alias con formato de respuesta adaptado al spec.

## Federación entre nodos

- **Este nodo (Go):** Federación vía mTLS + gossip (protocolo propio de red-de-intercambio-federada)
- **solarpunk_utopia (Python):** Federación vía DTN bundles + NATS + mesh networking
- **Puente:** Pendiente. Los objetos de la cosateca llevan `node_domain` y `federated_visible` para futura sincronización con nodos solarpunk_utopia vía DTN bundles.

## Cómo probar la integración

1. Arrancar el nodo Go: `go run ./cmd/node` (migración 171 se aplica automáticamente)
2. Iniciar sesión como miembro admitido
3. Ir a `/app/cosateca` y aportar objetos
4. Consultar `GET /api/cosateca/shared-resources` para ver recursos en formato compatible con spec
5. Los specs de referencia están en `docs/openspec/`

## Limitaciones

- No se fusionó código Python en el repo Go (pilas incompatibles)
- La federación DTN/NATS con solarpunk_utopia está pendiente (tarea INTEGRACION_HSCSG-cosateca-federación en ridf-orchestrator-state.json)
- Los specs asimilados son referencia de diseño, no código ejecutable
