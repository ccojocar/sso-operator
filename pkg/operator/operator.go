package operator

import (
	"context"

	"github.com/jenkins-x/sso-operator/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/sso-operator/pkg/dex"
	"github.com/jenkins-x/sso-operator/pkg/proxy"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// NewHandler returns a new SSO operator event handler
func NewHandler(dexClient *dex.Client, exposeServiceAccount string) sdk.Handler {
	return &Handler{
		dexClient:            dexClient,
		exposeServiceAccount: exposeServiceAccount,
	}
}

// Handler is a SSO operator event handler
type Handler struct {
	dexClient            *dex.Client
	exposeServiceAccount string
}

// Handle handles SSO operator events
func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1.SSO:
		sso := o.DeepCopy()

		// Ignore the delete event
		if event.Deleted && sso.Status.Initialized {
			err := proxy.Cleanup(sso, sso.GetName(), h.exposeServiceAccount)
			if err != nil {
				return errors.Wrapf(err, "cleaning up '%s' SSO proxy", sso.GetName())
			}
			clientID := sso.Status.ClientID
			err = h.dexClient.DeleteClient(ctx, clientID)
			if err != nil {
				return errors.Wrapf(err, "deleting OIDC client '%s' from dex", clientID)
			}
			return nil
		}

		// SSO was initialized already
		if sso.Status.Initialized {
			return nil
		}

		redirectUris := []string{proxy.FakeRedirectURL()}
		publicClient := true
		client, err := h.dexClient.CreateClient(ctx, redirectUris, []string{}, publicClient, sso.Name, "")
		if err != nil {
			return errors.Wrapf(err, "creating the OIDC client '%s' in dex", sso.GetName())
		}

		p, err := proxy.Deploy(sso, client)
		if err != nil {
			return errors.Wrapf(err, "deploying '%s' SSO proxy", sso.GetName())
		}

		err = proxy.Expose(sso, p.Service.GetName(), h.exposeServiceAccount)
		if err != nil {
			return errors.Wrapf(err, "exposing '%s' SSO proxy", sso.GetName())
		}

		sso.Status.ClientID = client.Id
		sso.Status.Initialized = true
		err = sdk.Update(sso)
		if err != nil {
			return errors.Wrapf(err, "updating '%s' SSO CRD", sso.GetName())
		}

		logrus.Infof("SSO proxy '%s' initialized", sso.GetName())
	}
	return nil
}
