// internal/cron/scheduler.go
// Scheduler de tareas periódicas — patrón del cron system de Dexter
package cron

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// TaskFunc es la función que ejecuta una tarea programada
type TaskFunc func(ctx context.Context) error

// Task representa una tarea programada
type Task struct {
	Name     string
	Interval time.Duration
	Fn       TaskFunc
	LastRun  time.Time
	NextRun  time.Time
	Running  bool
	mu       sync.Mutex
}

// Scheduler administra tareas periódicas
type Scheduler struct {
	tasks map[string]*Task
	mu    sync.RWMutex
}

// New crea un nuevo scheduler
func New() *Scheduler {
	return &Scheduler{
		tasks: make(map[string]*Task),
	}
}

// AddTask añade una tarea periódica
func (s *Scheduler) AddTask(name string, interval time.Duration, fn TaskFunc) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.tasks[name]; exists {
		return fmt.Errorf("tarea %s ya existe", name)
	}
	
	s.tasks[name] = &Task{
		Name:     name,
		Interval: interval,
		Fn:       fn,
		NextRun:  time.Now().Add(interval),
	}
	
	log.Printf("[cron] Tarea registrada: %s (cada %v)\n", name, interval)
	return nil
}

// RemoveTask elimina una tarea
func (s *Scheduler) RemoveTask(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if _, exists := s.tasks[name]; !exists {
		return fmt.Errorf("tarea %s no encontrada", name)
	}
	
	delete(s.tasks, name)
	log.Printf("[cron] Tarea eliminada: %s\n", name)
	return nil
}

// ListTasks devuelve las tareas registradas
func (s *Scheduler) ListTasks() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	names := make([]string, 0, len(s.tasks))
	for name := range s.tasks {
		names = append(names, name)
	}
	return names
}

// Run inicia el loop principal del scheduler
func (s *Scheduler) Run(ctx context.Context) {
	log.Println("[cron] Scheduler iniciado")
	
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			log.Println("[cron] Scheduler detenido")
			return
		case <-ticker.C:
			s.checkAndRun()
		}
	}
}

// checkAndRun ejecuta tareas vencidas
func (s *Scheduler) checkAndRun() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	for _, task := range s.tasks {
		task.mu.Lock()
		if !task.Running && now.After(task.NextRun) {
			task.Running = true
			task.LastRun = now
			task.mu.Unlock()
			
			go func(t *Task) {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[cron] PANIC en %s: %v\n", t.Name, r)
					}
					t.mu.Lock()
					t.Running = false
					t.NextRun = time.Now().Add(t.Interval)
					t.mu.Unlock()
				}()
				
				log.Printf("[cron] Ejecutando: %s\n", t.Name)
				if err := t.Fn(context.Background()); err != nil {
					log.Printf("[cron] Error en %s: %v\n", t.Name, err)
				}
			}(task)
		} else {
			task.mu.Unlock()
		}
	}
}

// DefaultTasks devuelve las tareas estándar del fondo PSG
func DefaultTasks() []struct {
	Name     string
	Interval time.Duration
} {
	return []struct {
		Name     string
		Interval time.Duration
	}{
		{"risk_check", 4 * time.Hour},          // Verificar CVaR cada 4h
		{"rebalance", 24 * time.Hour},          // Rebalanceo diario
		{"distribution", 30 * 24 * time.Hour},  // Distribución mensual
		{"report", 30 * 24 * time.Hour},        // Reporte mensual
		{"yield_scan", 12 * time.Hour},         // Escanear yields cada 12h
	}
}
