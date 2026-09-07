// internal/skills/dcf/skill.go
// Skill: Valoración por Flujos de Caja Descontados (DCF)
package dcf

import (
	"context"
	"fmt"
	"math"
)

// DCF implementa skills.Skill para valoración DCF
type DCF struct{}

func (d *DCF) Name() string        { return "dcf_valuation" }
func (d *DCF) Description() string { return "Valoración por Flujos de Caja Descontados" }

// Params esperados:
//   - cash_flows []float64 — flujos de caja proyectados (5-10 años)
//   - discount_rate float64 — tasa de descuento (WACC)
//   - terminal_growth float64 — crecimiento terminal (default 2%)
//   - years int — años de proyección (default 5)
func (d *DCF) Execute(ctx context.Context, params map[string]any) (any, error) {
	cashFlows, ok := params["cash_flows"].([]float64)
	if !ok {
		return nil, fmt.Errorf("parámetro requerido: cash_flows []float64")
	}
	
	discountRate, ok := params["discount_rate"].(float64)
	if !ok {
		discountRate = 0.10 // 10% default
	}
	
	terminalGrowth, ok := params["terminal_growth"].(float64)
	if !ok {
		terminalGrowth = 0.02 // 2% default
	}

	// Calcular valor presente de cada flujo
	presentValue := 0.0
	for i, cf := range cashFlows {
		discountFactor := math.Pow(1+discountRate, float64(i+1))
		presentValue += cf / discountFactor
	}

	// Valor terminal (Gordon Growth Model)
	lastCF := cashFlows[len(cashFlows)-1]
	terminalValue := (lastCF * (1 + terminalGrowth)) / (discountRate - terminalGrowth)
	terminalPV := terminalValue / math.Pow(1+discountRate, float64(len(cashFlows)))

	totalValue := presentValue + terminalPV

	return map[string]any{
		"present_value_flows":  presentValue,
		"terminal_value":       terminalValue,
		"present_value_terminal": terminalPV,
		"total_value":          totalValue,
		"discount_rate":        discountRate,
		"terminal_growth":      terminalGrowth,
	}, nil
}
