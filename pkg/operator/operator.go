package operator

import (
	"context"

	"github.com/jenkins-x/sso-operator/pkg/dex"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
)

// NewHandler returns a new SSO operator event handler
func NewHandler(dexClient *dex.Client) sdk.Handler {
	return &Handler{
		dex: dexClient,
	}
}

// Handler is a SSO operator event handler
type Handler struct {
	dex *dex.Client
}

// Handle handles SSO operator events
func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	return nil
}
