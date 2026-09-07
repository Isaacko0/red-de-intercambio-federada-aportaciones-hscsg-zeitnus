// internal/agent/types.go
// Interfaz del agente inversor — basado en el patrón de reflexión de Dexter
package agent

import (
	"context"
	"fmt"
	"time"
)

// MaxIterations define el máximo de ciclos Think-Act-Reflect
const MaxIterations = 10

// Action representa una decisión propuesta por el agente
type Action struct {
	Type        string         // "analyze", "execute", "report", "halt"
	Description string
	Tool        string         // herramienta a invocar
	Params      map[string]any // parámetros de la herramienta
	Reasoning   string         // justificación (Ley III: lucidez)
	Timestamp   time.Time
}

// ActionResult es el resultado de ejecutar una acción
type ActionResult struct {
	Action      Action
	Success     bool
	Output      any
	Error       error
	Duration    time.Duration
	Timestamp   time.Time
}

// Agent es la interfaz que todo agente especializado debe implementar
type Agent interface {
	// Name devuelve el nombre único del agente
	Name() string

	// Think analiza el estado y propone una acción
	Think(ctx context.Context, state InvestmentState) (*Action, error)

	// Act ejecuta la acción propuesta
	Act(ctx context.Context, action *Action) (*ActionResult, error)

	// Reflect aprende del resultado y ajusta estrategia
	Reflect(ctx context.Context, result *ActionResult) error

	// HasConverged determina si el agente terminó su trabajo
	HasConverged(result *ActionResult) bool
}

// InvestmentState representa el estado actual del fondo de inversión
type InvestmentState struct {
	NAV              float64            // Net Asset Value total
	Cash             float64            // Reserva en stablecoins
	Positions        []Position         // Posiciones abiertas
	PendingActions   []Action           // Acciones pendientes
	LastRebalance    time.Time          // Último rebalanceo
	RiskMetrics      RiskMetrics        // Métricas de riesgo actuales
	MemberCount      int                // Número de miembros activos
	MonthlyInflow    float64            // Aportaciones mensuales (1€ × miembros)
}

// Position representa una inversión activa
type Position struct {
	ID          string
	Asset       string  // "USDC", "RWA_Treasuries", "DeFi_Aave", etc.
	Allocation  float64 // % del NAV
	EntryPrice  float64
	CurrentPrice float64
	PnL         float64 // Profit & Loss
	YieldAPY    float64 // Rendimiento anual estimado
	EntryDate   time.Time
}

// RiskMetrics — métricas de riesgo (Ley I: no dañar base material)
type RiskMetrics struct {
	CVaR95      float64 // Conditional Value at Risk 95%
	SharpeRatio float64
	MaxDrawdown float64
	Volatility  float64 // Volatilidad anualizada
	Exposure    float64 // % del capital expuesto
}

// Validate verifica que el estado es válido (Ley III: lucidez)
func (s *InvestmentState) Validate() error {
	if s.NAV < 0 {
		return fmt.Errorf("NAV no puede ser negativo: %.2f", s.NAV)
	}
	if s.Cash < 0 {
		return fmt.Errorf("reserva de cash no puede ser negativo: %.2f", s.Cash)
	}
	totalAlloc := 0.0
	for _, p := range s.Positions {
		totalAlloc += p.Allocation
	}
	if totalAlloc > 1.0 {
		return fmt.Errorf("asignación total excede 100%%: %.2f", totalAlloc)
	}
	return nil
}
