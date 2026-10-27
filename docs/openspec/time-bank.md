# Time Bank Specification (asimilado de solarpunk_utopia_project)

**Origen:** https://github.com/lizTheDeveloper/solarpunk_utopia_project/specs/platform/time-bank.md
**Estado:** Referencia para la contabilidad de contribuciones TQ del nodo federado
**Integración:** El módulo Cosateca usa TQ para reconocer aportaciones al fondo colectivo de objetos

## Resumen

El Time Bank permite a miembros de la comunidad ofertar e intercambiar tiempo, habilidades y trabajo en un marco de ayuda mutua no monetario. Enfatiza principios de economía de regalo: contribuir según habilidad, recibir según necesidad, sin contabilidad estrictamente transaccional ni deuda.

## Mapeo a la implementación Go

| Spec REQ-TIME | Implementación en cosateca/TQ |
|---|---|
| REQ-TIME-001: Compartir tiempo basado en regalo | Aportaciones de objetos al fondo colectivo sin expectativa de reciprocidad directa |
| REQ-TIME-002: Seguimiento de abundancia sobre deuda | `cosateca_contributions` rastrea aportaciones sin crear deuda entre miembros |
| REQ-TIME-003: Ofertas de habilidad y servicio | Quien aporta un objeto aporta su habilidad para adquirirlo/mantenerlo |
| REQ-TIME-006: Solicitudes explícitas de tiempo | Préstamos de objetos como solicitud de uso de recurso compartido |
| REQ-TIME-012: Matching impulsado por IA | Pendiente (FRS del nodo podría hacer matching de objetos con necesidades) |
| REQ-TIME-014: Coordinación geográfica | Objetos filtrados por `node_domain` (nodo local) |
| REQ-TIME-019: Fomento de participación | Reconocimiento en TQ por aportaciones al fondo colectivo |
| REQ-TIME-020: Identificar brechas de habilidad | Mantenimiento y estado de objetos revelan necesidades de capacitación |

## TQ vs Time Bank

El nodo opera con **TQ** (1 TQ = 1 kWh, cinta métrica de reciprocidad del Libro Ecoaldeas Federadas), no con un time bank tradicional. Las diferencias clave:

- **TQ** está anclado a energía física (1 kWh = 3.6 MJ), no a horas de tiempo
- **TQ** tiene límites simétricos (-500/+500) para evitar acumulación
- **TQ** no se puede intercambiar por fiat/cripto (regla de la aldea)
- **TQ** se extingue al acceder a bienes (no se acumula indefinidamente)

Las aportaciones a la cosateca se reconocen en TQ por valor de la asamblea, siguiendo REQ-TIME-002 (seguimiento sin deuda entre miembros: el crédito va al contribuyente, la contrapartida es el puente nodal, no otro miembro).
