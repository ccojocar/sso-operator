package proxy

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
)

// Cookie holds the configuration for oauth2_proxy cookie
type Cookie struct {
	Name     string
	Secret   string
	Domain   string
	Expire   string
	Refresh  string
	Secure   bool
	HTTPOnly bool
}

// Config holds the configuration for oauth2_proxy
type Config struct {
	Port int32

	ClientID     string
	ClientSecret string

	OIDCIssuerURL string
	RedirectURL   string
	LoginURL      string
	RedeemURL     string

	Upstream string

	Cookie Cookie
}

const proxyConfigTemplate = `
# OAuth2 Proxy Config File

## <addr>:<port> to listen on for HTTP/HTTPS clients
http-address = ":{{.Port}}"

## the OAuth URLs.
redirect-url = "{{.RedirectURL}}"
login-url = "{{.LoginURL}}"
redeem-url = "{{.RedeemURL}}"

## the http url(s) of the upstream endpoint. If multiple, routing is based on path
upstreams = [
    "{{.Upstream}}"
]

## Log requests to stdout
request-logging = true

## The OAuth Client ID, Secret
client-id = "{{.ClientID}}"
client-secret = "{{.ClientSecret}}"

## Pass OAuth Access token to upstream via "X-Forwarded-Access-Token"
pass-basic-auth = false
pass-host-header = false

## Cookie Settings
## Name     - the cookie name
## Secret   - the seed string for secure cookies; should be 16, 24, or 32 bytes
##            for use with an AES cipher when cookie_refresh or pass_access_token
##            is set
## Domain   - (optional) cookie domain to force cookies to (ie: .yourcompany.com)
## Expire   - (duration) expire timeframe for cookie
## Refresh  - (duration) refresh the cookie when duration has elapsed after cookie was initially set.
##            Should be less than cookie_expire; set to 0 to disable.
##            On refresh, OAuth token is re-validated.
##            (ie: 1h means tokens are refreshed on request 1hr+ after it was set)
## Secure   - secure cookies are only sent by the browser of a HTTPS connection (recommended)
## HttpOnly - httponly cookies are not readable by javascript (recommended)
cookie-name = "{{.Cookie.Name}}"
cookie-secret = "{{.Cookie.Secret}}"
cookie-domain = "{{.Cookie.Domain}}"
cookie-expire = "{{.Cookie.Expire}}"
cookie-refresh = "{{.Cookie.Refresh}}"
cookie-secure = {{.Cookie.Secure}}
cookie-httponly = {{.Cookie.HTTPOnly}}

## Provider Specific Configurations
provider = "oidc"
oidc-issuer-url = "{{.OIDCIssuerURL}}"
scope = "openid email profile"
skip-provider-button = true
`

func renderConfig(config *Config) (string, error) {
	tmpl := template.New("oauth2_proxy.tpl")

	tmpl, err := tmpl.Parse(proxyConfigTemplate)
	if err != nil {
		return "", errors.Wrap(err, "parsing proxy config template")
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, *config)
	if err != nil {
		return "", errors.Wrap(err, "rendering proxy config")
	}

	return buf.String(), nil
}
