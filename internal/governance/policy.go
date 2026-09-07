// internal/governance/policy.go
// Políticas de gobernanza del motor de inversión — Leyes I/II/III
package governance

import (
	"fmt"
	"time"
)

// Policy define los límites de operación del agente inversor
type Policy struct {
	MaxExposure      float64       // % máximo del NAV en posiciones (Ley I)
	MinLiquidity     float64       // % mínimo en reserva estable (Ley I)
	MaxCVaR          float64       // CVaR máximo aceptable (Ley I)
	SinglePosition   float64       // % máximo en un solo activo
	RebalanceInterval time.Duration // Intervalo de rebalanceo
	StopLoss         float64       // Pérdida máxima antes de cerrar (Ley I)
	RequireApproval  float64       // Monto sobre el cual se requiere aprobación humana (Ley II)
}

// DefaultPolicy devuelve la política por defecto del fondo
func DefaultPolicy() Policy {
	return Policy{
		MaxExposure:      0.80,  // 80% invertido máx, 20% liquidez
		MinLiquidity:     0.20,  // 20% siempre en reserva
		MaxCVaR:          0.05,  // 5% CVaR máx
		SinglePosition:   0.25,  // 25% máx en un solo activo
		RebalanceInterval: 24 * time.Hour,
		StopLoss:         0.10,  // 10% stop-loss
		RequireApproval:  1000.0, // >1000€ requiere votación
	}
}

// Validate verifica que una acción propuesta cumple la política (Ley III: fail-closed)
func (p Policy) Validate(action Action, state InvestmentState) error {
	// Verificar exposición total
	totalExposure := 0.0
	for _, pos := range state.Positions {
		totalExposure += pos.Allocation
	}
	if totalExposure > p.MaxExposure {
		return fmt.Errorf("exposición total %.2f excede máximo %.2f", totalExposure, p.MaxExposure)
	}

	// Verificar posición individual
	for _, pos := range state.Positions {
		if pos.Allocation > p.SinglePosition {
			return fmt.Errorf("posición %s (%.2f) excede máximo individual %.2f", pos.Asset, pos.Allocation, p.SinglePosition)
		}
	}

	// Verificar liquidez mínima
	liquidity := 1.0 - totalExposure
	if liquidity < p.MinLiquidity {
		return fmt.Errorf("liquidez %.2f por debajo del mínimo %.2f", liquidity, p.MinLiquidity)
	}

	// Verificar CVaR
	if state.RiskMetrics.CVaR95 > p.MaxCVaR {
		return fmt.Errorf("CVaR %.4f excede máximo %.4f", state.RiskMetrics.CVaR95, p.MaxCVaR)
	}

	return nil
}

// Action y InvestmentState son réplicas para evitar dependencia cíclica
type Action struct {
	Type   string
	Params map[string]any
}

type InvestmentState struct {
	Positions   []Position
	RiskMetrics RiskMetrics
	NAV         float64
}

type Position struct {
	Asset      string
	Allocation float64
}

type RiskMetrics struct {
	CVaR95 float64
}
