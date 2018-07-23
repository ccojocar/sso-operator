package operator

import (
	"context"

	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

// NewHandler returns a new SSO operator event handler
func NewHandler() sdk.Handler {
	return &Handler{}
}

// Handler is a SSO operator event handler
type Handler struct {
}

// Handle handles SSO operator events
func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	return nil
}
