# Skills desplegadas en `red-de-intercambio-federada-aportaciones-hscsg-zeitnus`

Índice de skills Hermes portadas y adaptadas desde `Marca-Blanca-Zeitnus-Firma-Operaciones-Ecotomica-Networking-Recreativo`.

Cada skill está en `skills/<nombre>/SKILL.md` (y, para uso de Hermes local, espejo bajo `.hermes/skills/<nombre>/`).

## Skills de asimilación (core)

| Skill | Ruta | Propósito |
|-------|------|-----------|
| `hscsg-repo-assimilation` | `skills/hscsg-repo-assimilation/SKILL.md` | Metodología de 4 fases para asimilar repos externos como módulos vivos del nodo federado (Go). |
| `hscsg-unified-assimilation-science` | `skills/hscsg-unified-assimilation-science/SKILL.md` | Fusión asimilación + ciencia (EBD, vasos comunicantes, reproducibilidad). |
| `hscsg-fractal-commit-fusion` | `skills/hscsg-fractal-commit-fusion/SKILL.md` | Fusión fractal de commits con trazabilidad semántica. |
| `gaia-alrac-assimilation` | `.hermes/skills/gaia-alrac-assimilation/SKILL.md` | Asimilación de documentos maestros (Gaia) al arquitectura ALRAC de 5 capas + auditoría de licencias. |

## Skills de gobierno y orquestación

| Skill | Ruta | Propósito |
|-------|------|-----------|
| `hscsg-orquestador-skills` | `skills/hscsg-orquestador-skills/SKILL.md` | Router maestro entre skills; vasos comunicantes obligatorios; enrutamiento de intención. |
| `hscsg-orchestrator` | `skills/hscsg-orchestrator/SKILL.md` | Ciclos Alráico, gobierno del pipeline de asimilación, guarda `orchestrator-state.json`. |

## Skills de propósito operativo

| Skill | Ruta | Propósito |
|-------|------|-----------|
| `hscsg-nfc-wallet` | `skills/hscsg-nfc-wallet/SKILL.md` | Terminales NFC físicos (ESP32 + PN532), 4 modos, criptografía ed25519/AES-256-GCM/ECDH, offline-first. Relevante para `firmware/` del nodo. |
| `hscsg-monetary-integration` | `skills/hscsg-monetary-integration/SKILL.md` | Integración de sistemas monetarios alternativos (G1, Túmin, PAR, Bitcoin, stablecoins) a arquitectura ZCS/ZNU. |
| `hscsg-document-architect` | `skills/hscsg-document-architect/SKILL.md` | Documentos unificados de estrategia/modelo civilizatorio; integración de frameworks externos. |
| `hscsg-scientific-papers` | `skills/hscsg-scientific-papers/SKILL.md` | Protocolo de papers científicos y documentación técnica con metodología EBD y trazabilidad. |
| `brief-detector-recommender` | `skills/brief-detector-recommender/SKILL.md` | Detección de gaps, recomendación de briefs, proyección y extrapolación de patrones de asimilación. |

## Principio de trabajo (aplicar siempre)

- **Idioma español**, respuestas concisas.
- **Respaldo primero** (Fase 0) antes de tocar el repo.
- **Licencia LPF-1.0** del nodo federado: copyleft fuerte, federación obligatoria, registro de forks en `forks-registry/registry.md`, no relicenciamiento.
- **Principio Anfibio:** misma lógica en modo postmonetario (ZNU/CaaS offline) o conectado (USD/USDC vía oráculo); el render decide la etiqueta, la lógica es agnóstica a la unidad.
- **Vasos comunicantes:** toda fuente externa se asimila como módulo vivo Y documento trazable; cada claim conecta evidencia → decisión → código → test → doc.
- **Verificación:** `go build ./...` y `go test ./...` con 0 errores antes de entregar.

---\n*Portado y adaptado de `Marca-Blanca-Zeitnus-Firma-Operaciones-Ecotomica-Networking-Recreativo` → `red-de-intercambio-federada-aportaciones-hscsg-zeitnus`.*
