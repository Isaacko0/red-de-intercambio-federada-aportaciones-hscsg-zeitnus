// internal/skills/types.go
// Sistema de skills compiladas (no SKILL.md) — patrón de Dexter adaptado a Go
package skills

import "context"

// Skill es la interfaz que toda habilidad de inversión debe implementar
type Skill interface {
	Name() string
	Description() string
	Execute(ctx context.Context, params map[string]any) (any, error)
}

// Registry mantiene el registro de skills disponibles
type Registry struct {
	skills map[string]Skill
}

// NewRegistry crea un nuevo registro de skills
func NewRegistry() *Registry {
	return &Registry{skills: make(map[string]Skill)}
}

// Register añade un skill al registro
func (r *Registry) Register(skill Skill) {
	r.skills[skill.Name()] = skill
}

// Get devuelve un skill por nombre
func (r *Registry) Get(name string) (Skill, bool) {
	s, ok := r.skills[name]
	return s, ok
}

// List devuelve todos los skills registrados
func (r *Registry) List() []Skill {
	result := make([]Skill, 0, len(r.skills))
	for _, s := range r.skills {
		result = append(result, s)
	}
	return result
}
