# Backup: virattt/dexter

> **Fuente**: https://github.com/virattt/dexter  
> **Stack**: TypeScript + Bun + React (Ink CLI) + LangChain  
> **Licencia**: MIT  
> **Stars**: 27.6k ⭐ | **Forks**: 3.4k  
> **Último commit**: 2026-08-04

## ¿Qué es?

Dexter es un **agente autónomo de investigación financiera profunda** — "Claude Code, pero construido específicamente para investigación financiera". Piensa, planifica y aprende mientras trabaja.

## Arquitectura central

```
src/
├── agent/               # Loop del agente (iterativo, con reflexión)
│   ├── agent.ts         # Core: tool-calling loop con max iterations
│   ├── scratchpad.ts    # Fuente de verdad de resultados por query
│   ├── prompts.ts       # System prompts dinámicos
│   ├── run-context.ts   # Contexto de ejecución (estado, historial)
│   └── tool-executor.ts # Ejecución de herramientas con tracking
├── tools/               # Herramientas del agente
│   ├── finance/         # Precios, fundamentals, filings, insider trades
│   ├── search/          # Exa (preferido) + Tavily (fallback)
│   ├── browser/         # Playwright scraping
│   ├── memory/          # Memoria persistente
│   ├── cron/            # Tareas programadas (rebalanceo, alertas)
│   └── subagent/        # Subagentes especializados
├── skills/              # Workflows extensibles (SKILL.md)
│   ├── dcf/             # Valoración flujos de caja descontados
│   ├── x-research/      # Research extendido
│   └── write-memo/      # Memorándums de inversión
├── evals/               # Suite de evaluación (LangSmith)
├── gateway/             # WhatsApp gateway (notificaciones)
├── memory/              # Memoria persistente del agente
├── cron/                # Sistema de cron (tareas programadas)
└── controllers/         # Controladores de flujo
```

## Capacidades clave

| Capacidad | Descripción |
|-----------|-------------|
| **Agent loop iterativo** | Máximo 10 iteraciones, con reflexión entre pasos |
| **Scratchpad** | Registro append-only de todas las herramientas ejecutadas |
| **Multi-provider LLM** | OpenAI, Anthropic, Google, xAI, OpenRouter, Ollama |
| **Herramientas financieras** | Precios, métricas, filings SEC, insider trades, institutional holdings |
| **Web search + browser** | Búsqueda y scraping de páginas financieras |
| **Skills extensibles** | Sistema de plugins SKILL.md (DCF, research, memo) |
| **Evaluación** | Benchmarks con LangSmith, LLM-as-judge |
| **WhatsApp gateway** | Notificaciones en tiempo real |
| **Cron system** | Tareas programadas (rebalanceo, reportes, alertas) |
| **Memoria persistente** | Contexto entre sesiones |

## Lo que es útil para HSCSG/PSG

| Componente | Utilidad |
|------------|----------|
| **Agent loop con reflexión** | **Take**: arquitectura para motor IA inversión — piensa → actúa → reflexiona → repite |
| **Scratchpad append-only** | **Take**: Ley III MJ — auditoría completa de decisiones de inversión |
| **Herramientas financieras** | **Take**: dominios de datos (precios, fundamentals, filings) |
| **Skills (DCF, research)** | **Take**: habilidades de valoración y due diligence |
| **Cron system** | **Take**: rebalanceo automático, distribución periódica, reportes |
| **WhatsApp gateway** | **Take**: notificaciones a miembros del nodo |
| **Evaluación con benchmarks** | **Take**: backtesting de estrategias, stress testing |
| **Multi-provider LLM** | **Adapt**: abstracción para operar con cualquier provider (o local) |

## Lo que se extirpa

- TypeScript/Bun runtime → reescrito en Go
- LangChain dependencias → reemplazado por stdlib Go
- LangSmith tracing → audit trail local (Postwright)
- OpenAI/Anthropic API keys → abstracción multi-provider (o local con Ollama)
- Ink/React CLI → interfaz web (HTML templates del repo) o API REST
- WhatsApp gateway (opcional: mantener como notificación, no core)

---

*Documento generado para asimilación HSCSG v15 OS — Isaac Ko (Isaacko0) | 2026-09-06*
