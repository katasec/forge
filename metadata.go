package forge

import "context"

type metadataKey struct{}

// Metadata holds arbitrary key-value pairs attached to a context.
type Metadata struct {
	Values map[string]string
}

// WithMetadata returns a new context with the given Metadata stored in it.
func WithMetadata(ctx context.Context, m Metadata) context.Context {
	if m.Values == nil {
		m.Values = make(map[string]string)
	}
	return context.WithValue(ctx, metadataKey{}, m)
}

// MetadataFromContext retrieves Metadata from the context.
// Returns false if no Metadata is present.
func MetadataFromContext(ctx context.Context) (Metadata, bool) {
	m, ok := ctx.Value(metadataKey{}).(Metadata)
	return m, ok
}
