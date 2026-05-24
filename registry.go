package forge

import "github.com/katasec/forge/tool/registry"

type ToolRegistry = registry.Registry

func NewToolRegistry() *ToolRegistry {
	return registry.New()
}
