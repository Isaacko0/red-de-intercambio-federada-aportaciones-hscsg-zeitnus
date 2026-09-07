// internal/scratchpad/scratchpad.go
// Scratchpad append-only — Ley III MJ: auditoría inmutable de decisiones
// Basado en el scratchpad de Dexter
package scratchpad

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry representa una entrada en el log de auditoría
type Entry struct {
	Timestamp   time.Time       `json:"timestamp"`
	Type        string          `json:"type"`        // "action", "result", "error"
	Action      string          `json:"action"`      // nombre de la acción
	Description string          `json:"description"` // descripción legible
	Raw         json.RawMessage `json:"raw"`         // datos completos
	Agent       string          `json:"agent"`       // agente que la ejecutó
	Success     bool            `json:"success"`
}

// Scratchpad mantiene el log append-only
type Scratchpad struct {
	mu      sync.Mutex
	entries []Entry
	file    *os.File
	path    string
}

// New crea un nuevo scratchpad
func New(dataDir string) (*Scratchpad, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("creando directorio: %w", err)
	}

	path := filepath.Join(dataDir, fmt.Sprintf("investment_agent_%s.jsonl", time.Now().Format("20260102")))
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("abriendo archivo: %w", err)
	}

	return &Scratchpad{
		entries: make([]Entry, 0),
		file:    file,
		path:    path,
	}, nil
}

// Record añade una entrada al scratchpad (append-only, nunca se borra)
func (s *Scratchpad) Record(entryType, action, description, agent string, raw any, success bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	rawJSON, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("serializando raw: %w", err)
	}

	entry := Entry{
		Timestamp:   time.Now(),
		Type:        entryType,
		Action:      action,
		Description: description,
		Raw:         rawJSON,
		Agent:       agent,
		Success:     success,
	}

	// Escribir a disco inmediatamente (durabilidad)
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("serializando entry: %w", err)
	}

	if _, err := s.file.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("escribiendo a disco: %w", err)
	}

	s.entries = append(s.entries, entry)
	return nil
}

// GetEntries devuelve todas las entradas (solo lectura)
func (s *Scratchpad) GetEntries() []Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]Entry, len(s.entries))
	copy(result, s.entries)
	return result
}

// Close cierra el archivo
func (s *Scratchpad) Close() error {
	return s.file.Close()
}

// Path devuelve la ruta del archivo
func (s *Scratchpad) Path() string {
	return s.path
}
