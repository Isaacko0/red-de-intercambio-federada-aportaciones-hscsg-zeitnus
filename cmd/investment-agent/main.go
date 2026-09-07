// cmd/investment-agent/main.go
// Entrypoint del motor de inversión autónomo del fondo PSG
// Uso:
//   go run ./cmd/investment-agent                    # corre loop una vez
//   go run ./cmd/investment-agent --cron             # corre como daemon con cron
//   go run ./cmd/investment-agent --backtest 2020-2026 # backtesting histórico
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"federated-credit-node/internal/agent"
	"federated-credit-node/internal/governance"
	"federated-credit-node/internal/skills"
	"federated-credit-node/internal/skills/dcf"
	"federated-credit-node/internal/skills/risk_metrics"
)

func main() {
	// Flags
	mode := flag.String("mode", "single", "Modo: single | cron | backtest")
	backtestPeriod := flag.String("backtest", "", "Período backtest: YYYY-YYYY")
	flag.Parse()

	log.Println("=== Motor de Inversión PSG ===")
	log.Printf("Modo: %s\n", *mode)

	// Inicializar skills
	registry := skills.NewRegistry()
	registry.Register(&dcf.DCF{})
	registry.Register(&risk_metrics.RiskMetrics{})

	// Política de gobernanza (Leyes I/II/III)
	policy := governance.DefaultPolicy()
	log.Printf("Política: exposición máx %.0f%%, liquidez mín %.0f%%, CVaR máx %.2f\n",
		policy.MaxExposure*100, policy.MinLiquidity*100, policy.MaxCVaR)

	// Estado inicial (seed)
	state := agent.InvestmentState{
		NAV:           1000.0,  // NAV inicial 1000€
		Cash:          1000.0,  // Todo en cash al inicio
		Positions:     []agent.Position{},
		LastRebalance: time.Now(),
		MemberCount:   100,     // 100 miembros iniciales
		MonthlyInflow: 100.0,   // 100 × 1€/mes
	}

	log.Printf("Estado inicial: NAV=%.2f€, Cash=%.2f€, Miembros=%d\n",
		state.NAV, state.Cash, state.MemberCount)

	// Ejecutar según modo
	switch *mode {
	case "single":
		runSingleShot(state, registry, policy)
	case "cron":
		runCronDaemon(state, registry, policy)
	case "backtest":
		if *backtestPeriod == "" {
			log.Fatal("--backtest requiere período (ej: --backtest 2020-2026)")
		}
		runBacktest(*backtestPeriod, registry, policy)
	default:
		log.Fatalf("Modo desconocido: %s", *mode)
	}
}

// runSingleShot ejecuta un ciclo de inversión
func runSingleShot(state agent.InvestmentState, registry *skills.Registry, policy governance.Policy) {
	log.Println("\n--- Ciclo único de inversión ---")
	ctx := context.Background()

	// Crear agente inversor (con LLM stub — sin conexión externa)
	investor := agent.NewInvestorAgent(nil, registry.List(), nil)

	// Think
	action, err := investor.Think(ctx, state)
	if err != nil {
		log.Fatalf("Error en Think: %v", err)
	}
	log.Printf("Acción propuesta: %s — %s\n", action.Type, action.Description)
	log.Printf("Razonamiento: %s\n", action.Reasoning)

	// Validar contra política (Ley III: fail-closed)
	if err := policy.Validate(governance.Action{
		Type:   action.Type,
		Params: action.Params,
	}, governance.InvestmentState{
		Positions: convertToGovPositions(state.Positions),
		RiskMetrics: governance.RiskMetrics{
			CVaR95: state.RiskMetrics.CVaR95,
		},
		NAV: state.NAV,
	}); err != nil {
		log.Printf("⚠️  Acción RECHAZADA por política: %v\n", err)
		os.Exit(0)
	}

	// Act
	result, err := investor.Act(ctx, action)
	if err != nil {
		log.Fatalf("Error en Act: %v", err)
	}
	log.Printf("Resultado: success=%v, duration=%v\n", result.Success, result.Duration)

	// Reflect
	if err := investor.Reflect(ctx, result); err != nil {
		log.Printf("Error en Reflect: %v\n", err)
	}

	log.Println("\n✓ Ciclo completado")
}

// runCronDaemon ejecuta el agente como daemon con tareas programadas
func runCronDaemon(state agent.InvestmentState, registry *skills.Registry, policy governance.Policy) {
	log.Println("\n--- Iniciando daemon de inversión ---")
	log.Println("Tareas programadas:")
	log.Println("  - risk_check: cada 4 horas")
	log.Println("  - rebalance: cada 24 horas")
	log.Println("  - distribution: cada 30 días")
	log.Println("  - report: cada 30 días")
	log.Println("\n(Presiona Ctrl+C para detener)")

	// Loop principal (simplificado — en producción usar internal/cron/)
	ticker := time.NewTicker(4 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("\n[%s] Ejecutando verificación de riesgo...\n", time.Now().Format(time.RFC3339))
		// TODO: implementar risk check
	}
}

// runBacktest ejecuta backtesting histórico
func runBacktest(period string, registry *skills.Registry, policy governance.Policy) {
	log.Printf("\n--- Backtesting %s ---\n", period)
	log.Println("Cargando datos históricos...")
	log.Println("Ejecutando estrategia...")
	log.Println("Generando reporte...")
	// TODO: implementar backtesting real
	log.Println("\n✓ Backtesting completado (stub)")
}

// convertToGovPositions convierte posiciones al formato de governance
func convertToGovPositions(positions []agent.Position) []governance.Position {
	result := make([]governance.Position, len(positions))
	for i, p := range positions {
		result[i] = governance.Position{
			Asset:      p.Asset,
			Allocation: p.Allocation,
		}
	}
	return result
}
