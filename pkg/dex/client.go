package dex

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/coreos/dex/api"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client represent a client wrapper for Dex
type Client struct {
	dex api.DexClient
}

// NewClient creates a new Dex client
func NewClient(hostAndPort, caPath, clientCrt, clientKey string) (*Client, error) {
	clientCert, err := tls.LoadX509KeyPair(clientCrt, clientKey)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load the client cert '%s' and key '%s'", clientCrt, clientKey)
	}

	// The CA public cert should be packaged with the client cert
	if len(clientCert.Certificate) != 2 {
		return nil, fmt.Errorf("failed to load the CA cert from '%s'", clientCrt)
	}

	caCert := clientCert.Certificate[1]
	certPool := x509.NewCertPool()
	appended := certPool.AppendCertsFromPEM(caCert)
	if !appended {
		return nil, errors.New("failed to append the CA cert to the certs pool")
	}

	clientTLSConfig := &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientCert},
	}
	creds := credentials.NewTLS(clientTLSConfig)

	conn, err := grpc.Dial(hostAndPort, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open the gRPC connection with server '%s'", hostAndPort)
	}
	return &Client{
		dex: api.NewDexClient(conn),
	}, nil
}

// CreateClient a new OIDC client in Dex
func (c *Client) CreateClient(ctx context.Context, redirectUris []string, trustedPeers []string,
	public bool, name string, logoURL string) (*api.Client, error) {
	req := &api.CreateClientReq{
		Client: &api.Client{
			RedirectUris: redirectUris,
			TrustedPeers: trustedPeers,
			Public:       public,
			Name:         name,
			LogoUrl:      logoURL,
		},
	}

	res, err := c.dex.CreateClient(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create the OIDC client")
	}

	if res.AlreadyExists {
		return nil, errors.Wrapf(err, "client '%s' already exists", res.Client.Id)
	}

	return res.Client, nil
}

// UpdateClient updates an already registered OIDC client
func (c *Client) UpdateClient(ctx context.Context, clientID string, redirectUris []string,
	trustedPeers []string, public bool, name string, logoURL string) error {
	req := &api.UpdateClientReq{
		Id:           clientID,
		RedirectUris: redirectUris,
		TrustedPeers: trustedPeers,
		Public:       public,
		Name:         name,
		LogoUrl:      logoURL,
	}

	res, err := c.dex.UpdateClient(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "failed to update the client with id '%s'", clientID)
	}

	if res.NotFound {
		return fmt.Errorf("update did not find the client with id '%s'", clientID)
	}
	return nil
}

// DeleteClient deletes the client with given Id from Dex
func (c *Client) DeleteClient(ctx context.Context, id string) error {
	req := &api.DeleteClientReq{
		Id: id,
	}
	res, err := c.dex.DeleteClient(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "failed to delete the client with id '%s'", id)
	}
	if res.NotFound {
		return fmt.Errorf("delete did not find the client with id '%s'", id)
	}
	return nil
}
