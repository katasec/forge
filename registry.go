package forge

// ToolRegistry stores tools and provides lookup by name.
type ToolRegistry struct {
	tools map[string]Tool
	order []string
}

// NewToolRegistry creates an empty ToolRegistry.
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{
		tools: make(map[string]Tool),
	}
}

// Register adds one or more tools to the registry.
// Duplicate names overwrite silently (last-write-wins).
func (r *ToolRegistry) Register(tools ...Tool) {
	for _, t := range tools {
		name := t.Name()
		if _, exists := r.tools[name]; !exists {
			r.order = append(r.order, name)
		}
		r.tools[name] = t
	}
}

// Get returns a tool by name. Returns false if not found.
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// Definitions returns all registered tools as ToolDefinitions,
// in the order they were first registered.
func (r *ToolRegistry) Definitions() []ToolDefinition {
	defs := make([]ToolDefinition, 0, len(r.order))
	for _, name := range r.order {
		t := r.tools[name]
		defs = append(defs, ToolDefinition{
			Name:        t.Name(),
			Description: t.Description(),
			Schema:      t.Schema(),
		})
	}
	return defs
}
