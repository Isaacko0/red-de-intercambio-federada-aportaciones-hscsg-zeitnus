# Integración: anthropics/financial-services → red-de-intercambio-federada-isaacko

> **Triple perspectiva**: Usuario / LLM / HSCSG+CaaS  
> **Fecha**: 2026-09-06  
> **Principio anfibio aplicado**: Extirpar infraestructura Claude/Anthropic + LangSmith; conservar patrones de agentes financieros especializados y arquitectura de skills.

---

## 1. Perspectiva Usuario

**Qué quiere lograr el nodo con esto:**

El nodo (ecosaldea/makerspace) necesita un **motor de inversión autónomo** que:
- Genere rendimientos con el capital colectivo (1€/mes × miembros)
- Se gobernar socialmente (no es un bot ciego)
- Opera en múltiples franjas temporales (corto/mediano/largo plazo = 1/10/100 años)
- Reporta transparentemente (Ley III MJ — lucidez, nunca engañar)

**Lo que NO necesita:**
- Integrarse a Claude Code/Anthropic
- Usar Microsoft 365 ni LangSmith
- Servicios cloud propietarios

---

## 2. Perspectiva LLM

**Qué asimilar (lógica pura / patrones):**

| Patrón | Cómo se asimila en Go |
|--------|----------------------|
| **Arquitectura multi-agente financiera** (10 agentes colaborando) | Agentes por dominio: `internal/agents/investor/`, `internal/agents/risk/`, `internal/agents/auditor/` — cada uno con interfaz `Agent` (Think, Act, Reflect) |
| **Skills de due diligence** (KYC screener, valuation reviewer) | Paquete `internal/skills/` con interfaces `Skill` y registry. Skills compiladas, no interpretadas (Go). |
| **Steering de agentes** (límites, fail-closed) | Paquete `internal/governance/` — políticas de acción, circuit breakers, spending limits |
| **Evaluación de agentes** (LangSmith → benchmarks offline) | Paquete `internal/evals/` — backtesting de estrategias históricas, stress testing |
| **Orquestación multi-paso** (orchestrate.py) | Paquete `internal/orchestrator/` — DAG de agentes, orden ejecución, manejo de dependencias |

**Qué extirpar (infra ajena):**

- Claude Code CLI y plugin.json → no relevante
- MCP servers externos → no relevante
- LangSmith tracing → no relevante
- System prompts como strings → convertidos a estructuras Go (`AgentProfile`, `SkillSpec`)
- Python dependencies → reescrito en Go estándar

---

## 3. Perspectiva HSCSG + CaaS

**Isomorfismo con Leyes del Materialismo Jerárquico:**

| Ley | Mapeo con financial-services |
|-----|------------------------------|
| **Ley I (no dañar base material)** | Risk Agent verifica exposición antes de ejecutar — el capital no se arriesga sin cobertura. Patrón KYC/compliance → verificar que la inversión no daña la base productiva de la ecoaldea. |
| **Ley II (ganarse la vida soberanizando)** | Investment Agent genera rendimientos que se distribuyen vía CaaS — acceso por contribución (AUT), no por capital. El motor IA es un "trabajador soberano" del nodo. |
| **Ley III (lucidez, nunca engañar)** | Audit Agent + Reporter Agent publican métricas on-chain (o en dashboard transparente). Toda decisión de inversión es trazable, append-only, auditabilidad completa. |

**Vasos comunicantes con HSCSG_v15_OS:**

```
financial-services (patrones)
    ↓
internal/agents/          # Agentes especializados Go
    ↓
internal/orchestrator/    # DAG de agentes (como orchestrate.py)
    ↓
internal/governance/      # Ley III: políticas, límites, audit trail
    ↓
internal/skills/          # Skills compiladas (DCF, risk analysis, etc.)
    ↓
cmd/investment-agent/     # Entrypoint: `go run ./cmd/investment-agent`
    ↓
state/investment.go       # Estado persistido (Zustand en HSCSG → PostgreSQL acá)
```

---

## 4. Módulos Go a crear

### Paquete: `internal/agents/`

```go
// internal/agents/types.go
type Agent interface {
    Name() string
    Think(ctx context.Context, state *InvestmentState) (*Action, error)
    Reflect(ctx context.Context, result *ActionResult) error
}

// internal/agents/investor/agent.go — busca yield, proponeallocation
// internal/agents/risk/agent.go   — evalúa exposición, CVaR, stress test
// internal/agents/audit/agent.go  — verifica compliance, genera reportes
```

### Paquete: `internal/skills/`

```go
// internal/skills/types.go
type Skill interface {
    Name() string
    Execute(ctx context.Context, params map[string]any) (any, error)
}

// internal/skills/dcf/skill.go — valoración flujos de caja descontados
// internal/skills/risk_analysis/skill.go — métricas de riesgo (Sharpe, Sortino, VaR)
// internal/skills/yield_scan/skill.go — escaneo APY DeFi/RWA
```

### Paquete: `internal/orchestrator/`

```go
// internal/orchestrator/dag.go
type Orchestrator struct {
    agents []Agent
    deps   map[string][]string  // orden de ejecución
    limits GovernancePolicy
}
```

### Paquete: `internal/governance/`

```go
// internal/governance/policy.go
type GovernancePolicy struct {
    MaxExposure    float64  // % máximo del capital en una posición
    MinLiquidity   float64  // % mínimo en reserva estable
    CircuitBreaker func(state *InvestmentState) bool
}
```

### Entrypoint: `cmd/investment-agent/main.go`

```go
// cmd/investment-agent/main.go
// `go run ./cmd/investment-agent` → carga estado, corre orquestador, persiste resultado
```

---

## 5. Diferencias clave con financial-services original

| Original (Claude/Anthropic) | Adaptación HSCSG |
|------------------------------|------------------|
| System prompts como markdown | `AgentProfile` struct en Go |
| MCP servers (externos) | Skills compiladas (paquete `internal/skills/`) |
| LangSmith tracing (cloud) | Audit trail local (PostgreSQL, append-only) |
| Subagent orchestration (YAML) | DAG en Go (`internal/orchestrator/`) |
| API keys (env vars) | Config local (`config.yaml` existente) |

---

## 6. Verificación de principio anfibio

- ✅ **EXTIRPADO**: Claude Code CLI, Anthropic SDK, LangSmith, MCP servers externos, Microsoft 365
- ✅ **CONSERVADO**: Arquitectura multi-agente, patrones financieros, evaluación, orquestación, steering, fail-closed
- ✅ **ANFIBIO**: El motor opera "desconectado" (benchmark/backtest local) o "conectado" (ejecución real vía APIs DeFi) según `config.yaml` (como `priceParity` en HSCSG)

---

*Documento de integración — Isaac Ko (Isaacko0) | 2026-09-06*
