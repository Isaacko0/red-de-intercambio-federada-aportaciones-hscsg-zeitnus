// internal/skills/risk_metrics/skill.go
// Skill: Métricas de riesgo financiero (Sharpe, Sortino, VaR, CVaR)
package risk_metrics

import (
	"context"
	"fmt"
	"math"
	"sort"
)

// RiskMetrics implementa skills.Skill para cálculo de métricas de riesgo
type RiskMetrics struct{}

func (r *RiskMetrics) Name() string        { return "risk_metrics" }
func (r *RiskMetrics) Description() string { return "Cálculo de métricas de riesgo (Sharpe, Sortino, VaR, CVaR)" }

// Params esperados:
//   - returns []float64 — serie de retornos históricos (diarios o mensuales)
//   - risk_free_rate float64 — tasa libre de riesgo anual (default 4%)
//   - confidence float64 — nivel de confianza para VaR/CVaR (default 0.95)
func (r *RiskMetrics) Execute(ctx context.Context, params map[string]any) (any, error) {
	returns, ok := params["returns"].([]float64)
	if !ok || len(returns) == 0 {
		return nil, fmt.Errorf("parámetro requerido: returns []float64 (no vacío)")
	}

	riskFree, ok := params["risk_free_rate"].(float64)
	if !ok {
		riskFree = 0.04 // 4% anual default
	}

	confidence, ok := params["confidence"].(float64)
	if !ok {
		confidence = 0.95 // 95% default
	}

	// Calcular retorno promedio
	meanReturn := mean(returns)

	// Calcular volatilidad (desviación estándar)
	volatility := stdDev(returns, meanReturn)

	// Sharpe Ratio = (retorno - tasa libre) / volatilidad
	sharpe := (meanReturn - riskFree) / volatility

	// Sortino Ratio = (retorno - tasa libre) / downside deviation
	downsideDev := downsideDeviation(returns, riskFree)
	sortino := 0.0
	if downsideDev > 0 {
		sortino = (meanReturn - riskFree) / downsideDev
	}

	// VaR (Value at Risk) — método histórico
	sorted := make([]float64, len(returns))
	copy(sorted, returns)
	sort.Float64s(sorted)
	varIndex := int(math.Floor((1 - confidence) * float64(len(sorted))))
	valueAtRisk := -sorted[varIndex]

	// CVaR (Conditional VaR) — promedio de pérdidas más allá del VaR
	cvarSum := 0.0
	cvarCount := 0
	for i := 0; i <= varIndex; i++ {
		cvarSum += sorted[i]
		cvarCount++
	}
	conditionalVaR := 0.0
	if cvarCount > 0 {
		conditionalVaR = -cvarSum / float64(cvarCount)
	}

	// Max Drawdown
	maxDrawdown := maxDrawdown(returns)

	return map[string]any{
		"mean_return":      meanReturn,
		"volatility":       volatility,
		"sharpe_ratio":     sharpe,
		"sortino_ratio":    sortino,
		"var_95":           valueAtRisk,
		"cvar_95":          conditionalVaR,
		"max_drawdown":     maxDrawdown,
		"confidence":       confidence,
		"risk_free_rate":   riskFree,
		"sample_size":      len(returns),
	}, nil
}

// Funciones auxiliares

func mean(values []float64) float64 {
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func stdDev(values []float64, meanVal float64) float64 {
	sum := 0.0
	for _, v := range values {
		diff := v - meanVal
		sum += diff * diff
	}
	return math.Sqrt(sum / float64(len(values)))
}

func downsideDeviation(values []float64, target float64) float64 {
	sum := 0.0
	count := 0
	for _, v := range values {
		if v < target {
			diff := v - target
			sum += diff * diff
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return math.Sqrt(sum / float64(count))
}

func maxDrawdown(returns []float64) float64 {
	peak := 1.0
	maxDD := 0.0
	cumulative := 1.0
	for _, r := range returns {
		cumulative *= (1 + r)
		if cumulative > peak {
			peak = cumulative
		}
		dd := (peak - cumulative) / peak
		if dd > maxDD {
			maxDD = dd
		}
	}
	return maxDD
}
