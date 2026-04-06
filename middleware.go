package forge

import "context"

// RunFunc is the signature for a single provider call, used by middleware.
type RunFunc func(ctx context.Context, req ProviderRequest) (*ProviderResponse, error)

// Middleware wraps a RunFunc to intercept provider calls.
// Composition order: given [A, B, C], request flows A → B → C → provider → C → B → A.
type Middleware func(next RunFunc) RunFunc
