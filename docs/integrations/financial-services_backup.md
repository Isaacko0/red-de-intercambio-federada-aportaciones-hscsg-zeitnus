# Backup: anthropics/financial-services

> **Fuente**: https://github.com/anthropics/financial-services  
> **Stack**: Python + Markdown + YAML (plugins Claude Code/Managed Agents)  
> **Licencia**: Apache 2.0  
> **Stars**: 34.7k ⭐ | **Forks**: 5.2k  
> **Último commit**: 2026-08-25

## ¿Qué es?

Repositorio oficial de Anthropic con plugins y templates de agentes gestionados para servicios financieros. Diseñado para Claude Code y Claude Managed Agents (CMA), contiene:

- **10 agentes nombrados** autónomos (Pitch Agent, Market Researcher, Earnings Reviewer, Model Builder, Meeting Prep, GL Reconciler, Month-End Closer, Statement Auditor, Valuation Reviewer, KYC Screener)
- **Skills verticales**: client-report, investment-proposal, tax-loss-harvesting, y más
- **Cookbooks de agentes gestionados**: configuración de subagentes, steering, seguridad
- **Scripts de despliegue**: check.py, sync-agent-skills.py, deploy-managed-agent.sh, orchestrate.py

## Estructura relevante

```
plugins/
├── agent-plugins/          # Agentes autónomos (cada uno con system prompt + skills)
│   └── <slug>/
│       ├── agents/<slug>.md      # System prompt canónico
│       └── skills/               # Skills empaquetados
├── vertical-plugins/       # Fuentes de skills por vertical financiera
│   └── <vertical>/
│       ├── skills/               # SKILL.md files
│       ├── commands/             # Comandos slash
│       └── .mcp.json             # Configuración MCP
└── partner-built/           # Plugins de terceros (LSEG, S&P Global)
scripts/
├── check.py                 # Lint de manifests + referencias
├── sync-agent-skills.py     # Sincroniza skills vertical → agent bundles
├── deploy-managed-agent.sh  # Despliegue de agentes gestionados
└── orchestrate.py           # Loop de orquestación de referencia
```

## Lo que es útil para HSCSG/PSG

| Componente | Utilidad |
|------------|----------|
| Arquitectura multi-agente financiera (agentes especializados colaborando) | **Take**: modelo para motor IA inversión con agentes por franja temporal |
| Skills de due diligence, valoración, compliance KYC, earnings review, tax optimization | **Take**: dominios financieros para entrenar agentes de inversión |
| Patrón "managed agent" con subagentes y steering | **Take**: gobernanza de agentes autónomos con límites de seguridad |
| Evaluación de agentes (LangSmith integration) | **Take**: metodología de evaluación continua |
| Scripts orquestación y sincronización | **Adapt**: patrones para orquestador HSCSG |

## Lo que se extirpa

- Dependencias de Claude/Anthropic específicos (MCP servers, Claude Code CLI)
- Integraciones con Microsoft 365 (claude-for-msft-365-install)
- LangSmith tracing (servicio cloud)
- API keys y secrets de terceros

---

*Documento generado para asimilación HSCSG v15 OS — Isaac Ko (Isaacko0) | 2026-09-06*
