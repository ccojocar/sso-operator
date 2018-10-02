package operator

import (
	"context"

	"github.com/jenkins-x/sso-operator/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/sso-operator/pkg/dex"
	"github.com/jenkins-x/sso-operator/pkg/kubernetes"
	"github.com/jenkins-x/sso-operator/pkg/proxy"
	"github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// NewHandler returns a new SSO operator event handler
func NewHandler(dexClient *dex.Client, clusterRoleName string) (sdk.Handler, error) {
	ssoCookieKey, err := proxy.GenerateCookieKey()
	if err != nil {
		return nil, err
	}
	return &Handler{
		dexClient:       dexClient,
		clusterRoleName: clusterRoleName,
		ssoCookieKey:    ssoCookieKey,
	}, nil
}

// Handler is a SSO operator event handler
type Handler struct {
	dexClient       *dex.Client
	clusterRoleName string
	ssoCookieKey    string
}

// Handle handles SSO operator events
func (h *Handler) Handle(ctx context.Context, event sdk.Event) error {
	switch o := event.Object.(type) {
	case *v1.SSO:
		sso := o.DeepCopy()

		// Ensure that the operator has permissions in the namespace of the single sing-on resource
		saName, err := kubernetes.EnsureClusterRoleBinding(h.clusterRoleName, sso.GetNamespace())
		if err != nil {
			return errors.Wrapf(err, "ensuring cluster role '%s' has a binding to a service account in the namespace '%s'",
				h.clusterRoleName, sso.GetNamespace())
		}

		// Cleanup all resources when a SSO CR is deleted
		if event.Deleted && sso.Status.Initialized {
			err := proxy.Cleanup(sso, sso.GetName(), saName)
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

		// Check if SSO was already initialized
		initialized, err := kubernetes.IsSSOInitialized(sso)
		if err != nil {
			logrus.Infof("Failed to get the status of SSO '%s'. Error: %v", sso.GetName(), err)
			initialized = false
		}
		if initialized {
			return nil
		}
		logrus.Infof("Initializing SSO '%s'", sso.GetName())

		// Crate a new OIDC client in dex
		redirectURLs := []string{proxy.FakeRedirectURL()}
		publicClient := false
		client, err := h.dexClient.CreateClient(ctx, redirectURLs, []string{}, publicClient, sso.Name, "")
		if err != nil {
			return errors.Wrapf(err, "creating the OIDC client '%s' in dex", sso.GetName())
		}

		// Deploy the OIDC proxy
		proxyResources, err := proxy.Deploy(sso, client, h.ssoCookieKey)
		if err != nil {
			return errors.Wrapf(err, "deploying '%s' SSO proxy", sso.GetName())
		}

		// Expose the OIDC proxy service publicly
		err = proxy.Expose(sso, proxyResources.Service.GetName(), saName)
		if err != nil {
			return errors.Wrapf(err, "exposing '%s' SSO proxy", sso.GetName())
		}

		// Update in dex the redirect URL of the OIDC client
		ingressHosts, err := kubernetes.FindIngressHosts(proxyResources.AppName, sso.GetNamespace())
		if err != nil {
			return errors.Wrap(err, "searching ingress hosts")
		}
		redirectURLs = proxy.ConvertHostsToRedirectURLs(ingressHosts, sso)
		err = h.dexClient.UpdateClient(ctx, client.Id, redirectURLs, []string{}, publicClient, sso.Name, "")
		if err != nil {
			return errors.Wrapf(err, "updating the OIDC client '%s' in dex", client.Id)
		}
		client.RedirectUris = redirectURLs

		// Update the OIDC proxy
		err = proxy.Update(proxyResources, sso, client, h.ssoCookieKey)
		if err != nil {
			return errors.Wrapf(err, "updating '%s' SSO proxy", sso.GetName())
		}

		// Update the status of SSO CR
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
