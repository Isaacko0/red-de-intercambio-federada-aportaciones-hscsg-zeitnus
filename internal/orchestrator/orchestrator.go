// internal/orchestrator/orchestrator.go
// Orquestador de agentes — DAG de ejecución con límites de gobernanza
// Basado en orchestrate.py de financial-services
package orchestrator

import (
	"context"
	"fmt"
	"log"
	"time"
)

// Node representa un agente en el DAG de ejecución
type Node struct {
	Agent    AgentInterface
	Name     string
	Deps     []string  // nombres de agentes que deben correr antes
	Executed bool
	Result   *AgentResult
}

// AgentResult encapsula el resultado de un agente
type AgentResult struct {
	AgentName string
	Actions   []*Action
	StartTime time.Time
	EndTime   time.Time
	Error     error
}

// AgentInterface define lo que el orquestador necesita de un agente
type AgentInterface interface {
	Name() string
	Run(ctx context.Context, state map[string]any) ([]*Action, error)
}

// Action representa una acción a ejecutar
type Action struct {
	Type        string
	Description string
	Tool        string
	Params      map[string]any
	Priority    int  // 1=crítica, 2=alta, 3=media, 4=baja
	Timeout     time.Duration
}

// Orchestrator ejecuta agentes respetando dependencias
type Orchestrator struct {
	nodes   map[string]*Node
	history []*AgentResult
}

// New crea un nuevo orquestador
func New() *Orchestrator {
	return &Orchestrator{
		nodes:   make(map[string]*Node),
		history: make([]*AgentResult, 0),
	}
}

// AddNode añade un agente al DAG
func (o *Orchestrator) AddNode(agent AgentInterface, deps []string) {
	o.nodes[agent.Name()] = &Node{
		Agent: agent,
		Name:  agent.Name(),
		Deps:  deps,
	}
}

// Run ejecuta el DAG completo
func (o *Orchestrator) Run(ctx context.Context, initialState map[string]any) error {
	log.Println("=== Orquestador PSG: inicio de ciclo ===")
	
	// Ejecutar en orden topológico simplificado
	executed := make(map[string]bool)
	
	for len(executed) < len(o.nodes) {
		progress := false
		for name, node := range o.nodes {
			if executed[name] {
				continue
			}
			
			// Verificar dependencias
			depsReady := true
			for _, dep := range node.Deps {
				if !executed[dep] {
					depsReady = false
					break
				}
			}
			
			if !depsReady {
				continue
			}
			
			// Ejecutar nodo
			log.Printf("  Ejecutando agente: %s\n", name)
			start := time.Now()
			
			actions, err := node.Agent.Run(ctx, initialState)
			
			result := &AgentResult{
				AgentName: name,
				Actions:   actions,
				StartTime: start,
				EndTime:   time.Now(),
				Error:     err,
			}
			
			o.history = append(o.history, result)
			node.Result = result
			node.Executed = true
			executed[name] = true
			progress = true
			
			if err != nil {
				log.Printf("  ⚠️  Agente %s error: %v\n", name, err)
			} else {
				log.Printf("  ✓ Agente %s completado (%d acciones)\n", name, len(actions))
			}
		}
		
		if !progress {
			return fmt.Errorf("ciclo detectado en el DAG o dependencias faltantes")
		}
	}
	
	log.Println("=== Orquestador PSG: ciclo completado ===")
	return nil
}

// GetHistory devuelve el historial de ejecución
func (o *Orchestrator) GetHistory() []*AgentResult {
	return o.history
}

// GetActions devuelve todas las acciones propuestas por todos los agentes
func (o *Orchestrator) GetActions() []*Action {
	var all []*Action
	for _, result := range o.history {
		all = append(all, result.Actions...)
	}
	return all
}
