# Integración: virattt/dexter → red-de-intercambio-federada-isaacko

> **Triple perspectiva**: Usuario / LLM / HSCSG+CaaS  
> **Fecha**: 2026-09-06  
> **Principio anfibio aplicado**: Extirpar TypeScript/Bun/LangChain/Ink; conservar loop de agente con reflexión, scratchpad financiero, skills DCF, cron system.

---

## 1. Perspectiva Usuario

**Qué quiere lograr el nodo:**

Un agente que investigue oportunidades de inversión para el fondo colectivo:
- Analiza activos (RWA, DeFi, equity tokenizado, bienes raíces, energía)
- Propone allocation por franja temporal
- Ejecuta con límites de riesgo predefinidos
- Reporta cada decisión con justificación (Ley III)
- Notifica a miembros vía dashboard (y opcionalmente WhatsApp)
- Rebalancea automáticamente según condiciones de mercado

---

## 2. Perspectiva LLM

**Qué asimilar:**

| Patrón Dexter | Adaptación Go |
|---------------|---------------|
| **Agent loop iterativo** (Think → Act → Reflect) | Paquete `internal/agent/` — interfaz `Agent` con método `Run(ctx, state) error` |
| **Scratchpad append-only** | Paquete `internal/scratchpad/` — log persistente en PostgreSQL (como audit trail de Ley III) |
| **Tool registry** | Paquete `internal/tools/` — registry de herramientas (finance, risk, governance) |
| **Skills (DCF, research, memo)** | Paquete `internal/skills/` — interfaces compiladas (no SKILL.md) |
| **Cron system** (rebalanceo, alertas) | Paquete `internal/cron/` — scheduler con intervalos configurables |
| **Memory** (entre sesiones) | Paquete `internal/memory/` — estado persistido en PostgreSQL |
| **Evals** (benchmarks, backtesting) | Paquete `internal/evals/` — backtesting histórico de estrategias |
| **Multi-provider LLM** | Paquete `internal/llm/` — abstracción con Ollama (local) + OpenRouter (conectado) |

**Qué extirpar:**

- TypeScript/Bun runtime
- LangChain (dependencia externa)
- Ink/React CLI (reemplazado por API REST del repo)
- LangSmith tracing
- Exa/Tavily APIs específicos (reemplazados por fuentes de datos del ecosistema: CoinGecko, DeFiLlama, etc.)

---

## 3. Perspectiva HSCSG + CaaS

| Ley | Mapeo con Dexter |
|-----|------------------|
| **Ley I** | Risk tool verifica exposición — no ejecuta si CVaR > límite. El capital colectivo no se arriesga sin cobertura. |
| **Ley II** | Investment skills generan rendimientos vía DeFi/RWA → distribuidos como CaaS (acceso por contribución AUT). El agente es "trabajador soberano". |
| **Ley III** | Scrapath completo + audit trail en PostgreSQL. Cada decisión es trazable, reproducible, nunca se borra. |

**Vasos comunicantes con HSCSG_v15_OS:**

```
Dexter (patrones)
    ↓
internal/agent/          # Agente inversor con loop reflexivo
    ↓
internal/scratchpad/     # Audit trail append-only (Ley III)
    ↓
internal/skills/         # DCF, risk analysis, yield scanning (compiladas)
    ↓
internal/cron/           # Rebalanceo y distribución periódica
    ↓
internal/llm/            # Abstracción multi-provider (Ollama local / OpenRouter conectado)
    ↓
cmd/investment-agent/    # Entrypoint: `go run ./cmd/investment-agent`
```

---

## 4. Módulos Go a crear

### Paquete: `internal/agent/`

```go
// internal/agent/agent.go
type Agent struct {
    Name        string
    LLM         llm.Provider
    Tools       []tool.Tool
    Skills      []skill.Skill
    Scratchpad  *scratchpad.Scratchpad
    Policy      governance.Policy
}

func (a *Agent) Run(ctx context.Context, state *state.InvestmentState) (*state.InvestmentState, error) {
    // Loop: Think → Act → Reflect (hasta max iter o convergencia)
    for i := 0; i < MaxIterations; i++ {
        action := a.Think(ctx, state)
        result, err := a.Act(ctx, action)
        a.Scratchpad.Record(action, result)
        a.Reflect(ctx, result)
        if a.HasConverged(result) { break }
    }
    return state, nil
}
```

### Paquete: `internal/scratchpad/`

```go
// internal/scratchpad/scratchpad.go
type Scratchpad struct {
    db *sql.DB  // PostgreSQL
}

func (s *Scratchpad) Record(action *agent.Action, result *agent.Result) error {
    // Append-only: INSERT INTO agent_log (timestamp, action_type, args, result, summary)
    // Ley III: nunca UPDATE ni DELETE — auditoría inmutable
}
```

### Paquete: `internal/skills/`

```go
// internal/skills/types.go
type Skill interface {
    Name() string
    Description() string
    Execute(ctx context.Context, params map[string]any) (any, error)
}

// internal/skills/dcf/skill.go
// internal/skills/risk_metrics/skill.go
// internal/skills/yield_scanner/skill.go
```

### Paquete: `internal/cron/`

```go
// internal/cron/scheduler.go
type Scheduler struct {
    tasks []*Task
}

func (s *Scheduler) AddTask(name string, interval time.Duration, fn func()) {
    // Como el cron system de Dexter: at / every / cron expression
}

// Tareas predefinidas:
// - "rebalance": rebalancear portfolio cada 24h
// - "distribute": distribuir rendimientos cada 30 días
// - "report": generar reporte mensual cada 30 días
// - "risk_check": verificar CVaR cada 4 horas
```

### Paquete: `internal/llm/`

```go
// internal/llm/provider.go
type Provider interface {
    Chat(ctx context.Context, messages []Message, tools []ToolDef) (*Response, error)
}

// Implementaciones:
// internal/llm/ollama/provider.go  — local (default, offline)
// internal/llm/openrouter/provider.go — conectado (opcional)
```

### Entrypoint: `cmd/investment-agent/main.go`

```go
// cmd/investment-agent/main.go
// Uso:
//   go run ./cmd/investment-agent                    # corre loop una vez
//   go run ./cmd/investment-agent --cron             # corre como daemon con cron
//   go run ./cmd/investment-agent --backtest 2020-2026 # backtesting histórico
```

---

## 5. Diferencias clave con Dexter original

| Original (Dexter) | Adaptación HSCSG |
|-------------------|------------------|
| TypeScript/Bun runtime | Go 1.25 |
| LangChain | Go estándar (net/http, database/sql) |
| Ink/React CLI | API REST existente + templates HTML |
| LangSmith tracing | Audit trail PostgreSQL (append-only) |
| Exa/Tavily search | DeFiLlama API, CoinGecko, fuentes on-chain |
| OpenAI/Anthropic APIs | Ollama (local) + OpenRouter (conectado) |
| SKILL.md (texto) | Skills compiladas (interfaces Go) |

---

## 6. Verificación de principio anfibio

- ✅ **EXTIRPADO**: TypeScript, Bun, LangChain, Ink, LangSmith, Exa, Tavily, OpenAI/Anthropic SDKs
- ✅ **CONSERVADO**: Agent loop reflexivo, scratchpad, skills DCF, cron system, multi-provider abstraction, evals, memory
- ✅ **ANFIBIO**: El agente opera "desconectado" (Ollama local + backtest) o "conectado" (OpenRouter + ejecución real) según `config.yaml`

---

## 7. Roadmap sugerido

1. **Fase 1**: `internal/agent/` + `internal/scratchpad/` (loop básico, persistencia)
2. **Fase 2**: `internal/skills/` (DCF, risk_metrics, yield_scanner)
3. **Fase 3**: `internal/cron/` (scheduler, rebalanceo, distribución)
4. **Fase 4**: `internal/llm/` (Ollama + OpenRouter)
5. **Fase 5**: `internal/evals/` (backtesting 2020-2026)
6. **Fase 6**: `cmd/investment-agent/` (entrypoint completo)

---

*Documento de integración — Isaac Ko (Isaacko0) | 2026-09-06*
