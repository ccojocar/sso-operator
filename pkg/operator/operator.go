package operator

import (
	"context"
	"sync"

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
	mutex                sync.Mutex
	dexClient            *dex.Client
	exposeServiceAccount string
	contexts             []ssoContext
}

type ssoContext struct {
	Name     string
	Namespce string
}

func (h *Handler) start(sso *v1.SSO) bool {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for _, ctx := range h.contexts {
		if ctx.Name == sso.GetName() && ctx.Namespce == sso.GetNamespace() {
			return false
		}
	}
	h.contexts = append(h.contexts, ssoContext{
		Name:     sso.GetName(),
		Namespce: sso.GetNamespace(),
	})

	return true
}

func (h *Handler) end(sso *v1.SSO) {
	h.mutex.Lock()
	h.mutex.Unlock()

	for i, ctx := range h.contexts {
		if ctx.Name == sso.GetName() && ctx.Namespce == sso.GetNamespace() {
			h.contexts = append(h.contexts[:i], h.contexts[i+1:]...)
		}
	}
}

// Handle handles SSO operator events
func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1.SSO:
		// Ignore the delete event
		if event.Deleted {
			return nil
		}

		sso := o.DeepCopy()

		// Skip this SSO event, already a create is ongoing
		started := h.start(sso)
		if !started {
			return nil
		}
		defer h.end(sso)

		// SSO was initialized already
		if sso.Status.Initialized {
			return nil
		}

		redirectUris := []string{proxy.FakeRedirectURL()}
		publicClient := true
		client, err := h.dexClient.CreateClient(ctx, redirectUris, []string{}, publicClient, sso.Name, "")
		if err != nil {
			return errors.Wrapf(err, "creating the OIDC client '%s' in dex", sso.Name)
		}

		p, err := proxy.Deploy(sso, client)
		if err != nil {
			return errors.Wrap(err, "deploying SSO proxy")
		}

		err = proxy.Expose(sso, p.Service.GetName(), h.exposeServiceAccount)
		if err != nil {
			return errors.Wrap(err, "exposing SSO proxy")
		}

		sso.Status.Initialized = true
		err = sdk.Update(sso)
		if err != nil {
			return errors.Wrap(err, "updating SSO CRD")
		}

		logrus.Infof("SSO proxy '%s' initialized", sso.GetName())
	}
	return nil
}
