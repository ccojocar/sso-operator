package operator

import (
	"context"
	"fmt"

	"github.com/jenkins-x/sso-operator/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/sso-operator/pkg/dex"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// NewHandler returns a new SSO operator event handler
func NewHandler(dexClient *dex.Client) sdk.Handler {
	return &Handler{
		dexClient: dexClient,
	}
}

// Handler is a SSO operator event handler
type Handler struct {
	dexClient *dex.Client
}

// Handle handles SSO operator events
func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1.SSO:
		// Ignore the delete event since the garbage collector will clean up all secondary resources for the CR
		if event.Deleted {
			return nil
		}

		sso := o

		redirectUris := []string{fmt.Sprintf("http://%s/callback", sso.Name)}
		publicClient := true
		client, err := h.dexClient.CreateClient(ctx, redirectUris, []string{}, publicClient, sso.Name, "")
		if err != nil {
			return errors.Wrapf(err, "failed to create the OIDC client in dex for SSO CR with name '%s'", sso.Name)
		}

		logrus.Infof("Client created: %v", client)

	}
	return nil
}
