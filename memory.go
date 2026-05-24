package forge

import (
	memorypkg "github.com/katasec/forge/memory"
	"github.com/katasec/forge/memory/inmem"
)

type MemoryStore = memorypkg.Store
type InMemoryStore = inmem.Store

func NewInMemoryStore() *InMemoryStore {
	return inmem.New()
}
