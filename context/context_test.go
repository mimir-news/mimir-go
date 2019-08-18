package context_test

import (
	"context"
	"testing"

	customContext "github.com/mimir-news/mimir-go/context"
)

func TestNewBackground(t *testing.T) {
	ctx24 := customContext.NewBackground("teast-client-id", "sv", "auth-token")
	var ctx context.Context = ctx24

	if ctx24.ID == "" {
		t.Errorf("context.ID was empty should be UUID string")
	}

	val := ctx.Value(customContext.ContextIDKey)
	ctxID, ok := val.(string)
	if !ok {
		t.Errorf("context.Value returned unexpected type. Expected: stirng Got: %v", val)
	}

	if ctx24.ID != ctxID {
		t.Errorf("context.Value returned unexpected value. Expected: %s Got: %s", ctx24.ID, ctxID)
	}
}
