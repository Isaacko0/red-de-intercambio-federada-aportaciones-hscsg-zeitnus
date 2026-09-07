// internal/agent/investor.go
// Agente inversor — propone allocation basada en yield/riesgo
package agent

import (
	"context"
	"fmt"
	"time"
)

// InvestorAgent implementa Agent para proponer estrategias de inversión
type InvestorAgent struct {
	llm       LLMProvider
	skills    []SkillExecutor
	scratchpad ScratchpadWriter
}

// NewInvestorAgent crea un nuevo agente inversor
func NewInvestorAgent(llm LLMProvider, skills []SkillExecutor, sp ScratchpadWriter) *InvestorAgent {
	return &InvestorAgent{
		llm:       llm,
		skills:    skills,
		scratchpad: sp,
	}
}

func (a *InvestorAgent) Name() string { return "investor" }

// Think analiza el estado y propone una acción de inversión
func (a *InvestorAgent) Think(ctx context.Context, state InvestmentState) (*Action, error) {
	// Verificar si necesitamos rebalanceo
	if time.Since(state.LastRebalance) > 24*time.Hour {
		return &Action{
			Type:        "analyze",
			Description: "Rebalanceo diario necesario — evaluar allocation actual vs óptimo",
			Tool:        "rebalance_analysis",
			Params: map[string]any{
				"current_positions": state.Positions,
				"risk_metrics":      state.RiskMetrics,
				"nav":               state.NAV,
			},
			Reasoning: fmt.Sprintf("Último rebalanceo: %s (>24h)", state.LastRebalance.Format(time.RFC3339)),
			Timestamp: time.Now(),
		}, nil
	}

	// Verificar si hay oportunidades de yield
	return &Action{
		Type:        "analyze",
		Description: "Escanear oportunidades de yield en DeFi/RWA",
		Tool:        "yield_scan",
		Params: map[string]any{
			"min_apy":    0.03, // 3% mínimo
			"max_risk":   0.05, // 5% CVaR máximo
			"available":  state.Cash,
		},
		Reasoning: fmt.Sprintf("Cash disponible: %.2f — buscar yield > 3%% con CVaR < 5%%", state.Cash),
		Timestamp: time.Now(),
	}, nil
}

// Act ejecuta la acción propuesta
func (a *InvestorAgent) Act(ctx context.Context, action *Action) (*ActionResult, error) {
	start := time.Now()
	
	// Buscar skill correspondiente
	for _, skill := range a.skills {
		if skill.Name() == action.Tool {
			output, err := skill.Execute(ctx, action.Params)
			result := &ActionResult{
				Action:    *action,
				Success:   err == nil,
				Output:    output,
				Error:     err,
				Duration:  time.Since(start),
				Timestamp: time.Now(),
			}
			// Registrar en scratchpad (Ley III: auditoría)
			a.scratchpad.Record(result)
			return result, err
		}
	}

	return &ActionResult{
		Action:    *action,
		Success:   false,
		Error:     fmt.Errorf("skill no encontrado: %s", action.Tool),
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}, nil
}

// Reflect aprende del resultado
func (a *InvestorAgent) Reflect(ctx context.Context, result *ActionResult) error {
	// TODO: ajustar estrategia basado en resultado
	return nil
}

// HasConverged determina si terminó
func (a *InvestorAgent) HasConverged(result *ActionResult) bool {
	return result.Success && result.Action.Type == "report"
}

// Interfaces requeridas (definidas localmente para evitar ciclos)
type LLMProvider interface {
	Chat(ctx context.Context, messages []LLMMessage) (string, error)
}

type SkillExecutor interface {
	Name() string
	Execute(ctx context.Context, params map[string]any) (any, error)
}

type ScratchpadWriter interface {
	Record(result *ActionResult) error
}

type LLMMessage struct {
	Role    string
	Content string
}
