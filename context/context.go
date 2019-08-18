package context

import (
	"context"
	"fmt"

	"github.com/mimir-news/mimir-go/id"
)

type contextKey struct{}

// ContextIDKey id key for context.
var ContextIDKey = contextKey{}

// Context request context.
type Context struct {
	ID        string
	ClientID  string
	Language  string
	AuthToken string
	context.Context
}

// New creates new context from parent.
func New(parent context.Context, id, clientID, language, authToken string) *Context {
	return &Context{
		ID:        id,
		ClientID:  clientID,
		Language:  language,
		AuthToken: authToken,
		Context:   context.WithValue(parent, ContextIDKey, id),
	}
}

// NewBackground creates a new context based on a background context.
func NewBackground(clientID, language, authToken string) *Context {
	parent := context.Background()
	return New(parent, id.New(), clientID, language, authToken)
}

func (c *Context) String() string {
	return fmt.Sprintf("Context(id=[%s], clientId=[%s] language=%s)", c.ID, c.ClientID, c.Language)
}
