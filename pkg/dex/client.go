package dex

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"github.com/dexidp/dex/api"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Options keeps some configuration options for Dex client
type Options struct {
	// HostAndPort host name and port of gRPC server
	HostAndPort string
	// ClientCrt TLS certificate for gRPC client
	ClientCrt string
	// ClientKey TLS certificate key for gRPC client
	ClientKey string
	// ClientCA self signed CA certificate for gRPC TLS connection
	ClientCA string
}

// Client represent a client wrapper for Dex
type Client struct {
	dex api.DexClient
}

// NewClient creates a new Dex client
func NewClient(opts *Options) (*Client, error) {
	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(opts.ClientCA) // #nosec
	if err != nil {
		return nil, errors.Wrapf(err, "reading the public CA cert from %q", opts.ClientCA)
	}
	appended := certPool.AppendCertsFromPEM(caCert)
	if !appended {
		return nil, errors.New("failed to append the CA cert to the certs pool")
	}

	clientCert, err := tls.LoadX509KeyPair(opts.ClientCrt, opts.ClientKey)
	if err != nil {
		return nil, errors.Wrapf(err, "loading the client cert %q and private key %q",
			opts.ClientCrt, opts.ClientKey)
	}

	clientTLSConfig := &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientCert},
		MinVersion:   tls.VersionTLS12,
	}
	creds := credentials.NewTLS(clientTLSConfig)

	conn, err := grpc.Dial(opts.HostAndPort, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, errors.Wrapf(err, "opening the gRPC connection with server %q", opts.HostAndPort)
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
		return nil, errors.Wrapf(err, "client %q already exists", res.Client.Id)
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
		Name:         name,
		LogoUrl:      logoURL,
	}

	res, err := c.dex.UpdateClient(ctx, req)
	if err != nil {
		return errors.Wrapf(err, "failed to update the client with id %q", clientID)
	}

	if res.NotFound {
		return fmt.Errorf("update did not find the client with id %q", clientID)
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
		return errors.Wrapf(err, "failed to delete the client with id %q", id)
	}
	if res.NotFound {
		return fmt.Errorf("delete did not find the client with id %q", id)
	}
	return nil
}
