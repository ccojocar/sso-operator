package proxy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyConfig(t *testing.T) {
	config := &Config{
		Port:          4180,
		ClientID:      "123",
		ClientSecret:  "test",
		OIDCIssuerURL: "http://test-issuer",
		RedirectURL:   "http://test-proxy/calback",
		LoginURL:      "http://test-proxy/auth",
		RedeemURL:     "http://test-proxy/token",
		Upstream:      "http://test-upstream",
		ForwardToken:  false,
		Cookie: Cookie{
			Name:     "test-cookie",
			Secret:   "test",
			Domain:   "http://test-proxy",
			Expire:   "168h0m",
			Refresh:  "60m",
			Secure:   true,
			HTTPOnly: true,
		},
	}

	strConfig, err := renderConfig(config)

	assert.NoError(t, err, "should render proxy config without error")
	assert.NotEmpty(t, strConfig, "proxy config should not be empty")
}
